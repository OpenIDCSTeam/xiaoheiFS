package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20190711"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xiaoheiplay/pkg/pluginsdk"
	pluginv1 "xiaoheiplay/plugin/v1"
)

type config struct {
	SecretID   string `json:"secret_id"`
	SecretKey  string `json:"secret_key"`
	SdkAppID   string `json:"sdk_app_id"`
	SignName   string `json:"sign_name"`
	Region     string `json:"region"`
	TimeoutSec int    `json:"timeout_sec"`
}

type coreServer struct {
	pluginv1.UnimplementedCoreServiceServer
	cfg       config
	instance  string
	smsClient *sms.Client
}

func (s *coreServer) GetManifest(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.Manifest, error) {
	_ = ctx
	return &pluginv1.Manifest{
		PluginId:    "tencent_sms",
		Name:        "Tencent Cloud SMS",
		Version:     "1.0.0",
		Description: "Tencent Cloud SMS via official tencentcloud-sdk-go.",
		Sms:         &pluginv1.SmsCapability{Send: true},
	}, nil
}

func (s *coreServer) GetConfigSchema(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.ConfigSchema, error) {
	_ = ctx
	return &pluginv1.ConfigSchema{
		JsonSchema: `{
  "title": "Tencent Cloud SMS",
  "type": "object",
  "properties": {
    "secret_id": { "type": "string", "title": "SecretId" },
    "secret_key": { "type": "string", "title": "SecretKey", "format": "password" },
    "sdk_app_id": { "type": "string", "title": "SdkAppId" },
    "sign_name": { "type": "string", "title": "SignName" },
    "region": { "type": "string", "title": "Region", "default": "ap-guangzhou" },
    "timeout_sec": { "type": "integer", "title": "Request Timeout (sec)", "default": 10, "minimum": 1, "maximum": 60 }
  },
  "required": ["secret_id","secret_key","sdk_app_id","sign_name"]
}`,
		UiSchema: `{
  "secret_key": { "ui:widget": "password", "ui:help": "留空表示不修改（由宿主处理）" }
}`,
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
	var cfg config
	if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
		return &pluginv1.InitResponse{Ok: false, Error: "invalid config"}, nil
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 10
	}
	if strings.TrimSpace(cfg.Region) == "" {
		cfg.Region = "ap-guangzhou"
	}
	cred := common.NewCredential(strings.TrimSpace(cfg.SecretID), strings.TrimSpace(cfg.SecretKey))
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqTimeout = cfg.TimeoutSec
	client, err := sms.NewClient(cred, cfg.Region, cpf)
	if err != nil {
		return &pluginv1.InitResponse{Ok: false, Error: "init tencent sms client failed: " + err.Error()}, nil
	}
	s.cfg = cfg
	s.smsClient = client
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
	core *coreServer
}

func (s *smsServer) Send(ctx context.Context, req *pluginv1.SendSmsRequest) (*pluginv1.SendSmsResponse, error) {
	if s.core == nil || s.core.smsClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "plugin not initialized")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "missing request")
	}
	if strings.TrimSpace(req.GetTemplateId()) == "" {
		if strings.TrimSpace(req.GetContent()) != "" {
			return nil, status.Error(codes.InvalidArgument, "tencent sms requires template_id (content-only not supported)")
		}
		return nil, status.Error(codes.InvalidArgument, "template_id required")
	}
	if len(req.GetPhones()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "phones required")
	}

	phones := make([]*string, 0, len(req.GetPhones()))
	for _, p := range req.GetPhones() {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "+") {
			p = "+86" + p
		}
		phones = append(phones, common.StringPtr(p))
	}
	if len(phones) == 0 {
		return nil, status.Error(codes.InvalidArgument, "phones required")
	}

	var params []*string
	if len(req.GetVars()) > 0 {
		keys := make([]int, 0, len(req.GetVars()))
		valByIdx := map[int]string{}
		for k, v := range req.GetVars() {
			k = strings.TrimSpace(k)
			if k == "" {
				continue
			}
			var idx int
			if _, err := fmt.Sscanf(k, "%d", &idx); err != nil || idx <= 0 {
				return nil, status.Error(codes.InvalidArgument, "vars keys must be numeric strings starting from 1 (Tencent template param order)")
			}
			keys = append(keys, idx)
			valByIdx[idx] = v
		}
		sort.Ints(keys)
		for _, idx := range keys {
			params = append(params, common.StringPtr(valByIdx[idx]))
		}
	}

	r := sms.NewSendSmsRequest()
	r.SmsSdkAppid = common.StringPtr(strings.TrimSpace(s.core.cfg.SdkAppID))
	r.Sign = common.StringPtr(strings.TrimSpace(s.core.cfg.SignName))
	r.TemplateID = common.StringPtr(strings.TrimSpace(req.GetTemplateId()))
	r.PhoneNumberSet = phones
	if len(params) > 0 {
		r.TemplateParamSet = params
	}

	resp, err := s.core.smsClient.SendSms(r)
	if err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "unauthorized") || strings.Contains(strings.ToLower(msg), "auth") {
			return nil, status.Error(codes.PermissionDenied, "tencent sms auth failed")
		}
		return nil, status.Error(codes.Unavailable, "tencent sms send failed: "+sanitizeErr(err))
	}
	if resp == nil || resp.Response == nil {
		return nil, status.Error(codes.Unavailable, "tencent sms empty response")
	}
	if resp.Response.SendStatusSet != nil && len(resp.Response.SendStatusSet) > 0 {
		st := resp.Response.SendStatusSet[0]
		if st == nil || st.Code == nil || *st.Code != "Ok" {
			code := ""
			if st != nil && st.Code != nil {
				code = *st.Code
			}
			msg := ""
			if st != nil && st.Message != nil {
				msg = *st.Message
			}
			if msg == "" {
				msg = code
			}
			return nil, status.Error(codes.FailedPrecondition, "tencent sms rejected: "+msg)
		}
		messageID := ""
		if st.SerialNo != nil {
			messageID = *st.SerialNo
		}
		return &pluginv1.SendSmsResponse{Ok: true, MessageId: messageID}, nil
	}
	return &pluginv1.SendSmsResponse{Ok: true, MessageId: ""}, nil
}

func main() {
	core := &coreServer{}
	sms := &smsServer{core: core}
	pluginsdk.Serve(map[string]pluginsdk.Plugin{
		pluginsdk.PluginKeyCore: &pluginsdk.CoreGRPCPlugin{Impl: core},
		pluginsdk.PluginKeySMS:  &pluginsdk.SmsGRPCPlugin{Impl: sms},
	})
}

func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 400 {
		s = s[:400]
	}
	return s
}
