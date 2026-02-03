# 00 - START HERE

You are implementing real plugins for a host system with:
- HashiCorp go-plugin (gRPC) + protobuf services in plugin/v1
- Host passes raw HTTP request (headers/body/query/method/path) into VerifyNotify
- Host expects:
  - OrderNo == host order id (use out_trade_no, never create a new one)
  - TradeNo == provider transaction id when available (trade_no/transaction_id)
  - Amount == normalized (define if host uses fen/int64; if host uses yuan/decimal, adapt consistently)
  - Status mapping to host enum
  - AckBody == exact required response text for the provider callback (usually plain 'success')

## The hard rules
1) **Never mock** (no fixed values). If you don't support an operation, return Unimplemented and mark capability.
2) **Notify/Return URLs are host-generated**. Plugins must accept notify_url/return_url from host inputs and use them in CreatePayment.
3) **Signature verification must be strict and tested** (unit tests + sample vectors).
4) **Idempotent notify**: receiving the same notify twice must still verify and return OK + same parsed result.

## Recommended implementation order (to keep sanity)
1) Fix EZPay (易支付) first (it's small + your current plugin is wrong).
2) Then implement WeChatPay v3 official.
3) Then Alipay (domestic) + SMS + KYC.

Open these in order:
- docs/10-EZPAY.md
- docs/20-WECHATPAY.md
- docs/21-ALIPAY.md
- docs/30-SMS.md
- docs/40-KYC.md
- checklists/behavior.md
