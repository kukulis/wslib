<?php

namespace Tests\WsGatewayClient;

use GuzzleHttp\Client as GuzzleClient;
use GuzzleHttp\Psr7\Response;
use PHPUnit\Framework\TestCase;
use WsGatewayClient\Client;

class ClientCreateSendTagMessageResponseTest extends TestCase
{
    private Client $client;

    protected function setUp(): void
    {
        // createSendTagMessageResponse() does not use the network
        $this->client = new Client(new GuzzleClient(), 'http://wshub:8585');
    }

    public function testSuccessReturnsTheDeliveredCount(): void
    {
        $response = $this->client->createSendTagMessageResponse(new Response(200, [], '{"delivered": 3}'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertSame(3, $response->delivered);
    }

    public function testTagWithoutClientsIsASuccessWithZeroDelivered(): void
    {
        $response = $this->client->createSendTagMessageResponse(new Response(200, [], '{"delivered": 0}'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage, 'a tag without clients is not an error');
        $this->assertSame(0, $response->delivered, 'zero delivered is not the same as unknown');
    }

    public function testSuccessWithBodyThatIsNotJsonHasUnknownDelivered(): void
    {
        $response = $this->client->createSendTagMessageResponse(new Response(200, [], 'not json'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->delivered);
    }

    public function testSuccessWithDeliveredThatIsNotANumberHasUnknownDelivered(): void
    {
        $response = $this->client->createSendTagMessageResponse(new Response(200, [], '{"delivered": "many"}'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->delivered);
    }

    public function testErrorUsesTheGatewayErrorText(): void
    {
        $response = $this->client->createSendTagMessageResponse(
            new Response(400, [], '{"error": "waitResponse is not supported for a message to a tag"}')
        );

        $this->assertSame(400, $response->responseCode);
        $this->assertSame('waitResponse is not supported for a message to a tag', $response->errorMessage);
        $this->assertNull($response->delivered);
    }

    public function testErrorBodyIsNeverReadAsDelivered(): void
    {
        $response = $this->client->createSendTagMessageResponse(
            new Response(400, [], '{"error": "bad", "delivered": 5}')
        );

        $this->assertSame(400, $response->responseCode);
        $this->assertNull($response->delivered);
    }

    public function testErrorWithBodyThatIsNotJsonUsesReasonPhrase(): void
    {
        $response = $this->client->createSendTagMessageResponse(new Response(502, [], 'Bad Gateway page'));

        $this->assertSame(502, $response->responseCode);
        $this->assertSame('Bad Gateway', $response->errorMessage);
        $this->assertNull($response->delivered);
    }
}
