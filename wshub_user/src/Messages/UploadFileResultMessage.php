<?php

namespace App\Messages;

class UploadFileResultMessage
{
    public const string TYPE = 'upload_file_result';

    public bool $success = false;
    public bool $cancelledByUser = false;
    /** the name the file was uploaded with */
    public string $uploadedFileName = '';
    /** the name the file was saved with; empty when it was not saved */
    public string $resultFileName = '';

    public function setSuccess(bool $success): UploadFileResultMessage
    {
        $this->success = $success;
        return $this;
    }

    public function setCancelledByUser(bool $cancelledByUser): UploadFileResultMessage
    {
        $this->cancelledByUser = $cancelledByUser;
        return $this;
    }

    public function setUploadedFileName(string $uploadedFileName): UploadFileResultMessage
    {
        $this->uploadedFileName = $uploadedFileName;
        return $this;
    }

    public function setResultFileName(string $resultFileName): UploadFileResultMessage
    {
        $this->resultFileName = $resultFileName;
        return $this;
    }
}
