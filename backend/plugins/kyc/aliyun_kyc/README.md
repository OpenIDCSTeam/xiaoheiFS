# aliyun_kyc（阿里云 CloudAuth 实名/eKYC）

## 能力

- `Start`：发起认证（InitSmartVerify）
- `QueryResult`：查询结果（DescribeSmartVerify）

## 配置项（插件管理页 -> 配置）

此插件的 schema 由插件进程 `CoreService.GetConfigSchema()` 返回（目录下不提供静态 schema 文件）。主要字段：

- `access_key_id`
- `access_key_secret`（secret）
- `scene_id`
- `region`（默认 `cn-hangzhou`）
- `endpoint`（默认 `cloudauth.cn-hangzhou.aliyuncs.com`）
- `mode`（默认 `FULL`）
- `h5_base_url`（可选）：用于拼接返回 `url`（例如 H5 页面基地址）
- `timeout_sec`（默认 10）

## Start / QueryResult 约定

- `Start` 需要 `params.name` 与 `params.id_number`。
- 返回 token 使用阿里云 `certifyId`（或同等交易号字段）；宿主应保存并用于后续查询。

