<?php

namespace Tests\Handler;

use App\Handler\FileUploadHandler;
use App\Messages\CorrectUploadFileMessage;
use App\Messages\UploadFileResultMessage;
use PHPUnit\Framework\TestCase;
use Tests\Mock\ClientMock;
use Tests\Mock\FileSystemDelegatorMock;
use WsGatewayClient\Dto\Envelope;
use WsGatewayClient\Dto\SendMessageResponse;

class FileUploadHandlerTest extends TestCase
{
    private ClientMock $clientMock;
    private FileSystemDelegatorMock $fileSystemMock;
    /** @var array<int, array{clientId: string, envelope: Envelope}> */
    private array $sentMessages = [];
    /** @var array<string, string> path => content */
    private array $writtenFiles = [];

    protected function setUp(): void
    {
        $this->sentMessages = [];
        $this->writtenFiles = [];

        $this->clientMock = (new ClientMock())->setSendRegisteredClientMessageClosure(
            function (string $clientId, string $messageId, Envelope $envelope): SendMessageResponse {
                $this->sentMessages[] = ['clientId' => $clientId, 'envelope' => $envelope];
                return new SendMessageResponse(200, null, null);
            }
        );
        $this->fileSystemMock = (new FileSystemDelegatorMock())->setWriteFileClosure(
            function (string $path, string $content): bool {
                $this->writtenFiles[$path] = $content;
                return true;
            }
        );
    }

    private function createFileUploadHandler(): FileUploadHandler
    {
        return new FileUploadHandler($this->clientMock, $this->fileSystemMock, '/upload');
    }

    /**
     * @param array $answerData the payload of the user's answer; the closure wraps it in an envelope like the gateway does
     */
    private function createWsClientAnswerClosure(array $answerData): \Closure
    {
        $answer = (new Envelope())->setType('upload_file_user_response')->setData($answerData);

        return function (string $clientId, string $messageId, Envelope $envelope) use ($answer): SendMessageResponse {
            $this->sentMessages[] = ['clientId' => $clientId, 'envelope' => $envelope];
            return new SendMessageResponse(200, null, $envelope->waitResponse > 0 ? $answer : null);
        };
    }

    public function testValidNameIsWrittenAndResultSent(): void
    {
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');
        $this->assertSame(['/upload/a.txt' => 'content'], $this->writtenFiles);
        $this->assertSame('a.txt', $handler->getResultFileName());
        $this->assertCount(0, $this->sentMessages, 'the handler asks the user only; the result is reported by the caller');
    }

    public function testWriteFailureIsReported(): void
    {
        $this->fileSystemMock->setWriteFileClosure(fn() => false);
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload must have failed');
        $this->assertSame('', $handler->getResultFileName());
    }

    public function testWrongNameIsCorrectedByUser(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure(['correctedFileName' => 'good.txt', 'cancelUpload' => false])
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->assertTrue($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must not be finished yet');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload cannot be successful before the cycle is finished');
        $this->assertSame([], $this->writtenFiles);
        $this->assertSame(CorrectUploadFileMessage::TYPE, $this->sentMessages[0]['envelope']->type);
        $this->assertSame('client1', $this->sentMessages[0]['clientId']);

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');
        $this->assertSame(['/upload/good.txt' => 'content'], $this->writtenFiles);
        $this->assertSame('good.txt', $handler->getResultFileName());
    }

    public function testExistingFileAsksForAnotherName(): void
    {
        $this->fileSystemMock->setFilePathExistsClosure(fn(string $path) => $path === '/upload/a.txt');
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure(['correctedFileName' => 'b.txt'])
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertTrue($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must not be finished yet');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload cannot be successful before the cycle is finished');
        $this->assertSame([], $this->writtenFiles);

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');
        $this->assertSame(['/upload/b.txt' => 'content'], $this->writtenFiles);
    }

    public function testUserCancelStopsTheCycle(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure(['cancelUpload' => true])
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload must have failed');
        $this->assertTrue($handler->isUserCancelledUpload());
        $this->assertSame([], $this->writtenFiles);
    }

    public function testNoAnswerStopsTheCycle(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            fn() => new SendMessageResponse(504, 'no response from client within the given wait time', null)
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload must have failed');
        $this->assertFalse($handler->isUserCancelledUpload());
    }

    public function testUnreachableGatewayCrashesTheCycle(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            fn() => throw new \RuntimeException('connection refused')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->expectException(\RuntimeException::class);

        $handler->uploadFileCycleStep(fn($msg)=> $msg);
    }

    public function testSuccessResultIsSent(): void
    {
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');
        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');

        $handler->sendUploadResult();

        $this->assertCount(1, $this->sentMessages);
        $this->assertSame('client1', $this->sentMessages[0]['clientId']);
        $this->assertSame(UploadFileResultMessage::TYPE, $this->sentMessages[0]['envelope']->type);
        $this->assertSame(
            ['success' => true, 'cancelledByUser' => false, 'uploadedFileName' => 'a.txt', 'resultFileName' => 'a.txt'],
            $this->sentMessages[0]['envelope']->data
        );
    }

    public function testResultKeepsTheUploadedNameAfterCorrection(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure(['correctedFileName' => 'good.txt', 'cancelUpload' => false])
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');
        $this->assertTrue($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must not be finished yet');
        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');

        $handler->sendUploadResult();

        $result = $this->sentMessages[1]['envelope'];
        $this->assertSame(UploadFileResultMessage::TYPE, $result->type, 'the question is the first message, the result the second');
        $this->assertSame('bad name.txt', $result->data['uploadedFileName']);
        $this->assertSame('good.txt', $result->data['resultFileName']);
    }

    public function testCancelledResultIsSent(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure(['cancelUpload' => true])
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');
        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload must have failed');

        $handler->sendUploadResult();

        $this->assertSame(
            ['success' => false, 'cancelledByUser' => true, 'uploadedFileName' => 'bad name.txt', 'resultFileName' => ''],
            $this->sentMessages[1]['envelope']->data
        );
    }

    public function testWriteFailureResultIsSent(): void
    {
        $this->fileSystemMock->setWriteFileClosure(fn() => false);
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');
        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertFalse($handler->isUploadSuccess(), 'the upload must have failed');

        $handler->sendUploadResult();

        $this->assertSame(
            ['success' => false, 'cancelledByUser' => false, 'uploadedFileName' => 'a.txt', 'resultFileName' => ''],
            $this->sentMessages[0]['envelope']->data
        );
    }

    public function testUnreachableGatewayThrowsWhenSendingTheResult(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            fn() => throw new \RuntimeException('connection refused')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');
        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');

        $this->expectException(\RuntimeException::class);

        $handler->sendUploadResult();
    }

    public function testUploadDirWithTrailingSlash(): void
    {
        $handler = new FileUploadHandler($this->clientMock, $this->fileSystemMock, '/upload/');
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycleStep(fn($msg)=> $msg), 'the cycle must be finished');
        $this->assertTrue($handler->isUploadSuccess(), 'the upload must be successful');
        $this->assertSame(['/upload/a.txt' => 'content'], $this->writtenFiles);
    }
}
