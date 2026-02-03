package main

import (
	"context"
	"net/url"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pluginv1 "xiaoheiplay/plugin/v1"
)

func TestEZPay_CreatePayment_FormHTML_PlainSignModeMatchesPHPSDK(t *testing.T) {
	core := &coreServer{}
	core.cfg = config{
		GatewayBaseURL: "https://www.ezfpy.cn",
		SubmitPath:     "mapi.php",
		PID:            "10001",
		MerchantKey:    "testkey",
		SignType:       "MD5",
		SignKeyMode:    "plain",
	}
	pay := &payServer{core: core}

	resp, err := pay.CreatePayment(context.Background(), &pluginv1.CreatePaymentRpcRequest{
		Method: "wxpay",
		Request: &pluginv1.PaymentCreateRequest{
			OrderNo:   "ORDER-123",
			UserId:    "1",
			Amount:    1234,
			Currency:  "CNY",
			Subject:   "test",
			NotifyUrl: "https://host.example/api/v1/payments/notify/ezpay.wxpay",
			ReturnUrl: "https://host.example/return",
		},
	})
	if err != nil {
		t.Fatalf("CreatePayment err: %v", err)
	}
	if !resp.GetOk() {
		t.Fatalf("CreatePayment not ok: %v", resp.GetError())
	}
	if resp.GetPayUrl() != "" {
		t.Fatalf("expected empty pay_url (POST form flow), got %q", resp.GetPayUrl())
	}
	formHTML := resp.GetExtra()["form_html"]
	if formHTML == "" {
		t.Fatalf("expected extra.form_html")
	}
	if want := "ORDER-123"; !contains(formHTML, want) {
		t.Fatalf("form_html missing out_trade_no %q", want)
	}
	if !contains(formHTML, "/mapi.php") {
		t.Fatalf("form_html missing mapi.php")
	}
	if got := extractInputValue(formHTML, "sign"); got != "a283fed0dcdd2d3cf0376db2910e9d42" {
		t.Fatalf("unexpected sign: %q", got)
	}
}

func TestEZPay_CreatePayment_FormHTML_AmpKeySignModeMatchesPHPSDK(t *testing.T) {
	core := &coreServer{}
	core.cfg = config{
		GatewayBaseURL: "https://www.ezfpy.cn",
		SubmitPath:     "mapi.php",
		PID:            "10001",
		MerchantKey:    "testkey",
		SignType:       "MD5",
		SignKeyMode:    "amp_key",
	}
	pay := &payServer{core: core}

	resp, err := pay.CreatePayment(context.Background(), &pluginv1.CreatePaymentRpcRequest{
		Method: "wxpay",
		Request: &pluginv1.PaymentCreateRequest{
			OrderNo:   "ORDER-123",
			UserId:    "1",
			Amount:    1234,
			Currency:  "CNY",
			Subject:   "test",
			NotifyUrl: "https://host.example/api/v1/payments/notify/ezpay.wxpay",
			ReturnUrl: "https://host.example/return",
		},
	})
	if err != nil {
		t.Fatalf("CreatePayment err: %v", err)
	}
	if !resp.GetOk() {
		t.Fatalf("CreatePayment not ok: %v", resp.GetError())
	}
	formHTML := resp.GetExtra()["form_html"]
	if got := extractInputValue(formHTML, "sign"); got != "4a82b712ab298128dfbd508b90151079" {
		t.Fatalf("unexpected sign: %q", got)
	}
}

func TestEZPay_VerifyNotify_MD5_TwoKeyModes(t *testing.T) {
	core := &coreServer{}
	core.cfg = config{
		GatewayBaseURL: "https://www.ezfpy.cn",
		PID:            "10001",
		MerchantKey:    "testkey",
		SignType:       "MD5",
	}
	pay := &payServer{core: core}

	base := map[string]string{
		"pid":          core.cfg.PID,
		"out_trade_no": "ORDER-123",
		"trade_no":     "PLAT-999",
		"type":         "alipay",
		"money":        "12.34",
		"trade_status": "TRADE_SUCCESS",
		"sign_type":    "MD5",
	}

	cases := []struct {
		name string
		sign string
	}{
		{name: "plain", sign: "ff02336560b32f42db5808bb59c5f948"},
		{name: "amp_key", sign: "904edd6feed370c17f3ba6576912114f"},
	}
	for _, tc := range cases {
		params := map[string]string{}
		for k, v := range base {
			params[k] = v
		}
		params["sign"] = tc.sign
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		vr, err := pay.VerifyNotify(context.Background(), &pluginv1.VerifyNotifyRequest{
			Method: "alipay",
			Raw: &pluginv1.RawHttpRequest{
				Method:   "GET",
				RawQuery: q.Encode(),
			},
		})
		if err != nil {
			t.Fatalf("VerifyNotify err (%s): %v", tc.name, err)
		}
		if !vr.GetOk() {
			t.Fatalf("VerifyNotify not ok (%s): %v", tc.name, vr.GetError())
		}
		if vr.GetOrderNo() != "ORDER-123" {
			t.Fatalf("unexpected order_no: %q", vr.GetOrderNo())
		}
		if vr.GetTradeNo() != "PLAT-999" {
			t.Fatalf("unexpected trade_no: %q", vr.GetTradeNo())
		}
		if vr.GetAckBody() != "success" {
			t.Fatalf("unexpected ack_body: %q", vr.GetAckBody())
		}
		if vr.GetAmount() != 1234 {
			t.Fatalf("unexpected amount: %d", vr.GetAmount())
		}
	}
}

func TestEZPay_VerifyNotify_InvalidSign(t *testing.T) {
	core := &coreServer{}
	core.cfg = config{
		GatewayBaseURL: "https://www.ezfpy.cn",
		PID:            "10001",
		MerchantKey:    "testkey",
		SignType:       "MD5",
	}
	pay := &payServer{core: core}

	params := url.Values{}
	params.Set("pid", core.cfg.PID)
	params.Set("out_trade_no", "ORDER-123")
	params.Set("trade_no", "PLAT-999")
	params.Set("type", "wxpay")
	params.Set("money", "12.34")
	params.Set("trade_status", "TRADE_SUCCESS")
	params.Set("sign_type", "MD5")
	params.Set("sign", "bad")

	_, err := pay.VerifyNotify(context.Background(), &pluginv1.VerifyNotifyRequest{
		Method: "wxpay",
		Raw: &pluginv1.RawHttpRequest{
			Method:   "GET",
			RawQuery: params.Encode(),
		},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestEZPay_ParseMoneyToCentsStrict(t *testing.T) {
	cases := map[string]int64{
		"1.00":   100,
		"0.01":   1,
		"100":    10000,
		"100.10": 10010,
		"000.10": 10,
		"12.3":   1230,
		"12.30":  1230,
		"12.340": 1234,
	}
	for in, want := range cases {
		got, err := parseMoneyToCentsStrict(in)
		if err != nil {
			t.Fatalf("parseMoneyToCentsStrict(%q) err: %v", in, err)
		}
		if got != want {
			t.Fatalf("parseMoneyToCentsStrict(%q)=%d want %d", in, got, want)
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (stringIndex(s, sub) >= 0))
}

func stringIndex(s, sub string) int {
	// tiny helper to avoid importing strings in tests (keep explicit deps minimal)
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func extractInputValue(html string, name string) string {
	needle := "name=\"" + name + "\" value=\""
	idx := stringIndex(html, needle)
	if idx < 0 {
		return ""
	}
	start := idx + len(needle)
	end := start
	for end < len(html) && html[end] != '"' {
		end++
	}
	if end <= start {
		return ""
	}
	return html[start:end]
}
