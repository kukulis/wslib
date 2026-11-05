<?php

require_once __DIR__ . '/../vendor/autoload.php';

use App\WsHubUser\Container;

// The backend registers the session before it renders the page, and gives its
// id and token to the Web client, which logs in with them over the WebSocket.
// A new client id is needed on every load: the gateway rejects a repeated one.
// All the pages carry the same tag, so one message to the tag moves the object on every open page.
$clientId = bin2hex(random_bytes(16));
$sessionToken = bin2hex(random_bytes(16));
$tag = 'map';

try {
    (new Container())->getGatewayClient()->registerSession($clientId, $sessionToken, [$tag]);
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
    <title>Kaunas map (OSM tiles)</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css">
    <style>
        body { font-family: sans-serif; margin: 16px; }
        #map { width: 1000px; height: 667px; border: 1px solid #999; }
        #status { margin-top: 8px; font-size: 14px; }
    </style>
</head>
<body>
<h1>Kaunas map (OSM tiles)</h1>

<div id="map"></div>
<div id="status">connecting...</div>
<p>Start the movement for every open copy of this page in the wshub_user container:
    <code>php long_running/imitate_movement_registered.php <?= htmlspecialchars($tag) ?></code></p>

<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
<script>
    const CLIENT_ID = <?= json_encode($clientId) ?>;
    const SESSION_TOKEN = <?= json_encode($sessionToken) ?>;

    // the basemap is a static tile layer; the object is a client-side marker above it
    const map = L.map('map').setView([54.89, 23.956], 12);
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    }).addTo(map);

    const statusEl = document.getElementById('status');
    let poi = null;

    function movePoi(latitude, longitude) {
        if (poi === null) {
            poi = L.circleMarker([latitude, longitude], {
                radius: 8, color: '#fff', weight: 2, fillColor: 'red', fillOpacity: 1,
            }).addTo(map);
        } else {
            poi.setLatLng([latitude, longitude]);
        }
        statusEl.textContent = 'logged in; object at ' + latitude.toFixed(6) + ', ' + longitude.toFixed(6);
    }

    function startWsHandling(uri, clientId, sessionToken) {
        const protocol = document.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const connection = new WebSocket(protocol + document.location.host + uri);
        let loginFailed = false;

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
            if (loginFailed) {
                statusEl.textContent = 'login failed';
                return;
            }
            statusEl.textContent = 'disconnected; retrying...';
            setTimeout(() => startWsHandling(uri, clientId, sessionToken), 2000);
        };
        connection.onmessage = (evt) => {
            for (const line of evt.data.split('\n')) {
                if (line.trim() === '') {
                    continue;
                }
                const envelope = JSON.parse(line);
                if (envelope.type === 'login_result') {
                    if (envelope.data.success) {
                        statusEl.textContent = 'logged in; waiting for object_moved messages';
                    } else {
                        loginFailed = true;
                        connection.close();
                    }
                } else if (envelope.type === 'object_moved') {
                    movePoi(envelope.data.latitude, envelope.data.longitude);
                }
            }
        };
    }

    startWsHandling('/ws', CLIENT_ID, SESSION_TOKEN);
</script>
</body>
</html>
