# 插件系统（go-plugin GRPC + protobuf）迁移与最小闭环

本仓库已引入新的插件系统：HashiCorp `go-plugin`（`ProtocolGRPC`）+ protobuf/gRPC 协议（`plugin/v1`），并与旧支付插件（NetRPC）并存一段时间，便于平滑迁移。

## 1. 新插件目录结构（强制）

插件被安装到后端工作目录下的 `./plugins/<category>/<plugin_id>/`：

- `./plugins/sms/alisms/`
- `./plugins/kyc/aliyun_kyc/`
- `./plugins/payment/ezpay/`

每个插件目录至少包含：

- `manifest.json`
- `bin/<goos>_<goarch>/plugin(.exe)`（推荐：多平台二进制；由 `manifest.json.binaries` 指定）
- （兼容）根目录 `plugin.exe` / `plugin`（旧格式，仍可运行）
- `schemas/`（可选但推荐）
  - `config.schema.json`
  - `config.ui.json`（可选）
- `checksums.json`（可选）
- `signature.sig`（可选）

安装包支持：`.zip` / `.tar.gz` / `.tgz`。解压后必须落到上述目录结构。

### 1.1 多平台二进制（推荐）

`manifest.json` 支持 `binaries` 映射（key 为 `<goos>_<goarch>`，value 为相对路径）：

```json
{
  "plugin_id": "ezpay",
  "name": "易支付",
  "version": "1.0.0",
  "binaries": {
    "windows_amd64": "bin/windows_amd64/plugin.exe",
    "linux_amd64": "bin/linux_amd64/plugin"
  }
}
```

宿主运行时会按当前 `GOOS/GOARCH` 选择 entry 启动；若不支持，会提示支持的平台列表。

## 2. 后台管理入口（前端）

- 页面：`/admin/settings/plugins`
- 功能：统一管理所有类型插件（安装/卸载、启用/停用、配置、健康状态、能力查看）

安装安全规则：

- `signature_status=official`：一键安装
- `untrusted/unsigned`：会弹出管理员密码确认（后端复用现有管理员密码校验逻辑）

## 3. 配置规则（宿主统一存储 + 加密）

- 插件配置由宿主后端统一存储在 DB（字段 `plugin_installations.config_cipher`）
- 后端使用 `AES-GCM` 加密；密钥来自 `app.config.yaml` 的 `plugin_master_key`
  - 若未配置，启动时会自动生成并写入配置文件（行为与其他 secret 类似）
- 插件目录（`./plugins`）路径来自 `app.config.yaml` 的 `plugins_dir`（相对路径基于 `app.config.yaml` 所在目录解析）
- 前端配置表单来自插件 `CoreService.GetConfigSchema()` 返回的 `JSON Schema (+ 可选 UI Schema)`
- `secret` 字段：
  - Schema 中 `format=password` 或 `x-secret=true` 会被视为 secret
  - 前端显示为密码框，并提示“留空表示不修改”
  - 保存时如果 secret 字段为空，宿主会自动保留旧值（不会被清空）
- 保存配置前：宿主调用插件 `ValidateConfig` 校验，通过才落库；更新后调用 `ReloadConfig` 热更新（若插件正在运行）

## 4. 官方签名（Ed25519）

签名机制：

- `checksums.json`：记录目录内文件 sha256（由工具生成）
- `signature.sig`：对 `checksums.json` 的签名（Ed25519）
- 宿主通过 `plugin_official_ed25519_pubkeys` 配置的公钥列表来判定 `official`

补充约束（多平台二进制）：

- `checksums.json` 必须覆盖 `bin/**` 下的所有二进制文件；否则验签视为失败（`untrusted`）

### 4.1 生成/签名示例

对某个插件目录生成签名文件：

```powershell
cd d:\项目\golang\xiaohei\backend
go run ./cmd/tools/pluginsign -dir plugins/payment/ezpay
```

如果未传 `-ed25519-priv`，工具会生成一对 key 并打印 base64。把输出中的 **公钥** 写入 `app.config.yaml`：

