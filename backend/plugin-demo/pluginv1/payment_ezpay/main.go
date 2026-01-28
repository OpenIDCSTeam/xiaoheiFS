package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"xiaoheiplay/pkg/pluginsdk"
	pluginv1 "xiaoheiplay/plugin/v1"
)

type config struct {
	BaseURL   string `json:"base_url"`
	PID       string `json:"pid"`
	Key       string `json:"key"`
	PayType   string `json:"pay_type"`
	NotifyURL string `json:"notify_url"`
	ReturnURL string `json:"return_url"`
	SignType  string `json:"sign_type"`
}

type coreServer struct {
	pluginv1.UnimplementedCoreServiceServer

	cfg       config
	instance  string
	updatedAt time.Time
}

func (s *coreServer) GetManifest(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.Manifest, error) {
	_ = ctx
	return &pluginv1.Manifest{
		PluginId:    "ezpay",
		Name:        "EZPay Aggregator (Demo)",
		Version:     "0.1.0",
		Description: "Minimal demo payment plugin. YiPay works; WeChat/Alipay are stubs.",
		Payment:     &pluginv1.PaymentCapability{Methods: []string{"yipay", "wechatpay", "alipay"}},
	}, nil
}

func (s *coreServer) GetConfigSchema(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.ConfigSchema, error) {
	_ = ctx
	return &pluginv1.ConfigSchema{
		JsonSchema: `{
  "title": "EZPay Config",
  "type": "object",
  "properties": {
    "base_url": { "type": "string", "title": "Gateway URL", "default": "https://pays.org.cn/submit.php" },
    "pid": { "type": "string", "title": "Merchant PID" },
    "key": { "type": "string", "title": "Merchant Key", "format": "password" },
    "pay_type": { "type": "string", "title": "Pay Type", "default": "alipay" },
    "notify_url": { "type": "string", "title": "Notify URL" },
    "return_url": { "type": "string", "title": "Return URL" },
    "sign_type": { "type": "string", "title": "Sign Type", "default": "MD5" }
  },
  "required": ["base_url","pid","key"]
}`,
		UiSchema: `{
  "key": { "ui:widget": "password", "ui:help": "留空表示不修改（由宿主处理）" }
}`,
	}, nil
}

func (s *coreServer) ValidateConfig(ctx context.Context, req *pluginv1.ValidateConfigRequest) (*pluginv1.ValidateConfigResponse, error) {
	_ = ctx
	var cfg config
	if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "invalid json"}, nil
	}
	if strings.TrimSpace(cfg.BaseURL) == "" || strings.TrimSpace(cfg.PID) == "" || strings.TrimSpace(cfg.Key) == "" {
		return &pluginv1.ValidateConfigResponse{Ok: false, Error: "base_url/pid/key required"}, nil
	}
	return &pluginv1.ValidateConfigResponse{Ok: true}, nil
}

func (s *coreServer) Init(ctx context.Context, req *pluginv1.InitRequest) (*pluginv1.InitResponse, error) {
	if req.GetConfigJson() != "" {
		var cfg config
		if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
			return &pluginv1.InitResponse{Ok: false, Error: "invalid config"}, nil
		}
		s.cfg = cfg
	}
	s.instance = req.GetInstanceId()
	s.updatedAt = time.Now()
	return &pluginv1.InitResponse{Ok: true}, nil
}

func (s *coreServer) ReloadConfig(ctx context.Context, req *pluginv1.ReloadConfigRequest) (*pluginv1.ReloadConfigResponse, error) {
	_ = ctx
	var cfg config
	if err := json.Unmarshal([]byte(req.GetConfigJson()), &cfg); err != nil {
		return &pluginv1.ReloadConfigResponse{Ok: false, Error: "invalid config"}, nil
	}
	s.cfg = cfg
	s.updatedAt = time.Now()
	return &pluginv1.ReloadConfigResponse{Ok: true}, nil
}

