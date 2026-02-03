# 30 - SMS plugins (Aliyun + Tencent)

## Aliyun SMS (Dysmsapi)
Use official Go SDK package for Dysmsapi.
SendSms requires:
- PhoneNumbers (comma-separated)
- SignName
- TemplateCode
- TemplateParam (JSON string)

Behavior constraints:
- If host sends free-text content, return InvalidArgument (Aliyun SMS is template-based).
- Provide provider requestId/messageId in response if available.

Config schema:
- access_key_id (secret? usually not but treat as secret)
- access_key_secret (secret)
- region (default cn-hangzhou)
- sign_name (default)
- default_template_code (optional)

## Tencent Cloud SMS
Use official TencentCloud SDK (sms/v20210111).
SendSms requires:
- SmsSdkAppId
- SignName
- TemplateId
- TemplateParamSet
- PhoneNumberSet

Config schema:
- secret_id (secret? treat as secret)
- secret_key (secret)
- sms_sdk_app_id
- sign_name
- region (default ap-guangzhou)
- default_template_id (optional)

