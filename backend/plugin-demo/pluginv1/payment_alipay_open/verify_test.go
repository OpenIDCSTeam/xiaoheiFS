package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"testing"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"

	pluginv1 "xiaoheiplay/plugin/v1"
)

func TestAlipay_VerifyNotify_RSA2(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	appID := "2026000000000000"
	sellerID := "2088000000000000"

	core := &coreServer{
		cfg: config{
			AppID:        appID,
			AliPublicKey: pubB64,
			SellerID:     sellerID,
		},
	}
	pay := &payServer{core: core}

	orderNo := "ORDER-123"
	aliTradeNo := "2026012900000000000000000000"

	// Signature string does NOT include sign/sign_type.
	bmSign := gopay.BodyMap{}
	bmSign.Set("out_trade_no", orderNo).
		Set("trade_no", aliTradeNo).
		Set("trade_status", "TRADE_SUCCESS").
		Set("total_amount", "12.34").
		Set("app_id", appID).
		Set("seller_id", sellerID)

	sign, err := alipay.GetRsaSign(bmSign, "RSA2", priv)
	if err != nil {
		t.Fatalf("GetRsaSign: %v", err)
	}
	bmSign.Set("sign_type", "RSA2")
	bmSign.Set("sign", sign)

	q := url.Values{}
	for k, v := range bmSign {
		q.Set(k, fmt.Sprint(v))
	}

	vr, err := pay.VerifyNotify(context.Background(), &pluginv1.VerifyNotifyRequest{
		Method: "alipay_wap",
		Raw: &pluginv1.RawHttpRequest{
			Method: "POST",
			Body:   []byte(q.Encode()),
			Headers: map[string]*pluginv1.StringList{
				"Content-Type": {Values: []string{"application/x-www-form-urlencoded"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("VerifyNotify err: %v", err)
	}
	if !vr.GetOk() {
		t.Fatalf("VerifyNotify not ok: %v", vr.GetError())
	}
	if vr.GetTradeNo() != aliTradeNo {
		t.Fatalf("unexpected trade_no: %q", vr.GetTradeNo())
	}
	if vr.GetOrderNo() != orderNo {
		t.Fatalf("unexpected order_no: %q", vr.GetOrderNo())
	}
	if vr.GetAckBody() != "success" {
		t.Fatalf("unexpected ack_body: %q", vr.GetAckBody())
	}
}
