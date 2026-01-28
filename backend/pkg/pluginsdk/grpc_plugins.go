package pluginsdk

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pluginv1 "xiaoheiplay/plugin/v1"
)

type CoreGRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl pluginv1.CoreServiceServer
}

func (p *CoreGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	pluginv1.RegisterCoreServiceServer(s, p.Impl)
	return nil
}

func (p *CoreGRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pluginv1.NewCoreServiceClient(c), nil
}

type SmsGRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl pluginv1.SmsServiceServer
}

func (p *SmsGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	pluginv1.RegisterSmsServiceServer(s, p.Impl)
	return nil
}

func (p *SmsGRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pluginv1.NewSmsServiceClient(c), nil
}

type PaymentGRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl pluginv1.PaymentServiceServer
}

func (p *PaymentGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	pluginv1.RegisterPaymentServiceServer(s, p.Impl)
	return nil
}

func (p *PaymentGRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pluginv1.NewPaymentServiceClient(c), nil
}

type KycGRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl pluginv1.KycServiceServer
}

func (p *KycGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	pluginv1.RegisterKycServiceServer(s, p.Impl)
	return nil
}

func (p *KycGRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pluginv1.NewKycServiceClient(c), nil
}
