<?php
require_once("config.php");
$info=getbody("https://mzf.jzmohe.com/api.php?act=refund","pid=$pid&key=$key&out_trade_no=20230506030508&money=0.01","POST");
die($info);
?>