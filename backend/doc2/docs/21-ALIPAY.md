# 21 - Alipay (domestic open platform) plugin guide (Go)

## Reality check
Alipay EasySDK (official) is not Go-first for domestic alipay.trade.*.
For Go you typically:
- use a mature library OR
- implement RSA2 signing/verify strictly by OpenDocs

## Methods (ListMethods)
- `alipay_wap` (mobile browser)
- `alipay_page` (PC web)
(optional `alipay_app` if you support it)

## CreatePayment
- WAP/Page usually returns:
  - a URL to redirect OR
  - an HTML form that auto-submits (recommended because many gateways require POST)
Plugin should return one of:
- pay_url (if host supports redirect)
- form_html (auto submit)

## VerifyNotify (CRITICAL)
Must:
1) verify RSA2 signature with Alipay public key
2) validate fields:
   - app_id matches config
   - out_trade_no maps to host order
   - total_amount matches
   - trade_status indicates success (e.g., TRADE_SUCCESS / TRADE_FINISHED)
3) return AckBody = "success" (Alipay expects plain 'success')

## QueryPayment / Refund
- alipay.trade.query
- alipay.trade.refund

## Config schema
Required:
- app_id
- merchant_private_key (RSA2, secret)
- alipay_public_key
Optional:
- gateway_url (default https://openapi.alipay.com/gateway.do)
- notify_url_override (prefer host-generated)
- return_url_override
- charset=utf-8, sign_type=RSA2 fixed

