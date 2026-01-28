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
	Region    string `json:"region"`
}

type coreServer struct {
	pluginv1.UnimplementedCoreServiceServer
	cfg      config
	instance string
}

func (s *coreServer) GetManifest(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.Manifest, error) {
	_ = ctx
	return &pluginv1.Manifest{
		PluginId:    "tencent_kyc",
		Name:        "Tencent Cloud FaceID/eKYC (Mock)",
		Version:     "0.1.0",
		Description: "Mock KYC plugin. Replace with real Tencent Cloud FaceID calls.",
		Kyc:         &pluginv1.KycCapability{Start: true, QueryResult: true},
	}, nil
}

func (s *coreServer) GetConfigSchema(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.ConfigSchema, error) {
	_ = ctx
	return &pluginv1.ConfigSchema{
		JsonSchema: `{
  "title": "Tencent eKYC Config",
  "type": "object",
  "properties": {
    "secret_id": { "type": "string", "title": "SecretId" },
    "secret_key": { "type": "string", "title": "SecretKey", "format": "password" },
    "region": { "type": "string", "title": "Region", "default": "ap-guangzhou" }
  },
  "required": ["secret_id","secret_key"]
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
	if strings.TrimSpace(cfg.SecretID) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "secret_id/secret_key required"}, nil
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

type kycServer struct {
	pluginv1.UnimplementedKycServiceServer
}

func (k *kycServer) Start(ctx context.Context, req *pluginv1.KycStartRequest) (*pluginv1.KycStartResponse, error) {
	_ = ctx
	return &pluginv1.KycStartResponse{
		Ok:       true,
		Token:    "mock-token-" + strings.TrimSpace(req.GetUserId()),
		Url:      "https://example.com/faceid/start",
		NextStep: "redirect",
	}, nil
}

func (k *kycServer) QueryResult(ctx context.Context, req *pluginv1.KycQueryRequest) (*pluginv1.KycQueryResponse, error) {
	_ = ctx
	return &pluginv1.KycQueryResponse{
		Ok:      true,
		Status:  "PENDING",
		Reason:  "",
		RawJson: `{"mock":true}`,
	}, nil
}

func main() {
	core := &coreServer{}
	kyc := &kycServer{}
	pluginsdk.Serve(map[string]pluginsdk.Plugin{
		pluginsdk.PluginKeyCore: &pluginsdk.CoreGRPCPlugin{Impl: core},
		pluginsdk.PluginKeyKYC:  &pluginsdk.KycGRPCPlugin{Impl: kyc},
	})
}
