<?php

include_once __DIR__ . '/../vendor/autoload.php';

use GuzzleHttp\Exception\GuzzleException;
use WsGatewayClient\Dto\Envelope;

echo "test wshub client\n";

$container = new \WsHubUser\Container();

try {
    $response = $container->getGatewayClient()->sendClientMessage(
        'abc',
        'testMessageId',
        (new Envelope())
            ->setType('testType')
            ->setData(['test' => 'test data'])
            ->setWaitResponse(3000)
    );

    echo "Response status = " . $response->getStatusCode() . "\n";
    echo "Response body = " . $response->getBody()->getContents() . "\n";
} catch (GuzzleException $e) {
    echo "Guzzle error:". $e->getMessage()."\n";
} catch (\Exception $e) {
    echo "General error:". $e->getMessage()."\n";
}