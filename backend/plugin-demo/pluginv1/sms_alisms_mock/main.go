package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"xiaoheiplay/pkg/pluginsdk"
	
)

type config struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SignName        string `json:"sign_name"`
}

type coreServer struct {
	pluginv1.UnimplementedCoreServiceServer
	cfg      config
	instance string
}

func (s *coreServer) GetManifest(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.Manifest, error) {
	_ = ctx
	return &pluginv1.Manifest{
		PluginId:    "alisms",
		Name:        "Aliyun SMS (Mock)",
		Version:     "0.1.0",
		Description: "Mock SMS plugin. Replace with real Aliyun SMS SDK calls.",
		Sms:         &pluginv1.SmsCapability{Send: true},
	}, nil
}

func (s *coreServer) GetConfigSchema(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.ConfigSchema, error) {
	_ = ctx
	return &pluginv1.ConfigSchema{
		JsonSchema: `{
  "title": "Aliyun SMS Config",
  "type": "object",
  "properties": {
    "access_key_id": { "type": "string", "title": "AccessKey ID" },
    "access_key_secret": { "type": "string", "title": "AccessKey Secret", "format": "password" },
    "sign_name": { "type": "string", "title": "Sign Name" }
  },
  "required": ["access_key_id","access_key_secret","sign_name"]
}`,
		UiSchema: `{ "access_key_secret": { "ui:widget": "password" } }`,
	}, nil
}

func (s *coreServer) ValidateConfig(ctx context.Context, req *pluginv1.ValidateConfigRequest) (*pluginv1.ValidateConfigResponse, error) {
	_ = ctx
	var cfg config
	if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "invalid json"}, nil
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" || strings.TrimSpace(cfg.AccessKeySecret) == "" || strings.TrimSpace(cfg.SignName) == "" {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "access_key_id/access_key_secret/sign_name required"}, nil
	}
	return &pluginv1.ValidateConfigResponse{Ok: true}, nil
}

func (s *coreServer) Init(ctx context.Context, req *pluginv1.InitRequest) (*pluginv1.InitResponse, error) {
	_ = ctx
	var cfg config
	if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
		return &pluginv1.InitResponse{Ok: false, Error: "invalid config"}, nil
	}
	s.cfg = cfg
	s.instance = req.GetInstanceId()
	return &pluginv1.InitResponse{Ok: true}, nil
}

func (s *coreServer) ReloadConfig(ctx context.Context, req *pluginv1.ReloadConfigRequest) (*pluginv1.ReloadConfigResponse, error) {
	_, err := s.Init(ctx, &pluginv1.InitRequest{InstanceId: s.instance, ConfigJson: req.GetConfigJson()})
	if err != nil {
		return &pluginv1.ReloadConfigResponse{Ok: false, Error: err.Error()}, nil
	}
	return &pluginv1.ReloadConfigResponse{Ok: true}, nil
}

func (s *coreServer) Health(ctx context.Context, _ *pluginv1.HealthCheckRequest) (*pluginv1.HealthCheckResponse, error) {
	_ = ctx
	return &pluginv1.HealthCheckResponse{
		Status:     pluginv1.HealthStatus_HEALTH_STATUS_OK,
		Message:    "ok",
		UnixMillis: time.Now().UnixMilli(),
	}, nil
}

type smsServer struct {
	pluginv1.UnimplementedSmsServiceServer
}

func (s *smsServer) Send(ctx context.Context, req *pluginv1.SendSmsRequest) (*pluginv1.SendSmsResponse, error) {
	_ = ctx
	_ = req
	// TODO: Integrate Aliyun SMS SDK.
	return &pluginv1.SendSmsResponse{Ok: true, MessageId: "mock-" + time.Now().Format("20060102150405")}, nil
}

func main() {
	core := &coreServer{}
	sms := &smsServer{}
	pluginsdk.Serve(map[string]pluginsdk.Plugin{
		pluginsdk.PluginKeyCore: &pluginsdk.CoreGRPCPlugin{Impl: core},
		pluginsdk.PluginKeySMS:  &pluginsdk.SmsGRPCPlugin{Impl: sms},
	})
}
