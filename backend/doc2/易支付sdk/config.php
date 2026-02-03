<?php
$apiurl = 'https://mzf.jzmohe.com/';
$pid = '20220726190052';
$key = 'vg9ZRZN4FOKtDM06UfqH69GDJoG4gGIJ';

function GetBody($url, $xml,$method='POST'){    
    $second = 30;
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_TIMEOUT, $second);
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER,false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST,false);
    curl_setopt($ch, CURLOPT_HEADER, false);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, TRUE);
    curl_setopt($ch, CURLOPT_POST, TRUE);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $xml);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
    $data = curl_exec($ch);
    if($data){
      curl_close($ch);
      return $data;
    } else { 
      $error = curl_errno($ch);
      curl_close($ch);
      return false;
    }
}
function get_sign(array $datas,$hashkey){
    ksort($datas);
    reset($datas);
     
    $pre =array();
    foreach ($datas as $key => $data){
        if(is_null($data)||$data===''){continue;}
        if($key=='sign' || $key=='sign_type'){
            continue;
        }
        $pre[$key]=stripslashes($data);
    }
     
    $arg  = '';
    $qty = count($pre);
    $index=0;
     
    foreach ($pre as $key=>$val){
        $arg.="$key=$val";
        if($index++<($qty-1)){
            $arg.="&";
        }
    }
    
    return strtolower(md5($arg.$hashkey));
}
?>