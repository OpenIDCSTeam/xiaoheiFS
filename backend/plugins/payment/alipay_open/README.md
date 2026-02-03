# alipay_open（支付宝开放平台）

## Methods
- `alipay_wap`：手机网站支付（WAP）
- `alipay_page`：电脑网站支付（PAGE）

支付 provider key 形式：`pluginID.method`，例如：
- `alipay_open.alipay_wap`
- `alipay_open.alipay_page`

回调地址（宿主固定）：
- `POST/GET /api/v1/payments/notify/alipay_open.alipay_wap`
- `POST/GET /api/v1/payments/notify/alipay_open.alipay_page`

## 配置（插件管理页 -> 配置）
- `app_id`：应用 AppID
- `app_private_key`：应用私钥（支持“完整 PEM”或“仅 base64 内容”）
- `alipay_public_key`：支付宝公钥（支持“完整 PEM”或“仅 base64 内容”，用于 notify 验签）
- `seller_id`（可选）：卖家 ID（配置后会在回调中校验一致性）
- `is_prod`：是否生产环境（默认 true）
- `default_notify_url`（可选）：为空时使用宿主传入的 `notify_url`
- `default_return_url`（可选）：为空时使用宿主传入的 `return_url`
- `timeout_sec`（可选）：请求超时（秒）

注意：`notify_url` / `return_url` 必须由宿主生成并在 `CreatePayment` 里传入；插件配置里不要求手填。

## CreatePayment 行为（宿主约定）
- `out_trade_no` 必须等于宿主订单号（OrderNo）；插件不生成新订单号
- 返回 `pay_url`（前端直接跳转）

## VerifyNotify（回调验签）
- 插件使用 RSA2 验签（宿主只透传原始 HTTP 请求）
- 成功时返回 `ack_body=success`（纯文本，精确）
- 返回字段：
  - `order_no = out_trade_no`
  - `trade_no = trade_no`（支付宝平台流水号）
