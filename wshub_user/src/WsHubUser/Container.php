<?php

namespace App\WsHubUser;

use App\Handler\FileUploadHandler;
use WsGatewayClient\Client;

class Container
{
    private ?Client $gatewayClient = null;

    private ?FileUploadHandler $fileUploadHandler = null;
    public function getGatewayClient() : Client {
        if ($this->gatewayClient === null) {
            // TODO use config
            $this->gatewayClient = new Client(new \GuzzleHttp\Client(), 'http://wshub:8585');
        }

        return $this->gatewayClient;
    }

    public function getFileUploadHandler() : FileUploadHandler {
        if ($this->fileUploadHandler === null) {
            $this->fileUploadHandler = new FileUploadHandler(
                $this->getGatewayClient(),
                __DIR__ . '/../../upload'
            );
        }

        return $this->fileUploadHandler;
    }
}