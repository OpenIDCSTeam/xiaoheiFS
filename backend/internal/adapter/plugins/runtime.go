package plugins

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"

	"xiaoheiplay/pkg/pluginsdk"
	pluginv1 "xiaoheiplay/plugin/v1"
)

type Runtime struct {
	baseDir string

	mu      sync.Mutex
	running map[string]*runningPlugin
}

type runningPlugin struct {
	mu sync.Mutex

	category   string
	pluginID   string
	instanceID string

	client   *plugin.Client
	core     pluginv1.CoreServiceClient
	sms      pluginv1.SmsServiceClient
	payment  pluginv1.PaymentServiceClient
	kyc      pluginv1.KycServiceClient
	manifest *pluginv1.Manifest

	lastHealth time.Time
	health     *pluginv1.HealthCheckResponse
	cancelHB   context.CancelFunc
}

func NewRuntime(baseDir string) *Runtime {
	return &Runtime{baseDir: baseDir, running: map[string]*runningPlugin{}}
}

func (r *Runtime) key(category, pluginID string) string {
	return category + ":" + pluginID
}

func (r *Runtime) Start(ctx context.Context, category, pluginID, instanceID, configJSON string) (*pluginv1.Manifest, error) {
	if category == "" || pluginID == "" || instanceID == "" {
		return nil, errors.New("invalid input")
	}
	k := r.key(category, pluginID)

	r.mu.Lock()
	if existing := r.running[k]; existing != nil {
		r.mu.Unlock()
		return existing.manifest, nil
	}
	r.mu.Unlock()

	pluginDir := filepath.Join(r.baseDir, category, pluginID)
	manifestJSON, err := ReadManifest(pluginDir)
	if err != nil {
		return nil, err
	}
	entry, err := ResolveEntry(pluginDir, manifestJSON)
	if err != nil {
		if len(entry.SupportedPlatforms) > 0 {
			return nil, errors.New("unsupported platform " + entry.Platform + ", supported: " + strings.Join(entry.SupportedPlatforms, ", "))
		}
		return nil, err
	}

	cmd := exec.Command(entry.EntryPath)
	cmd.Dir = pluginDir

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginsdk.Handshake,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
		Plugins: map[string]plugin.Plugin{
			pluginsdk.PluginKeyCore:    &pluginsdk.CoreGRPCPlugin{},
			pluginsdk.PluginKeySMS:     &pluginsdk.SmsGRPCPlugin{},
			pluginsdk.PluginKeyPayment: &pluginsdk.PaymentGRPCPlugin{},
			pluginsdk.PluginKeyKYC:     &pluginsdk.KycGRPCPlugin{},
		},
		Cmd: cmd,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, err
	}
	rawCore, err := rpcClient.Dispense(pluginsdk.PluginKeyCore)
	if err != nil {
		client.Kill()
		return nil, err
	}
	core, ok := rawCore.(pluginv1.CoreServiceClient)
	if !ok {
		client.Kill()
		return nil, errors.New("invalid core client")
	}

	ctxm, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	manifest, err := core.GetManifest(ctxm, &pluginv1.Empty{})
	if err != nil {
		client.Kill()
		return nil, err
	}
	if manifest.GetPluginId() == "" {
		client.Kill()
		return nil, errors.New("invalid manifest")
	}

	var sms pluginv1.SmsServiceClient
	var payment pluginv1.PaymentServiceClient
	var kyc pluginv1.KycServiceClient

	if manifest.Sms != nil {
		raw, err := rpcClient.Dispense(pluginsdk.PluginKeySMS)
		if err != nil {
			client.Kill()
			return nil, err
		}
		c, ok := raw.(pluginv1.SmsServiceClient)
		if !ok {
			client.Kill()
			return nil, errors.New("invalid sms client")
		}
		sms = c
	}
	if manifest.Payment != nil {
		raw, err := rpcClient.Dispense(pluginsdk.PluginKeyPayment)
		if err != nil {
			client.Kill()
			return nil, err
		}
		c, ok := raw.(pluginv1.PaymentServiceClient)
		if !ok {
			client.Kill()
			return nil, errors.New("invalid payment client")
		}
		payment = c
	}
	if manifest.Kyc != nil {
		raw, err := rpcClient.Dispense(pluginsdk.PluginKeyKYC)
		if err != nil {
			client.Kill()
			return nil, err
		}
		c, ok := raw.(pluginv1.KycServiceClient)
		if !ok {
			client.Kill()
			return nil, errors.New("invalid kyc client")
		}
		kyc = c
	}

	ctxi, cancelInit := context.WithTimeout(ctx, 10*time.Second)
	defer cancelInit()
	initResp, err := core.Init(ctxi, &pluginv1.InitRequest{InstanceId: instanceID, ConfigJson: configJSON})
	if err != nil {
		client.Kill()
		return nil, err
	}
	if initResp != nil && !initResp.Ok {
		client.Kill()
		if initResp.Error != "" {
			return nil, errors.New(initResp.Error)
		}
		return nil, errors.New("plugin init failed")
	}

	hbCtx, hbCancel := context.WithCancel(context.Background())
	rp := &runningPlugin{
		category:   category,
		pluginID:   pluginID,
		instanceID: instanceID,
		client:     client,
		core:       core,
		sms:        sms,
		payment:    payment,
		kyc:        kyc,
		manifest:   manifest,
		cancelHB:   hbCancel,
		health:     nil,
		lastHealth: time.Time{},
	}
	go rp.heartbeatLoop(hbCtx)

	r.mu.Lock()
	r.running[k] = rp
	r.mu.Unlock()

	return manifest, nil
}

func (r *Runtime) Stop(category, pluginID string) {
	k := r.key(category, pluginID)
	r.mu.Lock()
	rp := r.running[k]
	delete(r.running, k)
	r.mu.Unlock()
	if rp == nil {
		return
	}
	if rp.cancelHB != nil {
		rp.cancelHB()
	}
	if rp.client != nil {
		rp.client.Kill()
	}
}

func (r *Runtime) GetRunning(category, pluginID string) (*runningPlugin, bool) {
	k := r.key(category, pluginID)
	r.mu.Lock()
	defer r.mu.Unlock()
	rp := r.running[k]
	return rp, rp != nil
}

func (p *runningPlugin) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if p.core == nil || p.manifest == nil {
				continue
			}
			cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			resp, err := p.core.Health(cctx, &pluginv1.HealthCheckRequest{InstanceId: p.instanceID})
			cancel()
			if err != nil {
				continue
			}
			p.mu.Lock()
			p.lastHealth = time.Now()
			p.health = resp
			p.mu.Unlock()
		}
	}
}
