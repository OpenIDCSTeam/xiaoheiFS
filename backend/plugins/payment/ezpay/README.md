# EZPay（易支付聚合网关）

## Methods
- `alipay`
- `wxpay`
- `qqpay`

支付 provider key 形式：`pluginID.method`，例如：
- `ezpay.alipay`
- `ezpay.wxpay`
- `ezpay.qqpay`

回调地址（宿主固定）：
- `POST/GET /api/v1/payments/notify/ezpay.alipay`
- `POST/GET /api/v1/payments/notify/ezpay.wxpay`
- `POST/GET /api/v1/payments/notify/ezpay.qqpay`

## CreatePayment 行为（关键）
- `out_trade_no` 必须等于宿主订单号（OrderNo）；插件不生成新订单号
- `notify_url` / `return_url` 必须使用宿主传入（不要求用户在插件配置里手填）
- 返回：`extra.pay_kind=form` + `extra.form_html`（前端打开新窗口并写入该 HTML 发起支付）

## 签名兼容（对齐 PHP SDK）
插件支持两种常见 key 拼接口径（`sign_key_mode`）：
- `plain`：`md5(sign_str + merchant_key)`（PHP SDK 默认）
- `amp_key`：`md5(sign_str + "&key=" + merchant_key)`

## VerifyNotify（回调验签）
- 严格 MD5 验签；只在 `trade_status == TRADE_SUCCESS` 时返回 `PAID`
- 成功时返回 `ack_body=success`（纯文本，精确）

## QueryPayment
支持两类查询接口（由 `query_api_url` 决定）：
- `POST /api/findorder`（`type=1&order_no=...`）
- `GET /api.php?act=order`（对齐 `doc2/易支付sdk/query.php`：`act=order&pid=...&key=...&out_trade_no=...`）
