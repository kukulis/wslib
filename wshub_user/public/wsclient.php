<?php
/**
 * The js client code
 */
?>

<html>
<head>
    <title>WsClient</title>
</head>
<body>
Ws client

<div id="messages"> </div>

<script type="text/javascript">


    function startWsHandling(uri, clientId) {
        let protocol = 'ws://';

        if (this.isTls) {
            protocol = 'wss://';
        }

        if (!window["WebSocket"]) {
            alert('Your browser does not support WebSockets.');
            return;
        }

        let connection = new WebSocket(protocol + document.location.host + uri+'?clientId='+clientId);
        // send our coordinates right away instead of waiting for the server to ask;
        // only once open, since send() throws while the socket is still connecting
        connection.onopen = () => {
            console.log('connected ws')
        };
        connection.onclose = (evt) => {
            console.log('disconnected ws')
        };
        connection.onmessage = (evt) => {
            let messages = evt.data.split('\n'); // TODO not sure if this is needed
            for (let i = 0; i < messages.length; i++) {
                const envelope = JSON.parse(messages[i]);

                console.log('received message', envelope.type );

                let p = document.createElement('p');

                let displayedTex = envelope.type + ': ' +  JSON.stringify( envelope.data );

                p.appendChild( document.createTextNode(displayedTex) );
                document.getElementById("messages").appendChild(p);
            }
        };
    }

    startWsHandling('/ws', 'abc')
</script>
</body>
</html>