func (s *coreServer) Health(ctx context.Context, req *pluginv1.HealthCheckRequest) (*pluginv1.HealthCheckResponse, error) {
	_ = ctx
	msg := "ok"
	if req.GetInstanceId() == "" || s.instance == "" {
		msg = "not initialized"
	}
	return &pluginv1.HealthCheckResponse{
		Status:     pluginv1.HealthStatus_HEALTH_STATUS_OK,
		Message:    msg,
		UnixMillis: time.Now().UnixMilli(),
	}, nil
}

type payServer struct {
	pluginv1.UnimplementedPaymentServiceServer
	core *coreServer
}

func (p *payServer) ListMethods(ctx context.Context, _ *pluginv1.Empty) (*pluginv1.ListMethodsResponse, error) {
	_ = ctx
	return &pluginv1.ListMethodsResponse{Methods: []string{"yipay", "wechatpay", "alipay"}}, nil
}

func (p *payServer) CreatePayment(ctx context.Context, req *pluginv1.CreatePaymentRpcRequest) (*pluginv1.PaymentCreateResponse, error) {
	_ = ctx
	method := strings.TrimSpace(req.GetMethod())
	switch method {
	case "yipay":
		return p.createYiPay(req.GetRequest())
	case "wechatpay", "alipay":
		return &pluginv1.PaymentCreateResponse{Ok: false, Error: "stub: method not implemented yet"}, nil
	default:
		return &pluginv1.PaymentCreateResponse{Ok: false, Error: "unknown method"}, nil
	}
}

func (p *payServer) QueryPayment(ctx context.Context, req *pluginv1.QueryPaymentRpcRequest) (*pluginv1.PaymentQueryResponse, error) {
	_ = ctx
	return &pluginv1.PaymentQueryResponse{
		Ok:      true,
		Status:  pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		TradeNo: req.GetTradeNo(),
	}, nil
}

func (p *payServer) Refund(ctx context.Context, req *pluginv1.RefundRpcRequest) (*pluginv1.RefundResponse, error) {
	_ = ctx
	_ = req
	return &pluginv1.RefundResponse{Ok: false, Error: "refund not implemented"}, nil
}

func (p *payServer) VerifyNotify(ctx context.Context, req *pluginv1.VerifyNotifyRequest) (*pluginv1.NotifyVerifyResult, error) {
	_ = ctx
	method := strings.TrimSpace(req.GetMethod())
	switch method {
	case "yipay":
		return p.verifyYiPay(req.GetRaw())
	case "wechatpay", "alipay":
		return &pluginv1.NotifyVerifyResult{Ok: false, Error: "stub: method not implemented yet"}, nil
	default:
		return &pluginv1.NotifyVerifyResult{Ok: false, Error: "unknown method"}, nil
	}
}

func (p *payServer) createYiPay(in *pluginv1.PaymentCreateRequest) (*pluginv1.PaymentCreateResponse, error) {
	if in == nil {
		return &pluginv1.PaymentCreateResponse{Ok: false, Error: "missing request"}, nil
	}
	cfg := p.core.cfg
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return &pluginv1.PaymentCreateResponse{Ok: false, Error: "config missing"}, nil
	}
	tradeNo := fmt.Sprintf("EZPAY-%s-%d", strings.TrimSpace(in.GetOrderNo()), time.Now().Unix())
	params := map[string]string{
		"pid":          cfg.PID,
		"type":         cfg.PayType,
		"out_trade_no": tradeNo,
		"name":         in.GetSubject(),
		"money":        fmt.Sprintf("%.2f", float64(in.GetAmount())/100.0),
		"notify_url":   firstNonEmpty(in.GetNotifyUrl(), cfg.NotifyURL),
		"return_url":   firstNonEmpty(in.GetReturnUrl(), cfg.ReturnURL),
		"sign_type":    firstNonEmpty(cfg.SignType, "MD5"),
	}
	params["sign"] = signYiPay(params, cfg.Key)
	payURL := buildURL(cfg.BaseURL, params)
	return &pluginv1.PaymentCreateResponse{
		Ok:      true,
		TradeNo: tradeNo,
		PayUrl:  payURL,
		Extra:   map[string]string{},
	}, nil
}