```yaml
plugin_official_ed25519_pubkeys:
  - "BASE64_ED25519_PUBLIC_KEY"
```

再次签名（使用固定私钥）：

```powershell
go run ./cmd/tools/pluginsign -dir plugins/payment/ezpay -ed25519-priv "BASE64_ED25519_PRIVATE_KEY"
```

## 5. 最小闭环示例：易支付聚合支付插件（demo）

仓库提供可运行 demo 插件源码（go-plugin GRPC）：

- `plugin-demo/pluginv1/payment_ezpay/main.go`

并提供构建脚本（输出到 `./plugins/**/plugin.exe`）：

```powershell
cd d:\项目\golang\xiaohei\backend
powershell -ExecutionPolicy Bypass -File scripts/build-demo-plugins.ps1
```

构建脚本会把各平台二进制输出到：

- `./plugins/<category>/<plugin_id>/bin/<goos>_<goarch>/plugin(.exe)`

### 5.1 启用与配置

1) 后端 `air` 启动，前端 `npm run dev` 启动  
2) 打开后台 `/admin/settings/plugins`  
3) 找到 `payment/ezpay`：点击“配置”，填写 `key`（示例里是易支付 md5 密钥），保存  
4) 点击启用开关，观察健康状态变化

### 5.2 支付回调透传（关键点）

支付回调路由仍为：

- `POST/GET /payments/notify/:provider`

对于插件支付方式，`provider` 采用 `pluginID.method` 形式，例如：

- `ezpay.yipay`

宿主会把原始 HTTP 请求（headers/body/query/path/method）原封不动交给插件的 `PaymentService.VerifyNotify()`，由插件完成验签与解析，并返回统一结构与建议 ack 文本（例如 `success`）。

## 6. 从旧支付插件（NetRPC）迁移到新插件系统

现状：

- 旧支付插件系统仍保留（NetRPC），相关页面/接口后续将逐步废弃
- 新系统使用 GRPC + protobuf 协议（`plugin/v1`），支持 sms/kyc/payment 多类型与扩展

迁移建议步骤：

1) **保留旧系统运行**：先上线新系统但不切流量（仅安装/启用 demo 插件验证链路）  
2) **迁移插件二进制与目录**：
   - 将旧插件改造为 go-plugin GRPC 插件（参考 `plugin-demo/pluginv1/*`）
   - 打包为 zip/tar.gz 并通过后台“安装插件”安装到 `./plugins/<category>/<plugin_id>/`
3) **迁移配置数据**：
   - 旧系统配置通常由旧接口/文件管理；新系统配置在 DB 加密存储
   - 建议按插件逐个迁移：先在新页面“配置”填入同等参数并保存（ValidateConfig 会校验）
4) **迁移启用状态**：
   - 新系统启用状态存储在 `plugin_installations.enabled`
   - 逐个启用新插件并观察健康状态 OK
5) **兼容策略**（支付）：
   - 新插件支付方式以 `pluginID.method` 形式出现在支付 provider 列表中（宿主侧桥接）
   - 可逐步把业务侧 provider key 从旧 key 切换为新 key（例如 `ezpay.alipay`）

注意：

- 本版本支持启动时自动导入磁盘插件目录：首次启动（或 `plugins_bootstrapped=false`）会扫描 `./plugins/<category>/*/manifest.json` 并注册到 DB（默认 `enabled=false`）；后续启动仅自动导入新增的 **official** 插件，其它需要后台“导入/安装”并管理员密码确认。

## 7. automation（多轻舟）+ Goods Type（商品类型）

### 7.1 插件源码位置（仓库内）

本仓库内置/示例插件源码在：

- `plugin-demo/pluginv1/payment_ezpay/main.go`（支付 demo）
- `plugin-demo/pluginv1/automation_lightboat/main.go`（automation/lightboat demo，复用轻舟 API 逻辑）
- `plugin-demo/pluginv1/payment_wechatpay_v3/main.go`（微信支付 v3：wechatpay-go）
- `plugin-demo/pluginv1/payment_alipay_open/main.go`（支付宝开放平台：gopay）
- `plugin-demo/pluginv1/sms_alisms_mock/main.go`、`plugin-demo/pluginv1/sms_tencent_mock/main.go`（短信：阿里云/腾讯云）
- `plugin-demo/pluginv1/kyc_aliyun_mock/main.go`、`plugin-demo/pluginv1/kyc_tencent_mock/main.go`（实名/ekyc：阿里云/腾讯云）

