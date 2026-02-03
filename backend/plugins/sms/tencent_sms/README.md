# tencent_sms（腾讯云短信）

## 能力

- `Send`：模板短信（TemplateId + TemplateParam）

## 配置项（插件管理页 -> 配置）

来自 `schemas/config.schema.json`：

- `secret_id`
- `secret_key`（secret）
- `sdk_app_id`
- `sign_name`

## Send 约定

- 按腾讯云短信 SendSms 接口发送模板短信。
- 变量参数需要按模板参数顺序传入；本插件会把 `vars` 视为有序参数（建议用 `"1".."n"` 作为 key）。

