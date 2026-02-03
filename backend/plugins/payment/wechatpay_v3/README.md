# wechatpay_v3（微信支付 API v3）

## Methods
- `wechat_native`：扫码支付（Native，下单后返回二维码链接）
- `wechat_jsapi`：公众号/小程序 JSAPI（下单后返回前端可直接调起的参数）

支付 provider key 形式：`pluginID.method`，例如：
- `wechatpay_v3.wechat_native`
- `wechatpay_v3.wechat_jsapi`

回调地址（宿主固定）：
- `POST/GET /api/v1/payments/notify/wechatpay_v3.wechat_native`
- `POST/GET /api/v1/payments/notify/wechatpay_v3.wechat_jsapi`

## 配置（插件管理页 -> 配置）
- `mch_id`：商户号 mchid
- `merchant_serial_no`：商户证书序列号（用于平台证书下载/验签）
- `merchant_private_key_pem`：商户私钥（支持“完整 PEM”或“仅 base64 内容”）
- `api_v3_key`：API v3 Key（32 字符）
- `app_id`：AppID
- `default_notify_url`（可选）：为空时使用宿主传入的 `notify_url`
- `default_return_url`（可选）：为空时使用宿主传入的 `return_url`
- `timeout_sec`（可选）：请求超时（秒）

注意：`notify_url` / `return_url` 必须由宿主生成并在 `CreatePayment` 里传入；插件配置里不要求手填。

## CreatePayment 返回（宿主约定）
- `wechat_native`：`extra.pay_kind=qr`，并返回 `pay_url`（code_url）
- `wechat_jsapi`：`extra.pay_kind=jsapi`，并返回 `extra.jsapi_params_json`

### JSAPI extra 约定
`wechat_jsapi` 必须传入：
- `extra.openid`：用户 openid

## VerifyNotify（验签与解密）
- 插件使用 `wechatpay-go` 完成：验签 + 资源解密 + 字段解析
- 成功时返回 `ack_body={"code":"SUCCESS","message":"SUCCESS"}`（宿主按此响应微信）
- `out_trade_no` 必须等于宿主订单号（OrderNo）；插件不生成新订单号
