<?php

namespace App\Messages;

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
}