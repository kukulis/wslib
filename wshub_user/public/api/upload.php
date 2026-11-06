<?php

function errorLog($message): void
{
    error_log($message, 3, '/tmp/upload.log');
}


$file = $_FILES['fileToUpload'];

$file_name = $file['name'];
$file_tmp = $file['tmp_name'];

$userId = $_POST['userId'];

// TODO allow only small files

// TODO check if the file upload to tmp was successful


$response = ["OK"];

echo json_encode($response);

// end the connection
if (function_exists('fastcgi_finish_request')) {
    fastcgi_finish_request();
}


$container = new \App\WsHubUser\Container();

$handler = $container->getFileUploadHandler();

$count = 0;
while ($handler->uploadFileCycle()) {
    $count++;

    errorLog("Cycle $count for uploading " . $file_name);
}

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
