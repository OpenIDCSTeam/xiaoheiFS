package plugins

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xiaoheiplay/internal/domain"
	"xiaoheiplay/internal/usecase"
)

const pluginsBootstrappedSettingKey = "plugins_bootstrapped"

type DiscoverItem struct {
	Category        string                       `json:"category"`
	PluginID        string                       `json:"plugin_id"`
	Name            string                       `json:"name"`
	Version         string                       `json:"version"`
	SignatureStatus domain.PluginSignatureStatus `json:"signature_status"`
	Entry           EntryInfo                    `json:"entry"`
}

func (m *Manager) BootstrapFromDisk(ctx context.Context, settings usecase.SettingsRepository) error {
	if m.repo == nil {
		return errors.New("plugin repo missing")
	}

	existing, err := m.repo.ListPluginInstallations(ctx)
	if err != nil {
		return err
	}
	existingMap := map[string]domain.PluginInstallation{}
	for _, inst := range existing {
		existingMap[inst.Category+":"+inst.PluginID] = inst
	}

	bootstrapped := false
	if settings != nil {
		if s, err := settings.GetSetting(ctx, pluginsBootstrappedSettingKey); err == nil {
			if v, ok := parseBoolSetting(s.ValueJSON); ok {
				bootstrapped = v
			}
		}
	}
	if len(existing) == 0 {
		bootstrapped = false
	}

	found, err := scanDiskPlugins(m.baseDir)
	if err != nil {
		return err
	}

	// First bootstrap: import everything (enabled=false, config empty).
	if !bootstrapped {
		for _, p := range found {
			key := p.Category + ":" + p.PluginID
			if _, ok := existingMap[key]; ok {
				continue
			}
			sigStatus, _ := VerifySignature(p.Dir, m.officialKeys)
			_, _ = ResolveEntry(p.Dir, p.Manifest) // parse existence (for operator visibility via list/discover)
			inst := domain.PluginInstallation{
				Category:        p.Category,
				PluginID:        p.PluginID,
				InstanceID:      newInstanceID(p.Category, p.PluginID),
				Enabled:         false,
				SignatureStatus: sigStatus,
				ConfigCipher:    "",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			_ = m.repo.UpsertPluginInstallation(ctx, &inst)
		}
		if settings != nil {
			_ = settings.UpsertSetting(ctx, domain.Setting{Key: pluginsBootstrappedSettingKey, ValueJSON: "true", UpdatedAt: time.Now()})
		}
		return nil
	}

	// Subsequent startup: auto-import ONLY new official plugins found on disk.
	for _, p := range found {
		key := p.Category + ":" + p.PluginID
		if _, ok := existingMap[key]; ok {
			continue
		}
		sigStatus, _ := VerifySignature(p.Dir, m.officialKeys)
		if sigStatus != domain.PluginSignatureOfficial {
			continue
		}
		inst := domain.PluginInstallation{
			Category:        p.Category,
			PluginID:        p.PluginID,
			InstanceID:      newInstanceID(p.Category, p.PluginID),
			Enabled:         false,
			SignatureStatus: sigStatus,
			ConfigCipher:    "",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		_ = m.repo.UpsertPluginInstallation(ctx, &inst)
	}
	return nil
}

func (m *Manager) DiscoverOnDisk(ctx context.Context) ([]DiscoverItem, error) {
	if m.repo == nil {
		return nil, errors.New("plugin repo missing")
	}
	existing, err := m.repo.ListPluginInstallations(ctx)
	if err != nil {
		return nil, err
	}
	existingMap := map[string]struct{}{}
	for _, inst := range existing {
		existingMap[inst.Category+":"+inst.PluginID] = struct{}{}
	}

	found, err := scanDiskPlugins(m.baseDir)
	if err != nil {
		return nil, err
	}

	out := make([]DiscoverItem, 0, len(found))
	for _, p := range found {
		if _, ok := existingMap[p.Category+":"+p.PluginID]; ok {
			continue
		}
		sigStatus, _ := VerifySignature(p.Dir, m.officialKeys)
		entry, err := ResolveEntry(p.Dir, p.Manifest)
		if err != nil {
			// keep entry info (supported platforms) for UI
		}
		out = append(out, DiscoverItem{
			Category:        p.Category,
			PluginID:        p.PluginID,
			Name:            p.Manifest.Name,
			Version:         p.Manifest.Version,
			SignatureStatus: sigStatus,
			Entry:           entry,
		})
	}
	return out, nil
}

func (m *Manager) ImportFromDisk(ctx context.Context, category, pluginID string) (domain.PluginInstallation, error) {
	if m.repo == nil {
		return domain.PluginInstallation{}, errors.New("plugin repo missing")
	}
	category = strings.TrimSpace(category)
	pluginID = strings.TrimSpace(pluginID)
	if category == "" || pluginID == "" {
		return domain.PluginInstallation{}, errors.New("invalid plugin")
	}
	pluginDir := filepath.Join(m.baseDir, category, pluginID)
	if _, err := os.Stat(pluginDir); err != nil {
		return domain.PluginInstallation{}, errors.New("plugin dir not found")
	}

	manifest, err := ReadManifest(pluginDir)
	if err != nil {
		return domain.PluginInstallation{}, err
	}
	if manifest.PluginID != pluginID {
		return domain.PluginInstallation{}, errors.New("manifest plugin_id mismatch")
	}

	entry, err := ResolveEntry(pluginDir, manifest)
	if err != nil {
		if len(entry.SupportedPlatforms) > 0 {
			return domain.PluginInstallation{}, errors.New("unsupported platform " + entry.Platform + ", supported: " + strings.Join(entry.SupportedPlatforms, ", "))
		}
		return domain.PluginInstallation{}, err
	}

	sigStatus, err := VerifySignature(pluginDir, m.officialKeys)
	if err != nil {
		return domain.PluginInstallation{}, err
	}

	// Upsert without touching enabled/config (new import defaults to disabled/empty).
	if inst, err := m.repo.GetPluginInstallation(ctx, category, pluginID); err == nil {
		inst.SignatureStatus = sigStatus
		_ = m.repo.UpsertPluginInstallation(ctx, &inst)
		return m.repo.GetPluginInstallation(ctx, category, pluginID)
	}

	inst := domain.PluginInstallation{
		Category:        category,
		PluginID:        pluginID,
		InstanceID:      newInstanceID(category, pluginID),
		Enabled:         false,
		SignatureStatus: sigStatus,
		ConfigCipher:    "",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := m.repo.UpsertPluginInstallation(ctx, &inst); err != nil {
		return domain.PluginInstallation{}, err
	}
	_ = entry // resolved above to validate availability
	return m.repo.GetPluginInstallation(ctx, category, pluginID)
}

func (m *Manager) SignatureStatusOnDisk(category, pluginID string) (domain.PluginSignatureStatus, error) {
	category = strings.TrimSpace(category)
	pluginID = strings.TrimSpace(pluginID)
	if category == "" || pluginID == "" {
		return domain.PluginSignatureUntrusted, errors.New("invalid plugin")
	}
	pluginDir := filepath.Join(m.baseDir, category, pluginID)
	if _, err := os.Stat(filepath.Join(pluginDir, "manifest.json")); err != nil {
		return domain.PluginSignatureUntrusted, errors.New("manifest.json not found")
	}
	return VerifySignature(pluginDir, m.officialKeys)
}

type diskPlugin struct {
	Category string
	PluginID string
	Dir      string
	Manifest Manifest
}

func scanDiskPlugins(baseDir string) ([]diskPlugin, error) {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		return nil, errors.New("missing base dir")
	}
	st, err := os.Stat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !st.IsDir() {
		return nil, errors.New("plugins base dir is not a directory")
	}

	var out []diskPlugin
	cats, _ := os.ReadDir(baseDir)
	for _, c := range cats {
		if !c.IsDir() {
			continue
		}
		category := strings.TrimSpace(c.Name())
		if category == "" || strings.HasPrefix(category, ".") {
			continue
		}
		catDir := filepath.Join(baseDir, category)
		plugins, _ := os.ReadDir(catDir)
		for _, p := range plugins {
			if !p.IsDir() {
				continue
			}
			pluginID := strings.TrimSpace(p.Name())
			if pluginID == "" || strings.HasPrefix(pluginID, ".") {
				continue
			}
			dir := filepath.Join(catDir, pluginID)
			if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
				continue
			}
			m, err := ReadManifest(dir)
			if err != nil {
				continue
			}
			if m.PluginID != pluginID {
				continue
			}
			out = append(out, diskPlugin{Category: category, PluginID: pluginID, Dir: dir, Manifest: m})
		}
	}
	return out, nil
}

func parseBoolSetting(v string) (bool, bool) {
	s := strings.TrimSpace(v)
	s = strings.Trim(s, "\"")
	switch strings.ToLower(s) {
	case "true", "1", "yes", "y":
		return true, true
	case "false", "0", "no", "n":
		return false, true
	default:
		return false, false
	}
}
