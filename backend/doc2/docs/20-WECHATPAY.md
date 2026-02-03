# 20 - WeChat Pay API v3 plugin guide (Go)

## Use official Go SDK
- Repo: https://github.com/wechatpay-apiv3/wechatpay-go
Use it for:
- request signing + cert handling
- notify signature verification + resource decryption

## Minimal methods to implement
Recommended methods (ListMethods):
- `wechat_native` (returns QR code URL)
- `wechat_jsapi` (returns JSAPI params; requires payer openid)

## CreatePayment mapping
### Native
- Call v3 native order API -> get `code_url`
- Plugin returns `pay_url = code_url`

### JSAPI
- Call v3 jsapi order API -> get `prepay_id`
- Return `extra` JSON for frontend:
  - appId, timeStamp, nonceStr, package="prepay_id=xxx", signType="RSA", paySign
- `pay_url` can be empty

## VerifyNotify (CRITICAL)
- Use SDK to:
  1) verify request signature (Wechatpay-* headers)
  2) decrypt resource (AEAD_AES_256_GCM)
- Parsed fields:
  - out_trade_no -> OrderNo
  - transaction_id -> TradeNo
  - amount.total (fen) -> AmountFen
  - trade_state -> map to Status

## Status mapping (suggested)
- SUCCESS -> PAID
- NOTPAY/USERPAYING -> PENDING
- CLOSED/REVOKED/PAYERROR -> FAILED or CLOSED (align with host enum)

## Config schema (suggested)
Required:
- mch_id
- merchant_private_key (pem, secret)
- merchant_cert_serial_no
- api_v3_key (secret)
- platform_cert (optional; SDK can download)
- app_id (for JSAPI)
Optional:
- notify_url_override (keep off by default; host should pass)
- timeout_sec, retry

