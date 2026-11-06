<?php

namespace App\Handler;

use App\Messages\CorrectUploadFileMessage;
use App\Messages\UploadFileUserResponseMessage;
use WsGatewayClient\ClientContract;
use WsGatewayClient\Dto\Envelope;

class FileUploadHandler
{

    const string FILE_NAME_PATTERN = '/^[a-zA-Z0-9_\-\.]+$/';

    const string RESULT_MESSAGE_TYPE = 'upload_file_result';

    /** how long the user is given to answer, milliseconds */
    const int USER_RESPONSE_WAIT_MS = 60000;

    private string $uploadDir;

    private string $resultFileName = '';
    private bool $isUploadSuccess = false;
    private bool $userCancelledUpload = false;

    private string $clientId;
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
        $this->fileName = $fileName;
        $this->fileContent = $fileContent;
        $this->clientId = $clientId;
    }

    /**
     * One step of the upload. Call it while it returns true.
     *
     * @return bool true when the user gave another file name and the cycle must run again;
     *              false when the upload is finished: it succeeded, failed, or the user cancelled it
     *              (see isUploadSuccess() and isUserCancelledUpload())
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

        $this->sendResult();

        return false;
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
            // includes the gateway timeout (504) and an unknown client (404)
            return null;
        }

        $body = json_decode((string)$response->getBody(), true);
        if (!is_array($body)) {
            return null;
        }
        // the answer fields may come at the top level or inside "data"
        $data = is_array($body['data'] ?? null) ? $body['data'] : $body;

        return (new UploadFileUserResponseMessage())
            ->setCorrectedFileName((string)($data['correctedFileName'] ?? ''))
            ->setCancelUpload((bool)($data['cancelUpload'] ?? false));
    }

    private function sendResult(): void
    {
        try {
            $this->client->sendRegisteredClientMessage(
                $this->clientId,
                '',
                (new Envelope())
                    ->setType(self::RESULT_MESSAGE_TYPE)
                    ->setData([
                        'success' => $this->isUploadSuccess,
                        'fileName' => $this->fileName,
                    ])
            );
        } catch (\Throwable) {
            // the file is already written or not; failing to tell the user does not change that
        }
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