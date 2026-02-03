# alisms（阿里云短信）

## 能力

- `Send`：模板短信（TemplateCode + TemplateParam）

## 配置项（插件管理页 -> 配置）

来自 `schemas/config.schema.json`：

- `access_key_id`
- `access_key_secret`（secret）
- `sign_name`：短信签名

## Send 约定

本插件按阿里云短信产品能力发送“模板短信”。如果宿主传入 `content` 但未传 `template_id`，插件会返回参数错误（InvalidArgument），不会静默忽略。

