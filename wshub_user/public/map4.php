<?php

require_once __DIR__ . '/../vendor/autoload.php';

use App\WsHubUser\Container;

// The backend registers the session before it renders the page, and gives its
// id and token to the Web client, which logs in with them over the WebSocket.
// A new client id is needed on every load: the gateway rejects a repeated one.
// All the pages carry the same tag, so one message to the tag moves the objects on every open page.
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
    <title>Kaunas map (moving objects)</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css">
    <style>
        body { font-family: sans-serif; margin: 16px; }
        #map { width: 1000px; height: 667px; border: 1px solid #999; }
        #status { margin-top: 8px; font-size: 14px; }
    </style>
</head>
<body>
<h1>Kaunas map (moving objects)</h1>

<div id="map"></div>
<div id="status">connecting...</div>
<p>Start the movement for every open copy of this page in the wshub_user container:
    <code>php long_running/imitate_movement_registered.php <?= htmlspecialchars($tag) ?></code></p>

<!--<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>-->
<script src="/assets/js/leaflet.js"></script>
<script type="module">
    import {MovingObject} from "/assets/js/moving-object.js";
    import {MovingObjectsMessage} from "/assets/js/moving-objects-message.js";
    // import L from "/assets/js/leaflet.js";

    const CLIENT_ID = <?= json_encode($clientId) ?>;
    const SESSION_TOKEN = <?= json_encode($sessionToken) ?>;

    // the basemap is a static tile layer; the objects are client-side markers above it
    const map = L.map('map').setView([54.89, 23.956], 12);
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    }).addTo(map);

    const statusEl = document.getElementById('status');
    // the marker of every moving object, by the object id
    const palette = {
        'aaa': 'red',
        'bbb': 'green',
        'ccc': 'blue',
        'ddd': 'yellow',
        'eee': 'orange',
        'fff': 'purple',
        'ggg': 'brown',
        'hhh': 'pink',
        'iii': 'gray',
        'jjj': 'black',
    }
    const markers = {};

    /**
     * @param {MovingObject[]} movingObjects
     */
    function showMovingObjects(movingObjects) {
        for (const movingObject of movingObjects) {
            const position = [movingObject.latitude, movingObject.longitude];

            let color = palette[movingObject.id];

            if (markers[movingObject.id] === undefined) {
                markers[movingObject.id] = L.circleMarker(position, {
                    radius: 8, color: '#fff', weight: 2, fillColor: color, fillOpacity: 1,
                }).addTo(map);
            } else {
                markers[movingObject.id].setLatLng(position);
            }
        }

        // what does 'logged in' mean?'
        statusEl.textContent = 'logged in; ' + Object.keys(markers).length + ' object(s) on the map';
    }

    /**
     *
     * @param {MovingObjectsMessage} movingObjectsMessage
     */
    function showMultipleMovingObjects(movingObjectsMessage) {
        for (const movingObject of movingObjectsMessage.objects) {
            const position = [movingObject.latitude, movingObject.longitude];
            let color = palette[movingObject.id];

            // console.log('map4.php[84] color:', color);

            if (markers[movingObject.id] === undefined) {
                markers[movingObject.id] = L.circleMarker(position, {
                    radius: 8, color: '#fff', weight: 2, fillColor: color, fillOpacity: 1,
                }).addTo(map);
            } else {
                markers[movingObject.id].setLatLng(position);
            }
        }

        statusEl.textContent = '' + Object.keys(markers).length + ' object(s) on the map';

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
                        // the token of this login is spent; a reconnect must use the new one
                        sessionToken = envelope.data.token;
                        statusEl.textContent = 'logged in; waiting for object_moved messages';
                    } else {
                        loginFailed = true;
                        connection.close();
                    }
                } else if (envelope.type === 'object_moved') {
                    // the message carries one object for now, the list has a single element
                    showMovingObjects([new MovingObject().fromDto(envelope.data)]);
                } else if (envelope.type === 'moving_objects') {
                    // the message carries one object for now, the list has a single element
                    // console.log('map4.php[147] envelope.data:', envelope.data);
                    showMultipleMovingObjects(new MovingObjectsMessage().fromDto(envelope.data));
                }
            }
        };
    }

    startWsHandling('/ws', CLIENT_ID, SESSION_TOKEN);
</script>
</body>
</html>
