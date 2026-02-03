$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$SdkDir = Join-Path $Root "sdks"
New-Item -ItemType Directory -Force -Path $SdkDir | Out-Null
Set-Location $SdkDir

function CloneOrPull($url, $dir) {
  if (!(Test-Path $dir)) {
    git clone $url $dir
  } else {
    Push-Location $dir
    git pull
    Pop-Location
  }
}

Write-Host "Fetching open-source SDK repos..."
CloneOrPull "https://github.com/wechatpay-apiv3/wechatpay-go.git" "wechatpay-go"
CloneOrPull "https://github.com/TencentCloud/tencentcloud-sdk-go.git" "tencentcloud-sdk-go"
CloneOrPull "https://github.com/alibabacloud-go/dysmsapi-20170525.git" "dysmsapi-20170525"
CloneOrPull "https://github.com/alibabacloud-go/ekyc-saas-20230112.git" "ekyc-saas-20230112"
CloneOrPull "https://github.com/alipay/global-open-sdk-go.git" "alipay-global-open-sdk-go"

Write-Host "Done. SDK sources are in: $SdkDir"
