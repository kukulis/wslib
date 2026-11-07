<?php

namespace App\Messages;

use WsGatewayClient\Dto\Envelope;

class CorrectUploadFileMessage
{
    public const string TYPE = 'correct_upload_file';
    public string $message= '';
    public string $fileName = '';

    public function setMessage(string $message): CorrectUploadFileMessage
    {
        $this->message = $message;
        return $this;
    }

    public function setFileName(string $fileName): CorrectUploadFileMessage
    {
        $this->fileName = $fileName;
        return $this;
    }

    public function toEnvelope(): Envelope
    {
        return (new Envelope())
            ->setType(self::TYPE)
            ->setData(get_object_vars($this));
    }
}
