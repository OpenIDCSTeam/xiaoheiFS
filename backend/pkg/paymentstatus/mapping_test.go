package paymentstatus

import (
	"testing"

	pluginv1 "xiaoheiplay/plugin/v1"
)

func TestWeChatTradeStateToStatus(t *testing.T) {
	cases := map[string]pluginv1.PaymentStatus{
		"SUCCESS":    pluginv1.PaymentStatus_PAYMENT_STATUS_PAID,
		"NOTPAY":     pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"USERPAYING": pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"REFUND":     pluginv1.PaymentStatus_PAYMENT_STATUS_REFUNDING,
		"CLOSED":     pluginv1.PaymentStatus_PAYMENT_STATUS_CLOSED,
		"REVOKED":    pluginv1.PaymentStatus_PAYMENT_STATUS_CLOSED,
		"PAYERROR":   pluginv1.PaymentStatus_PAYMENT_STATUS_FAILED,
		"":           pluginv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED,
	}
	for in, want := range cases {
		if got := WeChatTradeStateToStatus(in); got != want {
			t.Fatalf("WeChatTradeStateToStatus(%q)=%v want %v", in, got, want)
		}
	}
}

func TestAlipayTradeStatusToStatus(t *testing.T) {
	cases := map[string]pluginv1.PaymentStatus{
		"TRADE_SUCCESS":  pluginv1.PaymentStatus_PAYMENT_STATUS_PAID,
		"TRADE_FINISHED": pluginv1.PaymentStatus_PAYMENT_STATUS_PAID,
		"WAIT_BUYER_PAY": pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"TRADE_CLOSED":   pluginv1.PaymentStatus_PAYMENT_STATUS_CLOSED,
		"UNKNOWN":        pluginv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED,
		"":               pluginv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED,
	}
	for in, want := range cases {
		if got := AlipayTradeStatusToStatus(in); got != want {
			t.Fatalf("AlipayTradeStatusToStatus(%q)=%v want %v", in, got, want)
		}
	}
}

func TestEZPayStatusToStatus(t *testing.T) {
	cases := map[string]pluginv1.PaymentStatus{
		"trade_success": pluginv1.PaymentStatus_PAYMENT_STATUS_PAID,
		"success":       pluginv1.PaymentStatus_PAYMENT_STATUS_PAID,
		"notpay":        pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"wait":          pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"pending":       pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"userpaying":    pluginv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		"closed":        pluginv1.PaymentStatus_PAYMENT_STATUS_CLOSED,
		"fail":          pluginv1.PaymentStatus_PAYMENT_STATUS_FAILED,
		"failed":        pluginv1.PaymentStatus_PAYMENT_STATUS_FAILED,
		"":              pluginv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED,
	}
	for in, want := range cases {
		if got := EZPayStatusToStatus(in); got != want {
			t.Fatalf("EZPayStatusToStatus(%q)=%v want %v", in, got, want)
		}
	}
}