插件运行时二进制安装目录在 `./plugins/**`（并非源码目录），例如 `./plugins/automation/lightboat/`。

### 7.2 多实例（同一个 plugin_id 多套配置）

`plugin_installations` 支持同一个 `(category, plugin_id)` 下创建多个 `instance_id`，每个实例独立配置/启停，但共享同一份插件文件（目录仍是 `./plugins/<category>/<plugin_id>/`）。

后台插件管理页（`/admin/settings/plugins`）：

- 同一插件可点击 “Add instance” 新增实例
- 启用/停用、配置、删除都作用在具体 `instance_id`

### 7.3 最小闭环（多轻舟 + 商品类型隔离）

手动验证步骤（示例）：

1) 构建 demo 插件二进制：

```powershell
cd d:\项目\golang\xiaohei\backend
powershell -ExecutionPolicy Bypass -File scripts/build-demo-plugins.ps1
```

2) 启动后端 `air` + 前端 `npm run dev`

3) 打开后台插件管理 `/admin/settings/plugins`
   - 找到 `automation/lightboat/default`，配置 `base_url/api_key`，启用
   - 点击 “Add instance” 再创建 `qz_a`、`qz_b` 两个实例，并分别配置不同 `base_url/api_key`，启用

4) 打开后台售卖配置 `/admin/catalog`
   - “商品类型” Tab：创建两个 Goods Type，例如 `轻舟VPS-A` 绑定 `automation_plugin_id=lightboat, automation_instance_id=qz_a`；`轻舟VPS-B` 绑定 `...=qz_b`
   - 在页面右上角选择对应 Goods Type，点击 “同步当前类型（merge）”，分别同步两次
   - 确认地区/线路/套餐在不同 Goods Type 下互不覆盖

5) 用户侧购买页 `/console/buy-vps`
   - 先选择 Goods Type，再选择地区/线路/套餐，创建订单并开通

## 8. 支付插件（真实对接）

本仓库内置 3 个支付插件（category=`payment`）：

- `wechatpay_v3`：方法 `wechat_native` / `wechat_jsapi`
- `alipay_open`：方法 `alipay_wap` / `alipay_page`
- `ezpay`：方法 `alipay` / `wechat` / `qq`

支付回调路由保持不变：

- `POST/GET /api/v1/payments/notify/:provider`

对于 go-plugin 支付方式，`provider` 采用 `pluginID.method` 形式，例如：

- `wechatpay_v3.wechat_native`
- `alipay_open.alipay_wap`
- `ezpay.wxpay`

### 8.1 微信支付（JSAPI）extra 约定

`wechat_jsapi` 需要前端在发起支付时通过 `extra.openid` 传入用户 openid；宿主会把 `extra` 原样透传给插件。

插件返回：

- Native：`extra.pay_kind=qr` + `extra.code_url`（前端展示二维码）
- JSAPI：`extra.pay_kind=jsapi` + `extra.jsapi_params_json`（前端在微信内调起 `WeixinJSBridge.invoke`）

### 8.2 回调 ack_body

- 微信支付 v3：插件返回 JSON `ack_body`（宿主会按 JSON 响应 Content-Type）
- 支付宝/易支付：插件返回 `ack_body=success`

## 9. 支付 method 开关（宿主管理）

插件 `ListMethods` 返回“该插件支持的所有方法”，但是否允许业务侧使用由宿主 DB 管理：

- 后台页面：`/admin/settings/plugins` -> 选择支付插件实例 -> `Methods`
- API：
  - `GET /admin/api/v1/plugins/payment-methods?category=payment&plugin_id=...&instance_id=...`
  - `PATCH /admin/api/v1/plugins/payment-methods`
