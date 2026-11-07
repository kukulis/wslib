<?php

require_once __DIR__ . '/../../vendor/autoload.php';

use  \App\WsHubUser\Container;

function errorLog($message): void
{
    error_log($message . "\n", 3, '/tmp/upload.log');
}


$file = $_FILES['fileToUpload'];

$file_name = $file['name'];
$file_tmp = $file['tmp_name'];

$userId = $_POST['userId'];

$file_size = $file['size'];

if ( $file_size > 10000000 ) {

    $message = "File $file_name is too big";
    $response = [$message];

    echo json_encode($response);
    exit;
}

// check if the file upload to tmp was successful
if (!file_exists($file_tmp)) {
    $message = "Something wrong with the file upload.";
    $response = [$message];

    echo json_encode($response);
    exit;
}


$response = ["OK"];

echo json_encode($response);

// end the connection
if (function_exists('fastcgi_finish_request')) {
    fastcgi_finish_request();
}

// =================================================================================
// ============================== continuing as an independent process =============
// =================================================================================

$container = new Container();

$handler = $container->getFileUploadHandler();

$handler->run($file_name, $file_tmp, $userId, fn($msg)=>errorLog($msg));

errorLog(
    sprintf(
        "Finished uploading %s.
 Result file: %s.
 Success: %s.
 Cancelled: %s.",
        $file_name,
        $handler->getResultFileName(),
        $handler->isUploadSuccess(),
        $handler->isUserCancelledUpload()
    )
);
