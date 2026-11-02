<?php

namespace App\WsHubUser;

use WsGatewayClient\Client;

class Container
{
    private ?Client $gatewayClient = null;
    public function getGatewayClient() : Client {
        if ($this->gatewayClient === null) {
            // TODO use config
            $this->gatewayClient = new Client(new \GuzzleHttp\Client(), 'http://wshub:8585');
        }

        return $this->gatewayClient;
    }
}