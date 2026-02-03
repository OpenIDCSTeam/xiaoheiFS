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
$sign_type=$_GET["sign_type"];

$arr=array(
	"pid" => $pid,
	"type" => $type,
	"notify_url" => $notify_url,
	"out_trade_no" => $out_trade_no,
	"trade_no" => $trade_no,
	"name" => $name,
	"money"	=> $money,
	"param"	=> $param,
	"trade_status"	=> $trade_status,
	"sign_type"	=> $sign_type
);
?>
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>同步回调</title> 
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="https://cdn.staticfile.org/twitter-bootstrap/3.3.7/css/bootstrap.min.css">  
	<script src="https://cdn.staticfile.org/jquery/2.1.1/jquery.min.js"></script>
	<script src="https://cdn.staticfile.org/twitter-bootstrap/3.3.7/js/bootstrap.min.js"></script>
</head>
<body>
	
<div class="container">
	<?php
		if($sign==get_sign($arr,$key)){
			die("<h1>验证成功！</h1>");
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
			die("<h1>验证失败！</h1>");
		}
	?>
</div>
	
</body>
</html>