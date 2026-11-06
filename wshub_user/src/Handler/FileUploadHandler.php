<?php

namespace App\Handler;

use App\Messages\CorrectUploadFileMessage;
use App\Messages\UploadFileResultMessage;
use App\Messages\UploadFileUserResponseMessage;
use WsGatewayClient\ClientContract;
use WsGatewayClient\Dto\Envelope;

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
    public function uploadFileCycle(): bool
    {
        if (!$this->validateFileName($this->fileName)) {
            return $this->askForAnotherFileName(
                'The file name "' . $this->fileName . '" is not allowed. Use only letters, digits, "_", "-" and ".".'
            );
        }

        $filePath = rtrim($this->uploadDir, '/') . '/' . $this->fileName;

        if ($this->fileSystemDelegator->filePathExists($filePath)) {
            return $this->askForAnotherFileName(
                'The file "' . $this->fileName . '" already exists. Give another name.'
            );
        }

        $this->isUploadSuccess = $this->fileSystemDelegator->writeFile(
            $filePath,
            $this->fileContent
        );

        if ($this->isUploadSuccess) {
            $this->resultFileName = $this->fileName;
        }

        return false;
    }

    /**
     * Tells the user how the upload ended. Call it after uploadFileCycle() returned false.
     * A failure to reach the gateway is thrown.
     */
    public function sendUploadResult(): void
    {
        $result = (new UploadFileResultMessage())
            ->setSuccess($this->isUploadSuccess)
            ->setCancelledByUser($this->userCancelledUpload)
            ->setUploadedFileName($this->uploadedFileName)
            ->setResultFileName($this->resultFileName);

        $this->client->sendRegisteredClientMessage(
            $this->clientId,
            '',
            (new Envelope())
                ->setType(UploadFileResultMessage::TYPE)
                ->setData(get_object_vars($result))
        );
    }

    /**
     * Asks the user to correct the file name and waits for the answer.
     *
     * @return bool true when the user gave a new name (the cycle must run again), false when the upload
     *              must stop: the user cancelled it or did not answer
     */
    private function askForAnotherFileName(string $message): bool
    {
        $answer = $this->askUser($message);

        if ($answer === null) {
            return false;
        }

        if ($answer->cancelUpload) {
            $this->userCancelledUpload = true;
            return false;
        }

        $this->fileName = $answer->correctedFileName;

        return true;
    }

    /**
     * @return UploadFileUserResponseMessage|null null when the user did not answer in time or the message
     *                                            could not be delivered
     */
    private function askUser(string $message): ?UploadFileUserResponseMessage
    {
        $question = (new CorrectUploadFileMessage())
            ->setMessage($message)
            ->setFileName($this->fileName);

        try {
            // the message id must be unique: the gateway matches the answer to the waiter by it
            $response = $this->client->sendRegisteredClientMessage(
                $this->clientId,
                bin2hex(random_bytes(8)),
                (new Envelope())
                    ->setType(CorrectUploadFileMessage::TYPE)
                    ->setData(get_object_vars($question))
                    ->setWaitResponse(self::USER_RESPONSE_WAIT_MS)
            );
        } catch (\Throwable) {
            // the gateway could not be reached
            return null;
        }

        // no reply message means the gateway answered with an error: the client is gone or did not reply in time
        if ($response->wsMessage === null) {
            return null;
        }

        return (new UploadFileUserResponseMessage())
            ->setCorrectedFileName((string)($response->wsMessage['correctedFileName'] ?? ''))
            ->setCancelUpload((bool)($response->wsMessage['cancelUpload'] ?? false));
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
        return preg_match(self::FILE_NAME_PATTERN, $fileName, $matches);
    }
}