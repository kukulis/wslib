<?php

namespace WsGatewayClient;

use GuzzleHttp\Client as GuzzleClient;
use Psr\Http\Message\ResponseInterface;
use WsGatewayClient\Dto\Envelope;

class Client
{
    public function __construct(
        private GuzzleClient $client,
        private string $url,
    ) {
    }

    /**
     * @param string $clientId receiving client id
     * @param Envelope $envelope is the message to send
     */
    public function sendClientMessage( string $clientId, Envelope $envelope) : ResponseInterface {
        return $this->client->post($this->url, [
            'json' => $envelope,
            'query' => ['clientId' => $clientId],
        ]);
    }
}