func (p *payServer) verifyYiPay(raw *pluginv1.RawHttpRequest) (*pluginv1.NotifyVerifyResult, error) {
	if raw == nil {
		return &pluginv1.NotifyVerifyResult{Ok: false, Error: "missing raw request"}, nil
	}
	cfg := p.core.cfg
	params := rawToParams(raw)
	sign := params["sign"]
	if sign == "" {
		return &pluginv1.NotifyVerifyResult{Ok: false, Error: "missing sign"}, nil
	}
	expected := signYiPay(params, cfg.Key)
	if !strings.EqualFold(sign, expected) {
		return &pluginv1.NotifyVerifyResult{Ok: false, Error: "invalid sign"}, nil
	}
	status := strings.ToLower(params["trade_status"])
	if status == "" {
		status = strings.ToLower(params["status"])
	}
	paid := status == "trade_success" || status == "success"
	ps := pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING
	if paid {
		ps = pluginv1.PaymentStatus_PAYMENT_STATUS_PAID
	}
	amountCents := parseMoneyToCents(params["money"])
	tradeNo := params["out_trade_no"]
	if tradeNo == "" {
		return &pluginv1.NotifyVerifyResult{Ok: false, Error: "missing out_trade_no"}, nil
	}
	rawJSON, _ := json.Marshal(params)
	return &pluginv1.NotifyVerifyResult{
		Ok:      true,
		OrderNo: tradeNo,
		TradeNo: tradeNo,
		Amount:  amountCents,
		Status:  ps,
		AckBody: "success",
		RawJson: string(rawJSON),
	}, nil
}

func rawToParams(req *pluginv1.RawHttpRequest) map[string]string {
	out := map[string]string{}
	if req.GetRawQuery() != "" {
		if q, err := url.ParseQuery(req.GetRawQuery()); err == nil {
			for k, v := range q {
				if len(v) > 0 {
					out[k] = v[0]
				}
			}
		}
	}
	if len(req.GetBody()) > 0 {
		ct := ""
		if v := req.GetHeaders()["Content-Type"]; v != nil && len(v.Values) > 0 {
			ct = v.Values[0]
		}
		if strings.Contains(strings.ToLower(ct), "application/x-www-form-urlencoded") || strings.Contains(string(req.GetBody()), "=") {
			if q, err := url.ParseQuery(string(req.GetBody())); err == nil {
				for k, v := range q {
					if len(v) > 0 && out[k] == "" {
						out[k] = v[0]
					}
				}
			}
		}
	}
	return out
}

func signYiPay(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		if strings.TrimSpace(params[k]) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(params[k])
	}
	buf.WriteString(key)
	sum := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(sum[:])
}

func buildURL(base string, params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		if strings.TrimSpace(v) == "" {
			continue
		}
		q.Set(k, v)
	}
	if strings.Contains(base, "?") {
		return base + "&" + q.Encode()
	}
	return base + "?" + q.Encode()
}

func parseMoneyToCents(v string) int64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if strings.Contains(v, ".") {
		parts := strings.SplitN(v, ".", 2)
		i, _ := strconv.ParseInt(parts[0], 10, 64)
		frac := parts[1]
		if len(frac) > 2 {
			frac = frac[:2]
		}
		for len(frac) < 2 {
			frac += "0"
		}
		f, _ := strconv.ParseInt(frac, 10, 64)
		return i*100 + f
	}
	i, _ := strconv.ParseInt(v, 10, 64)
	return i * 100
}

func firstNonEmpty(v string, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(fallback)
}

func main() {
	core := &coreServer{}
	pay := &payServer{core: core}

	pluginsdk.Serve(map[string]pluginsdk.Plugin{
		pluginsdk.PluginKeyCore:    &pluginsdk.CoreGRPCPlugin{Impl: core},
		pluginsdk.PluginKeyPayment: &pluginsdk.PaymentGRPCPlugin{Impl: pay},
	})
}
