<?php
require_once("config.php");

$money=$_GET["money"];
$name=$_GET["name"];
$pid=$_GET["pid"];
$out_trade_no=$_GET["out_trade_no"];
$trade_no=$_GET["trade_no"];
$trade_status=$_GET["trade_status"];
$type=$_GET["type"];
$param=$_GET["param"];
$trade_status=$_GET["trade_status"];
$sign=$_GET["sign"];
$sign_type=$_GET["sign_type"];

$arr=array(
	"pid" => $pid,
	"type" => $type,
	"out_trade_no" => $out_trade_no,
	"trade_no" => $trade_no,
	"name" => $name,
	"money"	=> $money,
	"param"	=> $param,
	"trade_status"	=> $trade_status,
	"sign_type"	=> $sign_type
);

if($sign==get_sign($arr,$key)){
	echo "success"; //返回success说明通知成功，不要删除本行
	//
	//
	//
	//
	//验证成功，在这里处理您的网站逻辑
	//
	//
	//
	//
	//
}else{
	echo "error";
}

?>