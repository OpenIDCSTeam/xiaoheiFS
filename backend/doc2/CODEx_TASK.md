# CODEx TASK: Fix plugins using this pack

Read docs/00-START-HERE.md then implement in this order:
1) EZPay plugin strictly per docs/10-EZPAY.md
   - fixed method names: alipay/wxpay/qqpay
   - host-generated notify_url/return_url
   - correct signing + verify notify + ack 'success'
   - implement findorder query mapping

2) Add method-level enable/disable in host (NOT in plugin config)
   - host stores switches for alipay/wxpay/qqpay per plugin instance
   - payment options UI uses enabled methods only

3) Implement WeChatPay v3 plugin using official SDK (docs/20-WECHATPAY.md)
4) Implement Alipay plugin (docs/21-ALIPAY.md)
5) Replace SMS + KYC mocks with real SDK calls (docs/30-SMS.md, docs/40-KYC.md)

Non-negotiable:
- out_trade_no must be host order no
- no mock
- unit tests for signature verification
