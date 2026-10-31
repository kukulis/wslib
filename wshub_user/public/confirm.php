<?php

$clientId = uniqid();
?>
<html lang="en">
<head>
    <title>Confirm from Web Socket example</title>
    <style>
        .spinner {
            display: inline-block;
            width: 1em;
            height: 1em;
            border: 2px solid #ccc;
            border-top-color: #333;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
            vertical-align: middle;
            margin-right: 0.5em;
        }

        @keyframes spin {
            to {
                transform: rotate(360deg);
            }
        }
    </style>
</head>
<body>
<h1>Confirm</h1>
<p>Your client id is: <?php
    echo $clientId; ?></p>
<button id="button-background-check">Start backend check</button>
<p class="elapsed-counter" id="timer-counter"></p>
<div id="check-result"></div>

<script type="module">

    import {ClearE} from "../assets/js/util.js";
    import {GetE, NewEC, NewECT, NewT} from "../assets/js/util.js";

    let clientId = '<?php echo $clientId; ?>';

    const isTls = document.location.protocol === 'https:';

    const checkResultEl = GetE('check-result');
    const CHECK_TIMEOUT_MS = 10000;
    // global variable for timer
    let timeoutId = null;
    let counterIntervalId = null;

    const buttonBackgroundCheck = GetE('button-background-check');
    buttonBackgroundCheck.addEventListener('click', () => {
        startBackgroundCheck(clientId)
    })

    startWsHandling('/ws-old', clientId, isTls);

    let timeouted = false;

    // ====================================================================================================
    // ============== functions declarations ==============================================================
    // ====================================================================================================

    // =========== button listener ================
    async function startBackgroundCheck(clientId) {
        try {
            let data = await apiBackendCheck(clientId);

            const startTime = Date.now();

            timeoutId = setTimeout(() => {
                stopWaiting();
                checkResultEl.textContent = 'Timed out waiting for the background check result.';
                timeouted = true;
            }, CHECK_TIMEOUT_MS);

            ClearE(checkResultEl);
            checkResultEl.appendChild(NewECT('p', '', 'Background check: ' + JSON.stringify(data)));
            checkResultEl.appendChild(NewEC('span', 'spinner'));
            checkResultEl.appendChild(NewECT('p', '', 'Waiting for background check result (' + (CHECK_TIMEOUT_MS / 1000) + 's )...'));

            const counterEl = GetE('timer-counter')
            counterEl.textContent = 'Elapsed: 0s';

            counterIntervalId = setInterval(() => {
                const elapsedSeconds = Math.floor((Date.now() - startTime) / 1000);
                counterEl.textContent = 'Elapsed: ' + elapsedSeconds + 's';
            }, 1000);

        } catch (e) {
            console.error(e);
        }
    }

    // clears whichever of the timeout/counter are currently running
    function stopWaiting() {
        if (timeoutId !== null) {
            clearTimeout(timeoutId);
            timeoutId = null;
        }
        if (counterIntervalId !== null) {
            clearInterval(counterIntervalId);
            counterIntervalId = null;
        }
    }

    // =========== api call ===========================
    async function apiBackendCheck(clientId) {
        const formData = new FormData();
        formData.append('clientId', clientId);

        let response = await fetch('/api/background-check.php', {method: 'POST', body: formData})

        if (!response.ok) {
            throw new Error('Network response was not ok');
        }

        return await response.json();
    }

    // =========== websocket handling ================
    function startWsHandling(uri, clientId, isTls) {
        let protocol = 'ws://';

        if (isTls) {
            protocol = 'wss://';
        }

        if (!window["WebSocket"]) {
            alert('Your browser does not support WebSockets.');
            return;
        }

        let connection = new WebSocket(protocol + document.location.host + uri + '?clientId=' + clientId);

        connection.onopen = () => {
            console.log('connected ws');
        };
        connection.onclose = () => {
            console.log('disconnected ws');
        };
        connection.onmessage = (evt) => {
            let message = evt.data;
            const envelope = JSON.parse(message);

            console.log('received message', envelope.type);

            if (envelope.type === 'background_check_result') {

                stopWaiting();

                if (!timeouted) {
                    ClearE(checkResultEl);
                }
                checkResultEl.appendChild(NewECT('p', '', 'Background check result: ' + JSON.stringify(envelope.data)));
            }
        };
    }
</script>
</body>
</html>

