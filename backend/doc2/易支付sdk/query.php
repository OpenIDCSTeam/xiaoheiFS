<?php
require_once("config.php");

$info=getbody("https://mzf.jzmohe.com/api.php?act=order&pid=$pid&key=$key&out_trade_no=20230506030508","","GET");
die($info);
?>