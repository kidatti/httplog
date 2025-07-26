<?php
header("Access-Control-Allow-Origin: *");

// ログディレクトリ作成
$dir = 'logs/' . date("Ymd-His");
$num = 1;
do {
    $dir_log = $dir . '-' . sprintf("%03d", $num);
    if (!file_exists($dir_log)) {
        mkdir($dir_log, 0777, true);
        break;
    }
    $num++;
} while (true);

// ログ出力
$log_file = $dir_log . '/log.txt';
$data_file = $dir_log . '/data.txt';

$data = file_get_contents("php://input");
$request = [
    'method' => $_SERVER["REQUEST_METHOD"],
    'url' => $_SERVER['REQUEST_URI'],
    'header' => [],
    'get' => [],
    'data' => $data,
];
// header
foreach (getallheaders() as $name => $value) {
    $request['header'][$name] = $value;
}
// get
foreach ($_GET as $key => $value) {
    $request['get'][$key] = $value;
}
// post
foreach ($_POST as $key => $value) {
    $request['post'][$key] = $value;
}
// post
foreach ($_FILES as $key => $value) {
    $request['files'][$key] = $value;
    move_uploaded_file($value['tmp_name'], $dir_log . '/' . $value['name']);
}
file_put_contents($log_file, print_r($request,true));
if (isset($data) AND $data != "") {
    file_put_contents($data_file, $data);
}