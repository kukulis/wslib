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
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Upload</title>
    <style>
        body { font-family: sans-serif; margin: 16px; }
        #question { display: none; margin-top: 16px; padding: 12px; border: 1px solid #999; max-width: 480px; }
        #question input { width: 100%; box-sizing: border-box; margin: 8px 0; }
        #status, #result { margin-top: 12px; font-size: 14px; }
    </style>
</head>
<body>
<h1>Upload</h1>

<form method="post" enctype="multipart/form-data" id="uploadForm" action="/api/upload.php">
    <input type="hidden" name="userId" value="<?= htmlspecialchars($clientId) ?>">

    <label for="fileToUpload">File to upload:</label>
    <input type="file" name="fileToUpload" id="fileToUpload">
    <button type="submit" id="uploadButton" disabled>Upload file</button>
</form>

<div id="question">
    <div id="questionMessage"></div>
    <input type="text" id="correctedFileName">
    <button type="button" id="sendCorrectionButton">Upload with this name</button>
    <button type="button" id="cancelUploadButton">Cancel upload</button>
</div>

<div id="status">connecting...</div>
<div id="result"></div>

<script type="module">
    import {GetE} from "/assets/js/util.js";

    const CLIENT_ID = <?= json_encode($clientId) ?>;
    const SESSION_TOKEN = <?= json_encode($sessionToken) ?>;

    const uploadForm = GetE('uploadForm');
    const uploadButton = GetE('uploadButton');
    const questionEl = GetE('question');
    const questionMessageEl = GetE('questionMessage');
    const correctedFileNameEl = GetE('correctedFileName');
    const statusEl = GetE('status');
    const resultEl = GetE('result');

    let connection = null;
    // the id of the question the user is answering; the answer must carry the same id
    let questionId = null;

    // =========== upload ================
    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        uploadButton.disabled = true;
        resultEl.textContent = '';
        statusEl.textContent = 'uploading; waiting for the server';

        // The server answers this request at once and goes on processing the file. Its question
        // or result can reach the WebSocket before this request is resolved here, so only a failure
        // may change the status below: a success must not overwrite what the WebSocket already showed.
        try {
            const res = await fetch(uploadForm.getAttribute('action'), {method: 'POST', body: new FormData(uploadForm)});
            if (!res.ok) {
                statusEl.textContent = 'upload failed: HTTP ' + res.status;
                uploadButton.disabled = false;
            } else {
                // the answer is ["OK"] when the server accepted the file, otherwise [the reason]
                const answer = await res.json();
                if (answer[0] !== 'OK') {
                    // the server stops here, so nothing arrives over the WebSocket after a rejection
                    statusEl.textContent = 'upload rejected: ' + answer[0];
                    uploadButton.disabled = false;
                }
            }
        } catch (error) {
            statusEl.textContent = 'upload failed: ' + error.message;
            uploadButton.disabled = false;
        }
    });

    // =========== the answer to a question of the server ================
    function sendAnswer(correctedFileName, cancelUpload) {
        connection.send(JSON.stringify({
            id: questionId,
            type: 'upload_file_user_response',
            data: {correctedFileName: correctedFileName, cancelUpload: cancelUpload},
        }));

        questionEl.style.display = 'none';
        statusEl.textContent = 'answer sent; waiting for the server';
    }

    GetE('sendCorrectionButton').addEventListener('click', () => {
        sendAnswer(correctedFileNameEl.value, false);
    });

    GetE('cancelUploadButton').addEventListener('click', () => {
        sendAnswer('', true);
    });

    // =========== websocket handling ================
    function startWsHandling(uri, clientId, sessionToken) {
        const protocol = document.location.protocol === 'https:' ? 'wss://' : 'ws://';
        connection = new WebSocket(protocol + document.location.host + uri);

        // the first message of a connection must be the login
        connection.onopen = () => {
            statusEl.textContent = 'connected; logging in...';
            connection.send(JSON.stringify({
                id: 'login',
                type: 'login',
                data: {clientId: clientId, token: sessionToken},
            }));
        };
        connection.onclose = () => {
            statusEl.textContent = 'disconnected';
            uploadButton.disabled = true;
        };
        connection.onmessage = (evt) => {
            const envelope = JSON.parse(evt.data);

            if (envelope.type === 'login_result') {
                if (envelope.data.success) {
                    statusEl.textContent = 'logged in; choose a file';
                    // the server cannot ask anything before the login
                    uploadButton.disabled = false;
                } else {
                    statusEl.textContent = 'login failed';
                    connection.close();
                }
                return;
            }

            if (envelope.type === 'correct_upload_file') {
                questionId = envelope.id;
                questionMessageEl.textContent = envelope.data.message;
                correctedFileNameEl.value = envelope.data.fileName;
                questionEl.style.display = 'block';
                statusEl.textContent = 'the server asks for another file name';
                return;
            }

            if (envelope.type === 'upload_file_result') {
                const result = envelope.data;
                statusEl.textContent = 'done';
                if (result.success) {
                    resultEl.textContent = 'Uploaded as ' + result.resultFileName;
                } else if (result.cancelledByUser) {
                    resultEl.textContent = 'Upload of ' + result.uploadedFileName + ' was cancelled';
                } else {
                    resultEl.textContent = 'Upload of ' + result.uploadedFileName + ' failed';
                }
                uploadButton.disabled = false;
            }
        };
    }

    startWsHandling('/ws', CLIENT_ID, SESSION_TOKEN);
</script>
</body>
</html>
