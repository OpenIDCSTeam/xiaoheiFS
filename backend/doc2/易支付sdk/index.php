<?php
require_once("config.php");

$type="alipay";//支持设置支付宝：alipay 微信支付：wxpay
$notify_url="http://www.baidu.com/epay/notify.php";//异步回调地址
$return_url="http://www.baidu.com/epay/return.php";//同步回调地址
$out_trade_no=date("Ymdhms");//商户端订单号，不可重复
$name="测试商品";//商品名称
$money="0.01";//订单金额，最多保留2位小数
$sign_type="MD5";//签名方式，目前仅支持MD5
$param="附加信息";//附加信息，回调时原样返回

$arr = array(
    "pid" => $pid,
    "type" => $type,
    "notify_url" => $notify_url,
    "return_url" => $return_url,
    "out_trade_no" => $out_trade_no,
    "name" => $name,
    "money" => $money,
    "param" => $param,
    "sign_type" => $sign_type
);

$sign=get_sign($arr,$key);

header("location:".$apiurl."submit.php?pid=$pid&type=$type&notify_url=$notify_url&return_url=$return_url&out_trade_no=$out_trade_no&name=$name&money=$money&param=$param&sign_type=$sign_type&sign=$sign");
?>