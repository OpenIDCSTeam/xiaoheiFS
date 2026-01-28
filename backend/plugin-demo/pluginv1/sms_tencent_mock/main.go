package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"xiaoheiplay/pkg/pluginsdk"
	
)

type config struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	SdkAppID  string `json:"sdk_app_id"`
	SignName  string `json:"sign_name"`
}

type coreServer struct {
	pluginv1.UnimplementedCoreServiceServer
	cfg      config
	instance string
}

func (s *coreServer) GetManifest(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.Manifest, error) {
	_ = ctx
	return &pluginv1.Manifest{
		PluginId:    "tencent_sms",
		Name:        "Tencent Cloud SMS (Mock)",
		Version:     "0.1.0",
		Description: "Mock SMS plugin. Replace with real Tencent Cloud SMS SDK calls.",
		Sms:         &pluginv1.SmsCapability{Send: true},
	}, nil
}

func (s *coreServer) GetConfigSchema(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.ConfigSchema, error) {
	_ = ctx
	return &pluginv1.ConfigSchema{
		JsonSchema: `{
  "title": "Tencent SMS Config",
  "type": "object",
  "properties": {
    "secret_id": { "type": "string", "title": "SecretId" },
    "secret_key": { "type": "string", "title": "SecretKey", "format": "password" },
    "sdk_app_id": { "type": "string", "title": "SdkAppId" },
    "sign_name": { "type": "string", "title": "SignName" }
  },
  "required": ["secret_id","secret_key","sdk_app_id","sign_name"]
}`,
		UiSchema: `{ "secret_key": { "ui:widget": "password" } }`,
	}, nil
}

func (s *coreServer) ValidateConfig(ctx context.Context, req *pluginv1.ValidateConfigRequest) (*pluginv1.ValidateConfigResponse, error) {
	_ = ctx
	var cfg config
	if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "invalid json"}, nil
	}
	if strings.TrimSpace(cfg.SecretID) == "" || strings.TrimSpace(cfg.SecretKey) == "" || strings.TrimSpace(cfg.SdkAppID) == "" || strings.TrimSpace(cfg.SignName) == "" {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "secret_id/secret_key/sdk_app_id/sign_name required"}, nil
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
	// TODO: Integrate Tencent Cloud SMS SDK.
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
