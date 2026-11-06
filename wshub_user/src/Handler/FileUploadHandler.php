<?php

namespace App\Handler;

use WsGatewayClient\ClientContract;

class FileUploadHandler
{
    private string $uploadDir;

    private string $resultFileName;
    private bool $isUploadSuccess = false;
    private bool $userCancelledUpload = false;

    private string $fileName;
    private string $tmpFilePath;

    public function __construct(
        private readonly ClientContract $client,
        string $uploadDir,
    ) {
        $this->uploadDir = $uploadDir;
    }

    private function startUploadCycle(string $fileName, string $tmpFilePath) {
        $this->fileName = $fileName;
        $this->tmpFilePath = $tmpFilePath;
    }


    public function uploadFileCycle(): bool {
        // the single cycle function

        // TODO

        return $this->fileName;
    }

    public function isUploadSuccess(): bool
    {
        return $this->isUploadSuccess;
    }

    public function isUserCancelledUpload(): bool
    {
        return $this->userCancelledUpload;
    }

    public function getResultFileName(): string
    {
        return $this->resultFileName;
    }
}