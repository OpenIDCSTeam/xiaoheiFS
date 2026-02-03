#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SDK_DIR="$ROOT_DIR/sdks"
mkdir -p "$SDK_DIR"
cd "$SDK_DIR"

clone_or_pull () {
  local url="$1"
  local dir="$2"
  if [ ! -d "$dir" ]; then
    git clone "$url" "$dir"
  else
    (cd "$dir" && git pull)
  fi
}

echo "Fetching open-source SDK repos..."
clone_or_pull "https://github.com/wechatpay-apiv3/wechatpay-go.git" "wechatpay-go"
clone_or_pull "https://github.com/TencentCloud/tencentcloud-sdk-go.git" "tencentcloud-sdk-go"
clone_or_pull "https://github.com/alibabacloud-go/dysmsapi-20170525.git" "dysmsapi-20170525"
clone_or_pull "https://github.com/alibabacloud-go/ekyc-saas-20230112.git" "ekyc-saas-20230112"
clone_or_pull "https://github.com/alipay/global-open-sdk-go.git" "alipay-global-open-sdk-go"

echo "Done. SDK sources in $SDK_DIR"
