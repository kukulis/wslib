<?php

namespace App\WsHubUser;

use App\Handler\FileSystemDelegator;
use App\Handler\FileSystemDelegatorContract;
use App\Handler\FileUploadHandler;
use WsGatewayClient\Client;

class Container
{
    private ?Client $gatewayClient = null;

    private ?FileUploadHandler $fileUploadHandler = null;

    private ?FileSystemDelegatorContract $fileSystemDelegator = null;

    public function getGatewayClient(): Client
    {
        if ($this->gatewayClient === null) {
            // TODO use config
            $this->gatewayClient = new Client(new \GuzzleHttp\Client(), 'http://wshub:8585');
        }

        return $this->gatewayClient;
    }

    public function getFileUploadHandler(): FileUploadHandler
    {
        if ($this->fileUploadHandler === null) {
            $this->fileUploadHandler = new FileUploadHandler(
                $this->getGatewayClient(),
                $this->getFileSystemDelegator(),
                __DIR__ . '/../../upload'
            );
        }

        return $this->fileUploadHandler;
    }


    public function getFileSystemDelegator(): FileSystemDelegatorContract
    {
        if ($this->fileSystemDelegator === null) {
            $this->fileSystemDelegator = new FileSystemDelegator();
        }

        return $this->fileSystemDelegator;
    }
}