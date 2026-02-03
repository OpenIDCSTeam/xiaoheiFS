# tencent_kyc（腾讯云 FaceID 实名/eKYC）

## 能力

- `Start`：发起认证（DetectAuth）
- `QueryResult`：查询结果（GetDetectInfo）

## 配置项（插件管理页 -> 配置）

此插件的 schema 由插件进程 `CoreService.GetConfigSchema()` 返回（目录下不提供静态 schema 文件）。主要字段：

- `secret_id`
- `secret_key`（secret）
- `rule_id`
- `region`（默认 `ap-guangzhou`）
- `timeout_sec`（默认 10）

## Start / QueryResult 约定

- `Start` 需要 `params.name` 与 `params.id_number`，可选：
  - `redirect_url`：完成后跳转 URL
  - `extra`：回传透传字段
  - `image_base64`：可选（具体能力取决于 RuleId 配置）
- `Start` 返回 token 为 `BizToken`，并返回 `url`（H5 认证地址）。

