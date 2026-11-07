<?php

namespace Tests\Messages;

use App\Messages\UploadFileResultMessage;
use PHPUnit\Framework\TestCase;

class UploadFileResultMessageTest extends TestCase
{
    public function testToEnvelopeHasTheTypeAndTheFieldsAsData(): void
    {
        $message = (new UploadFileResultMessage())
            ->setSuccess(true)
            ->setCancelledByUser(false)
            ->setUploadedFileName('bad name.txt')
            ->setResultFileName('good.txt');

        $envelope = $message->toEnvelope();

        $this->assertSame('upload_file_result', $envelope->type);
        $this->assertSame(
            ['success' => true, 'cancelledByUser' => false, 'uploadedFileName' => 'bad name.txt', 'resultFileName' => 'good.txt'],
            $envelope->data
        );
    }
}
