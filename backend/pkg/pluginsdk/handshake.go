package pluginsdk

import "github.com/hashicorp/go-plugin"

const (
	ProtocolVersion  = 1
	MagicCookieKey   = "XIAOHEI_PLUGIN"
	MagicCookieValue = "xiaoheiplay"
)

const (
	PluginKeyCore    = "core"
	PluginKeySMS     = "sms"
	PluginKeyPayment = "payment"
	PluginKeyKYC     = "kyc"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieValue,
}
