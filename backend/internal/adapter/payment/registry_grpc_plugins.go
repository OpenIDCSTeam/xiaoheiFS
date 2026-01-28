package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	plugins "xiaoheiplay/internal/adapter/plugins"
	"xiaoheiplay/internal/usecase"
	pluginv1 "xiaoheiplay/plugin/v1"
)

type grpcPaymentProvider struct {
	mgr      *plugins.Manager
	category string
	pluginID string
	method   string
	name     string
}

func (p *grpcPaymentProvider) Key() string {
	return p.pluginID + "." + p.method
}

func (p *grpcPaymentProvider) Name() string {
	if p.name == "" {
		return p.Key()
	}
	return p.name + " / " + p.method
}

func (p *grpcPaymentProvider) SchemaJSON() string { return "" }

func (p *grpcPaymentProvider) CreatePayment(ctx context.Context, req usecase.PaymentCreateRequest) (usecase.PaymentCreateResult, error) {
	if p.mgr == nil {
		return usecase.PaymentCreateResult{}, errors.New("plugin manager missing")
	}
	client, ok := p.mgr.GetPaymentClient(p.category, p.pluginID)
	if !ok {
		return usecase.PaymentCreateResult{}, usecase.ErrForbidden
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resp, err := client.CreatePayment(cctx, &pluginv1.CreatePaymentRpcRequest{
		Method: p.method,
		Request: &pluginv1.PaymentCreateRequest{
			OrderNo:   req.OrderNo,
			UserId:    fmt.Sprintf("%d", req.UserID),
			Amount:    req.Amount,
			Currency:  req.Currency,
			Subject:   req.Subject,
			ReturnUrl: req.ReturnURL,
			NotifyUrl: req.NotifyURL,
			Extra:     map[string]string{},
		},
	})
	if err != nil {
		return usecase.PaymentCreateResult{}, err
	}
	if resp != nil && !resp.Ok {
		if resp.Error != "" {
			return usecase.PaymentCreateResult{}, errors.New(resp.Error)
		}
		return usecase.PaymentCreateResult{}, errors.New("create payment failed")
	}
	return usecase.PaymentCreateResult{
		TradeNo: resp.TradeNo,
		PayURL:  resp.PayUrl,
		Extra:   resp.Extra,
	}, nil
}

func (p *grpcPaymentProvider) VerifyNotify(ctx context.Context, req usecase.RawHTTPRequest) (usecase.PaymentNotifyResult, error) {
	if p.mgr == nil {
		return usecase.PaymentNotifyResult{}, errors.New("plugin manager missing")
	}
	client, ok := p.mgr.GetPaymentClient(p.category, p.pluginID)
	if !ok {
		return usecase.PaymentNotifyResult{}, usecase.ErrForbidden
	}
	headers := map[string]*pluginv1.StringList{}
	for k, v := range req.Headers {
		copied := make([]string, len(v))
		copy(copied, v)
		headers[k] = &pluginv1.StringList{Values: copied}
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resp, err := client.VerifyNotify(cctx, &pluginv1.VerifyNotifyRequest{
		Method: p.method,
		Raw: &pluginv1.RawHttpRequest{
			Method:   req.Method,
			Path:     req.Path,
			RawQuery: req.RawQuery,
			Headers:  headers,
			Body:     req.Body,
		},
	})
	if err != nil {
		return usecase.PaymentNotifyResult{}, err
	}
	if resp != nil && !resp.Ok {
		if resp.Error != "" {
			return usecase.PaymentNotifyResult{}, errors.New(resp.Error)
		}
		return usecase.PaymentNotifyResult{}, errors.New("verify notify failed")
	}
	paid := resp.Status == pluginv1.PaymentStatus_PAYMENT_STATUS_PAID
	raw := map[string]string{
		"order_no": resp.OrderNo,
		"raw_json": resp.RawJson,
	}
	return usecase.PaymentNotifyResult{
		TradeNo: resp.TradeNo,
		Paid:    paid,
		Amount:  resp.Amount,
		Raw:     raw,
		AckBody: resp.AckBody,
	}, nil
}

func (r *Registry) grpcProviders(ctx context.Context) []usecase.PaymentProvider {
	items, err := r.grpcPlugins.List(ctx)
	if err != nil {
		return nil
	}
	var out []usecase.PaymentProvider
	for _, it := range items {
		if !it.Enabled || it.Capabilities.Capabilities.Payment == nil {
			continue
		}
		methods := it.Capabilities.Capabilities.Payment.Methods
		for _, m := range methods {
			m = strings.TrimSpace(m)
			if m == "" || strings.Contains(m, ".") {
				continue
			}
			if _, ok := r.grpcPlugins.GetPaymentClient(it.Category, it.PluginID); !ok {
				continue
			}
			out = append(out, &grpcPaymentProvider{
				mgr:      r.grpcPlugins,
				category: it.Category,
				pluginID: it.PluginID,
				method:   m,
				name:     it.Name,
			})
		}
	}
	return out
}

func (r *Registry) grpcProviderByKey(ctx context.Context, key string) usecase.PaymentProvider {
	parts := strings.SplitN(strings.TrimSpace(key), ".", 2)
	if len(parts) != 2 {
		return nil
	}
	pluginID := strings.TrimSpace(parts[0])
	method := strings.TrimSpace(parts[1])
	if pluginID == "" || method == "" {
		return nil
	}
	items, err := r.grpcPlugins.List(ctx)
	if err != nil {
		return nil
	}
	for _, it := range items {
		if !it.Enabled || it.PluginID != pluginID || it.Capabilities.Capabilities.Payment == nil {
			continue
		}
		for _, m := range it.Capabilities.Capabilities.Payment.Methods {
			if m == method {
				if _, ok := r.grpcPlugins.GetPaymentClient(it.Category, it.PluginID); !ok {
					return nil
				}
				return &grpcPaymentProvider{
					mgr:      r.grpcPlugins,
					category: it.Category,
					pluginID: it.PluginID,
					method:   method,
					name:     it.Name,
				}
			}
		}
	}
	return nil
}
