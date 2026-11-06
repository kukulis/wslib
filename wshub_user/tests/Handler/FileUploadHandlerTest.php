<?php

namespace Tests\Handler;

use App\Handler\FileUploadHandler;
use App\Messages\CorrectUploadFileMessage;
use GuzzleHttp\Psr7\Response;
use PHPUnit\Framework\TestCase;
use Tests\Mock\ClientMock;
use Tests\Mock\FileSystemDelegatorMock;
use WsGatewayClient\Dto\Envelope;

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
            function (string $clientId, string $messageId, Envelope $envelope): Response {
                $this->sentMessages[] = ['clientId' => $clientId, 'envelope' => $envelope];
                return new Response(200, [], '{}');
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

    private function createWsClientAnswerClosure(string $json): \Closure
    {
        return function (string $clientId, string $messageId, Envelope $envelope) use ($json): Response {
            $this->sentMessages[] = ['clientId' => $clientId, 'envelope' => $envelope];
            return new Response(200, [], $envelope->waitResponse > 0 ? $json : '{}');
        };
    }

    public function testValidNameIsWrittenAndResultSent(): void
    {
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycle());
        $this->assertSame(['/upload/a.txt' => 'content'], $this->writtenFiles);
        $this->assertTrue($handler->isUploadSuccess());
        $this->assertSame('a.txt', $handler->getResultFileName());
        $this->assertCount(1, $this->sentMessages);
        $this->assertSame(FileUploadHandler::RESULT_MESSAGE_TYPE, $this->sentMessages[0]['envelope']->type);
        $this->assertTrue($this->sentMessages[0]['envelope']->data['success']);
    }

    public function testWriteFailureIsReported(): void
    {
        $this->fileSystemMock->setWriteFileClosure(fn() => false);
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycle());
        $this->assertFalse($handler->isUploadSuccess());
        $this->assertSame('', $handler->getResultFileName());
        $this->assertFalse($this->sentMessages[0]['envelope']->data['success']);
    }

    public function testWrongNameIsCorrectedByUser(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure('{"correctedFileName": "good.txt", "cancelUpload": false}')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->assertTrue($handler->uploadFileCycle());
        $this->assertSame([], $this->writtenFiles);
        $this->assertSame(CorrectUploadFileMessage::TYPE, $this->sentMessages[0]['envelope']->type);
        $this->assertSame('client1', $this->sentMessages[0]['clientId']);

        $this->assertFalse($handler->uploadFileCycle());
        $this->assertSame(['/upload/good.txt' => 'content'], $this->writtenFiles);
        $this->assertTrue($handler->isUploadSuccess());
        $this->assertSame('good.txt', $handler->getResultFileName());
    }

    public function testExistingFileAsksForAnotherName(): void
    {
        $this->fileSystemMock->setFilePathExistsClosure(fn(string $path) => $path === '/upload/a.txt');
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure('{"correctedFileName": "b.txt"}')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertTrue($handler->uploadFileCycle());
        $this->assertFalse($handler->uploadFileCycle());
        $this->assertSame(['/upload/b.txt' => 'content'], $this->writtenFiles);
        $this->assertTrue($handler->isUploadSuccess());
    }

    public function testUserCancelStopsTheCycle(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            $this->createWsClientAnswerClosure('{"cancelUpload": true}')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycle());
        $this->assertTrue($handler->isUserCancelledUpload());
        $this->assertFalse($handler->isUploadSuccess());
        $this->assertSame([], $this->writtenFiles);
    }

    public function testNoAnswerStopsTheCycle(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            fn() => throw new \RuntimeException('504 gateway timeout')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('bad name.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycle());
        $this->assertFalse($handler->isUserCancelledUpload());
        $this->assertFalse($handler->isUploadSuccess());
    }

    public function testResultSendFailureDoesNotFailTheUpload(): void
    {
        $this->clientMock->setSendRegisteredClientMessageClosure(
            fn() => throw new \RuntimeException('gateway down')
        );
        $handler = $this->createFileUploadHandler();
        $handler->startUploadCycle('a.txt', 'content', 'client1');

        $this->assertFalse($handler->uploadFileCycle());
        $this->assertTrue($handler->isUploadSuccess());
    }

    public function testUploadDirWithTrailingSlash(): void
    {
        $handler = new FileUploadHandler($this->clientMock, $this->fileSystemMock, '/upload/');
        $handler->startUploadCycle('a.txt', 'content', 'client1');
        $handler->uploadFileCycle();

        $this->assertSame(['/upload/a.txt' => 'content'], $this->writtenFiles);
        $this->assertTrue($handler->isUploadSuccess());
    }
}
