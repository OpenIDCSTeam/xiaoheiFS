package pluginsdk

import "github.com/hashicorp/go-plugin"

type Plugin = plugin.Plugin

func Serve(plugins map[string]Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         plugins,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
