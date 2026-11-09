<?php

namespace Tests\WsGatewayClient;

use GuzzleHttp\Client as GuzzleClient;
use GuzzleHttp\Psr7\Response;
use PHPUnit\Framework\TestCase;
use WsGatewayClient\Client;
use WsGatewayClient\Dto\Envelope;

class ClientCreateSendMessageResponseTest extends TestCase
{
    private Client $client;

    protected function setUp(): void
    {
        // createSendMessageResponse() does not use the network
        $this->client = new Client(new GuzzleClient(), 'http://wshub:8585');
    }

    public function testSuccessWithEmptyBodyHasNoWsMessage(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(200, [], ''));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testSuccessWithReplyReturnsItAsWsMessage(): void
    {
        $response = $this->client->createSendMessageResponse(
            new Response(200, [], '{"id": "m1", "type": "answer", "data": {"cancelUpload": true}}')
        );

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertInstanceOf(Envelope::class, $response->wsMessage, 'the reply must arrive as an envelope');
        $this->assertSame('m1', $response->wsMessage->id);
        $this->assertSame('answer', $response->wsMessage->type);
        $this->assertSame(['cancelUpload' => true], $response->wsMessage->data);
    }

    public function testReplyWithoutDataHasEmptyData(): void
    {
        $response = $this->client->createSendMessageResponse(
            new Response(200, [], '{"id": "m1", "type": "answer"}')
        );

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertSame('answer', $response->wsMessage->type);
        $this->assertSame([], $response->wsMessage->data);
    }

    public function testSuccessWithEmptyObjectHasNoWsMessage(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(200, [], '{}'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->wsMessage, 'a body without a type is not an envelope');
    }

    public function testSuccessWithBodyThatIsNotJsonHasNoWsMessage(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(200, [], 'not json'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testSuccessWithJsonThatIsNotAnObjectHasNoWsMessage(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(200, [], '"just a string"'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testEveryCodeFrom200To299IsSuccess(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(299, [], ''));

        $this->assertSame(299, $response->responseCode);
        $this->assertNull($response->errorMessage, '299 is still a success');
    }

    public function testCode300IsNotSuccess(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(300, [], ''));

        $this->assertSame(300, $response->responseCode);
        $this->assertSame('Multiple Choices', $response->errorMessage, '300 is the first code that is not a success');
        $this->assertNull($response->wsMessage);
    }

    public function testErrorUsesTheGatewayErrorText(): void
    {
        $response = $this->client->createSendMessageResponse(
            new Response(504, [], '{"error": "no response from client within the given wait time"}')
        );

        $this->assertSame(504, $response->responseCode);
        $this->assertSame('no response from client within the given wait time', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testErrorBodyIsNeverReturnedAsWsMessage(): void
    {
        $response = $this->client->createSendMessageResponse(
            new Response(404, [], '{"error": "no client", "type": "x"}')
        );

        $this->assertNull($response->wsMessage);
    }

    public function testErrorJsonWithoutErrorTextUsesReasonPhrase(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(400, [], '{"message": "something else"}'));

        $this->assertSame(400, $response->responseCode);
        $this->assertSame('Bad Request', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testErrorTextThatIsNotAStringUsesReasonPhrase(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(400, [], '{"error": {"code": 5}}'));

        $this->assertSame('Bad Request', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testErrorWithBodyThatIsNotJsonUsesReasonPhrase(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(502, [], 'Bad Gateway page'));

        $this->assertSame(502, $response->responseCode);
        $this->assertSame('Bad Gateway', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testErrorWithEmptyBodyUsesReasonPhrase(): void
    {
        $response = $this->client->createSendMessageResponse(new Response(404, [], ''));

        $this->assertSame(404, $response->responseCode);
        $this->assertSame('Not Found', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }
}
