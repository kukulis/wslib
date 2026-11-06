<?php
require_once __DIR__ . '/../vendor/autoload.php';

use \App\WsHubUser\Container;

$clientId = bin2hex(random_bytes(16));
$sessionToken = bin2hex(random_bytes(16));

try {
    (new Container())->getGatewayClient()->registerSession($clientId, $sessionToken, []);
} catch (\Throwable $e) {
    http_response_code(502);
    echo 'Could not register the session: ' . htmlspecialchars($e->getMessage());
    exit;
}


?>
<html lang="en">
<head>
    <title>Upload</title>
</head>
<body>
TODO upload form / js uplaod form. with additional hidden parameter of user id
<form method="post" enctype="multipart/form-data" id="uploadForm" action="/api/upload.php">
    <input type="hidden" name="userId" value="<?php echo $clientId ?>">>

    <label for="fileToUpload">File to upload:</label>
    <input type="file" name="fileToUpload" id="fileToUpload">
    <button type="submit" id="uploadButton" value="Upload file" name="submit">
</form>


TODO js code that connects to WS and generates user GUI to send answers to the

<script type="module">
    import {GetE} from "/assets/js/util.js";

    const uploadForm = GetE('uploadForm');

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const res = await fetch(uploadForm.getAttribute('action').valueOf(), {method: 'POST', body: new FormData(uploadForm)});

        console.log('Upload response:', res);
    });

</script>
</body>
</html>



