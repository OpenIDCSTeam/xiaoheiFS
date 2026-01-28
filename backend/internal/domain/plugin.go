package domain

import "time"

type PluginSignatureStatus string

const (
	PluginSignatureOfficial  PluginSignatureStatus = "official"
	PluginSignatureUntrusted PluginSignatureStatus = "untrusted"
	PluginSignatureUnsigned  PluginSignatureStatus = "unsigned"
)

type PluginInstallation struct {
	ID              int64
	Category        string
	PluginID        string
	InstanceID      string
	Enabled         bool
	SignatureStatus PluginSignatureStatus
	ConfigCipher    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
