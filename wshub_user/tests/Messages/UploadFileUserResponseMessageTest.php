<?php

namespace Tests\Messages;

use App\Messages\UploadFileUserResponseMessage;
use PHPUnit\Framework\TestCase;

class UploadFileUserResponseMessageTest extends TestCase
{
    public function testFromWsMessageReadsBothFields(): void
    {
        $message = UploadFileUserResponseMessage::fromArray(
            ['correctedFileName' => 'good.txt', 'cancelUpload' => true]
        );

        $this->assertSame('good.txt', $message->correctedFileName);
        $this->assertTrue($message->cancelUpload);
    }

    public function testFromWsMessageUsesDefaultsForMissingFields(): void
    {
        $message = UploadFileUserResponseMessage::fromArray([]);

        $this->assertSame('', $message->correctedFileName);
        $this->assertFalse($message->cancelUpload);
    }

    public function testFromWsMessageIgnoresTheFieldsTheGatewayAdds(): void
    {
        $message = UploadFileUserResponseMessage::fromArray([
            'message_type' => 'upload_file_user_response',
            'message_id' => 'abc',
            'correctedFileName' => 'good.txt',
        ]);

        $this->assertSame('good.txt', $message->correctedFileName);
        $this->assertFalse($message->cancelUpload);
    }

    public function testFromWsMessageConvertsTheTypesOfTheFields(): void
    {
        $message = UploadFileUserResponseMessage::fromArray(['correctedFileName' => 123, 'cancelUpload' => 1]);

        $this->assertSame('123', $message->correctedFileName);
        $this->assertTrue($message->cancelUpload);
    }
}
