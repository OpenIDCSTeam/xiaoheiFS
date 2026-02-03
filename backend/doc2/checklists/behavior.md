# Behavior compatibility checklist (host <-> plugin)

## Payment plugins
- out_trade_no MUST be host OrderNo (never generate a new one)
- CreatePayment must:
  - use notify_url + return_url passed by host
  - return either pay_url (redirect/qr) OR extra (JSAPI params) depending on method
- VerifyNotify must:
  - verify signature (and decrypt if needed)
  - parse and return: OrderNo, TradeNo, Amount, Status, PaidAt
  - return AckBody EXACTLY as provider requires (plain text)
- Amount normalization:
  - If provider uses yuan string "1.00", host uses fen int64: AmountFen = round(yuan*100)
  - Never trust float; parse decimal string precisely.
- Status mapping:
  - Success only when provider indicates success (EZPay: TRADE_SUCCESS)
  - Pending when not success and not closed/failed.
- Host method-level enable/disable:
  - ListMethods returns all supported methods
  - Host filters by enabled methods; plugin doesn't need to know switches

## SMS plugins
- Template SMS only; if host passes free-text content, return InvalidArgument with clear reason.
- Never log secrets.
- Provide SendResult with provider message id (if available).

## KYC plugins
- Use Start -> (URL/token) -> QueryResult flow
- Never "mock" with random success.
- Support "not supported" clearly (Unimplemented) and surface in UI tabs.

