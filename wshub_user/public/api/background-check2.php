<?php

use App\WsHubUser\Container;
use WsGatewayClient\Dto\Envelope;

require_once __DIR__ . '/../../vendor/autoload.php';

function errorLog($message)
{
    error_log($message, 3, '/tmp/background-check2.log');
}

ignore_user_abort(true);

$clientId = $_POST['clientId'];

$randomWait = rand(1, 20);

$response = [
    'message' => "Background check started for client $clientId ($randomWait seconds)",
];

header('Content-Type: application/json');
echo json_encode($response);

if (function_exists('fastcgi_finish_request')) {
    fastcgi_finish_request();
}


errorLog("$clientId: Sleeping for $randomWait seconds...\n");
sleep($randomWait);
errorLog("$clientId: Finished sleeping\n");


try {
    $container = new Container();
    errorLog("$clientId: Sending message to WS Gateway\n");

    $container->getGatewayClient()->sendRegisteredClientMessage(
        $clientId,
        '',
        (new Envelope())
            ->setType('background_check_result')
            ->setData(['result' => 'approved'])
    );
    errorLog("$clientId: End\n");
} catch ( \Throwable $e) {
    errorLog("$clientId: Error: " . $e->getMessage() . "\n");
}
