<?php

namespace App\Handler;

use App\Messages\CorrectUploadFileMessage;
use App\Messages\UploadFileResultMessage;
use App\Messages\UploadFileUserResponseMessage;
use WsGatewayClient\ClientContract;
use WsGatewayClient\Dto\SendMessageResponse;

class FileUploadHandler
{

    const string FILE_NAME_PATTERN = '/^[a-zA-Z0-9_\-\.]+$/';

    /** how long the user is given to answer, milliseconds */
    const int USER_RESPONSE_WAIT_MS = 60000;

    private string $uploadDir;

    private string $resultFileName = '';
    private bool $isUploadSuccess = false;
    private bool $userCancelledUpload = false;

    private string $clientId;
    private string $uploadedFileName;
    private string $fileName;
    private string $fileContent;

    public function __construct(
        private readonly ClientContract $client,
        private readonly FileSystemDelegatorContract $fileSystemDelegator,
        string $uploadDir,
    ) {
        $this->uploadDir = $uploadDir;
    }

    /**
     * @param string $fileName the name the file was uploaded with
     * @param string $fileContent read before the cycle starts, because the system may delete the tmp file if the cycle goes too long
     * @param string $clientId id of the registered client that gets the messages
     */
    public function startUploadCycle(string $fileName, string $fileContent, string $clientId): void
    {
        $this->uploadedFileName = $fileName;
        $this->fileName = $fileName;
        $this->fileContent = $fileContent;
        $this->clientId = $clientId;
    }

    /**
     * One step of the upload. Call it while it returns true.
     */
    public function uploadFileCycleStep(\Closure $errorLog): bool
    {
        $this->isUploadSuccess = false;
        $message = '';

        if (!$this->validateFileName($this->fileName)) {
            $message = 'The file name "' . $this->fileName . '" is not allowed. Use only letters, digits, "_", "-" and ".".';
        }

        $filePath = rtrim($this->uploadDir, '/') . '/' . $this->fileName;

        if ($message == '') {
            if ($this->fileSystemDelegator->filePathExists($filePath)) {
                $message = 'The file "' . $this->fileName . '" already exists. Give another name.';
            }
        }

        if ($message !== '') {
            $userResponse = $this->askUser($message, self::USER_RESPONSE_WAIT_MS, $errorLog);

            if ($userResponse->responseCode != 200) {
                // user answer Timeout or other error
                // finish the cycle
                return false;
            }

            $userMessage = UploadFileUserResponseMessage::fromArray((array)$userResponse->wsMessage?->data);

            if ($userMessage->cancelUpload) {
                // user cancel
                $this->userCancelledUpload = true;
                // finish the cycle
                return false;
            }

            $this->fileName = $userMessage->correctedFileName;

            // continue the cycle
            return true;
        }


        $this->isUploadSuccess = $this->fileSystemDelegator->writeFile(
            $filePath,
            $this->fileContent
        );

        if ($this->isUploadSuccess) {
            $this->resultFileName = $this->fileName;
        }

        // finish the cycle
        return false;
    }

    /**
     * Tells the user how the upload ended. Call it after uploadFileCycleStep() returned false.
     * A failure to reach the gateway is thrown.
     */
    public function sendUploadResult(): void
    {
        $result = (new UploadFileResultMessage())
            ->setSuccess($this->isUploadSuccess)
            ->setCancelledByUser($this->userCancelledUpload)
            ->setUploadedFileName($this->uploadedFileName)
            ->setResultFileName($this->resultFileName);

        $this->client->sendRegisteredClientMessage($this->clientId, '', $result->toEnvelope());
    }

    private function askUser(string $message, int $waitMs, \Closure $errorLog): SendMessageResponse
    {
        $question = (new CorrectUploadFileMessage())
            ->setMessage($message)
            ->setFileName($this->fileName);

        // the message id must be unique: the gateway matches the answer to the waiter by it
        $userResponse = $this->client->sendRegisteredClientMessage(
            $this->clientId,
            bin2hex(random_bytes(8)),
            $question->toEnvelope()->setWaitResponse($waitMs)
        );

        if ($userResponse->responseCode != 200) {
            $errorLog(
                sprintf("User response code %d, message %s ", $userResponse->responseCode, $userResponse->errorMessage)
            );
        }

        return $userResponse;
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

    public function validateFileName(string $fileName): bool
    {
        $rez = preg_match(self::FILE_NAME_PATTERN, $fileName, $matches);

        return $rez === 1;
    }

    public function run(string $file_name, string $file_tmp_path, string $userId, \Closure $errorLog): void
    {
        $this->startUploadCycle($file_name, file_get_contents($file_tmp_path), $userId);
        $count = 0;
        while ($this->uploadFileCycleStep($errorLog)) {
            $count++;

            $errorLog("Cycle $count for uploading " . $file_name);
        }

        try {
            $this->sendUploadResult();
        } catch (\Throwable $e) {
            $errorLog("Could not send the upload result to $userId: " . $e->getMessage());
        }
    }

}