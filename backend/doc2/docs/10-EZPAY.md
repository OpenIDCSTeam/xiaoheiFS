# 10 - EZPay (易支付) implementation guide (based on EZFpy docs UI screenshots)

This guide is intentionally *actionable* (not just links). It matches the fields shown in your screenshots.

## Endpoints (as shown)
### Page redirect pay (页面跳转支付)
- Method: POST
- URL: `https://www.ezfpy.cn/submit.php` (or your own gateway base URL + `/submit.php`)

### Payment result notify (支付结果通知)
- Method: GET
- Called by EZPay to your:
  - `notify_url` (server-to-server async)
  - `return_url` (browser redirect)

### Single order query (单笔订单查询)
- Method: POST
- URL: `https://www.ezfpy.cn/api/findorder`

## 1) CreatePayment (page redirect flow)
You must build a form/query request with these required fields:

| Field | Param | Required | Example | Notes |
|---|---|---:|---|---|
| Merchant ID | pid | ✅ | 1000 | From config |
| Pay type | type | ✅ | alipay | **fixed values**: `alipay`, `wxpay`, `qqpay` |
| Merchant order no | out_trade_no | ✅ | 20160806151343349 | MUST be **host OrderNo** |
| Async notify URL | notify_url | ✅ | http://your.site/payments/notify/ezpay.wxpay | host-generated |
| Return URL | return_url | ✅ | http://your.site/pay/return?... | host-generated |
| Product name | name | ✅ | 一个奥利奥 | from host (order title) |
| Amount (yuan) | money | ✅ | 1.00 | decimal string |
| Website name | sitename | (doc says can be empty) | 站点名 | optional |
| Sign type | sign_type | ✅ | MD5 | default MD5 |
| Signature | sign | ✅ | md5... | computed |

### Method mapping (3 channels)
Implement ListMethods returning:
- `alipay`
- `wxpay`
- `qqpay`

CreatePayment chooses `type` based on method:
- method `alipay` -> type=`alipay`
- method `wxpay` -> type=`wxpay`
- method `qqpay` -> type=`qqpay`

### Signature (MD5)
The doc UI says: "签名算法与支付宝签名算法相同". In common EZPay/EPay implementations, the MD5 signature is:

1. Take all request params **excluding** `sign` and `sign_type`
2. Remove empty values
3. Sort by parameter name ascending (ASCII)
4. Build query string: `k1=v1&k2=v2&...`
5. Append merchant key: `...&key=MERCHANT_KEY`  (some providers append `MERCHANT_KEY` without `key=`; see below)
6. MD5, output lower-case hex

⚠️ Providers differ slightly in step 5. To avoid mismatch, do this:
- Implement **one canonical** signing method first: `...&key=MERCHANT_KEY`
- Add a fallback verifier for notify that also checks the alternative: `... + MERCHANT_KEY` (no `&key=`)
- Unit test both; accept whichever matches (signature still requires key, so this doesn't weaken security materially).

### Return to host
Your plugin CreatePayment should return:
- `pay_url`: either a GET redirect URL, or an HTML form submission URL, depending on host expectation.
Recommended simplest:
- Return a **GET URL** to the gateway where params are query-encoded (if provider accepts GET). If they require POST, return `form_html` (auto-submit form).

For safety across gateways:
- Return `form_html`: a small HTML page with an auto-submitting form POSTing to `/submit.php` with params.

## 2) VerifyNotify (payment result notification)
EZPay will call your `notify_url` and `return_url` with these parameters (shown):

| Param | Meaning |
|---|---|
| pid | merchant id |
| trade_no | platform order id |
| out_trade_no | your merchant order id |
| type | alipay/wxpay/qqpay |
| name | product name |
| money | amount in yuan string |
| trade_status | payment status (success only if `TRADE_SUCCESS`) |
| sign | signature |
| sign_type | MD5 |

### Verification steps (MUST)
1) Parse query params (GET)
2) Verify signature (MD5) using the same signing rules (exclude sign/sign_type)
3) Check `trade_status == "TRADE_SUCCESS"` => success. Otherwise treat as pending/failed (host mapping).
4) Validate order exists: out_trade_no maps to a host order
5) Validate amount equals your order amount (normalize decimals)

### What to return (AckBody)
Common EZPay expects plain text `success` for notify_url ack.
Implement:
- On valid success notify: return `AckBody = "success"` (exact)
- On invalid signature / mismatch: return `AckBody = "fail"` (or empty) and `ok=false` (host should respond accordingly)

### Mapping to host fields
- OrderNo = out_trade_no
- TradeNo = trade_no
- Amount = parse money (yuan) -> host amount (prefer fen int64 or decimal string)
- Status:
  - TRADE_SUCCESS -> PAID
  - else -> PENDING (or FAILED/CLOSED if doc provides such values; your screenshot states only TRADE_SUCCESS is success)

## 3) QueryPayment (单笔订单查询)
Endpoint: POST `.../api/findorder`

Request params (shown):
- `order_no` (required): order id to query
- `type` (required int): **1=merchant order no** , 2=platform order no

Response (shown):
- `code` int: 200 success
- `msg` string
- `data` array: order info objects

### How to use in plugin
Implement QueryPayment as:
- call findorder with type=1 and order_no=out_trade_no
- if code!=200 -> return error
- parse data[0].trade_status etc and map

## 4) Config schema (what the UI should ask)
Your current UI is wrong. Correct config is:

Required:
- gateway_base_url (e.g. https://www.ezfpy.cn) OR submit_url (full https://.../submit.php)
- pid (int)
- merchant_key (secret string)

Optional/Advanced:
- sign_type (default MD5; keep hidden unless needed)
- query_api_url (default {gateway_base_url}/api/findorder)
- site_name default for sitename

Not needed / must be host-generated (read-only display ok):
- notify_url
- return_url

## 5) Common mistakes to ban (fix current plugin)
- ❌ generating a new out_trade_no (must use host OrderNo)
- ❌ asking user to fill notify_url/return_url
- ❌ letting user type "Type Param for ..." (type values are fixed: alipay/wxpay/qqpay)
- ❌ treating any trade_status as success (only TRADE_SUCCESS is success per screenshot)

