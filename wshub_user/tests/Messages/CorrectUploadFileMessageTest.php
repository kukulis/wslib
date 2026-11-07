<?php

namespace Tests\Messages;

use App\Messages\CorrectUploadFileMessage;
use PHPUnit\Framework\TestCase;

class CorrectUploadFileMessageTest extends TestCase
{
    public function testToEnvelopeHasTheTypeAndTheFieldsAsData(): void
    {
        $message = (new CorrectUploadFileMessage())
            ->setMessage('The file already exists')
            ->setFileName('a.txt');

        $envelope = $message->toEnvelope();

        $this->assertSame('correct_upload_file', $envelope->type);
        $this->assertSame(['message' => 'The file already exists', 'fileName' => 'a.txt'], $envelope->data);
        $this->assertSame(0, $envelope->waitResponse, 'waiting for the answer is decided by the sender, not by the message');
    }
}
