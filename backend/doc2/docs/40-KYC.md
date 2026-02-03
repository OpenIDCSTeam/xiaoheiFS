# 40 - KYC / eKYC plugins (Aliyun + Tencent)

Goal: real verification flow, not mock.

## Aliyun eKYC (ID Verification)
Typical flow:
1) Initialize -> returns transactionId (token)
2) User completes verification (if H5/SDK needed) OR server-side checks
3) CheckResult(transactionId) -> result

Plugin API mapping:
- Start => returns token=transactionId + (optional) url
- QueryResult => returns status + reason + raw_json

Config schema:
- access_key_id / access_key_secret
- region/endpoint
- product/app config ids if needed

## Tencent FaceID (Mobile HTML5)
Flow:
1) ApplyWebVerificationBizTokenIntl -> returns BizToken + VerificationURL
2) User opens URL and completes verification
3) GetWebVerificationResultIntl(BizToken) -> final result

Plugin mapping:
- Start => returns token=BizToken + url=VerificationURL
- QueryResult => queries BizToken and maps result

Config schema:
- secret_id / secret_key
- region
- app_id / rule_id or similar (depends on product settings)

