<?php

namespace App\Messages;

class UploadFileUserResponseMessage
{
    public const string TYPE = 'upload_file_user_response';

    public string $correctedFileName = '';
    public bool $cancelUpload = false;

    public function setCorrectedFileName(string $correctedFileName): UploadFileUserResponseMessage
    {
        $this->correctedFileName = $correctedFileName;
        return $this;
    }

    public function setCancelUpload(bool $cancelUpload): UploadFileUserResponseMessage
    {
        $this->cancelUpload = $cancelUpload;
        return $this;
    }
}