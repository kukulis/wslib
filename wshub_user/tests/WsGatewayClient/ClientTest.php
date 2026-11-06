<?php

namespace Tests\WsGatewayClient;

use GuzzleHttp\Client as GuzzleClient;
use GuzzleHttp\Exception\ConnectException;
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Middleware;
use GuzzleHttp\Psr7\Request;
use GuzzleHttp\Psr7\Response;
use PHPUnit\Framework\TestCase;
use WsGatewayClient\Client;
use WsGatewayClient\Dto\Envelope;

class ClientTest extends TestCase
{
    /** @var array<int, array{request: Request}> filled by the history middleware */
    private array $requests = [];

    private function createClient(MockHandler $gatewayMock): Client
    {
        $this->requests = [];
        $handlerStack = HandlerStack::create($gatewayMock);
        $handlerStack->push(Middleware::history($this->requests));

        return new Client(new GuzzleClient(['handler' => $handlerStack]), 'http://wshub:8585/');
    }

    public function testSendWithoutWaitingReturnsNoWsMessage(): void
    {
        $client = $this->createClient(new MockHandler([new Response(200, [], '')]));

        $response = $client->sendRegisteredClientMessage('client1', 'm1', (new Envelope())->setType('ping'));

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testWaitedReplyIsReturnedAsWsMessage(): void
    {
        $client = $this->createClient(new MockHandler([
            new Response(200, [], '{"message_type": "answer", "cancelUpload": true}'),
        ]));

        $response = $client->sendRegisteredClientMessage(
            'client1',
            'm1',
            (new Envelope())->setType('ask')->setWaitResponse(1000)
        );

        $this->assertSame(200, $response->responseCode);
        $this->assertNull($response->errorMessage);
        $this->assertSame(['message_type' => 'answer', 'cancelUpload' => true], $response->wsMessage);
    }

    public function testEmptyReplyIsNotTheSameAsNoReply(): void
    {
        $client = $this->createClient(new MockHandler([new Response(200, [], '{}')]));

        $response = $client->sendRegisteredClientMessage(
            'client1',
            'm1',
            (new Envelope())->setType('ask')->setWaitResponse(1000)
        );

        $this->assertSame([], $response->wsMessage);
    }

    public function testGatewayTimeoutIsReturnedNotThrown(): void
    {
        $client = $this->createClient(new MockHandler([
            new Response(504, [], '{"error": "no response from client within the given wait time"}'),
        ]));

        $response = $client->sendRegisteredClientMessage(
            'client1',
            'm1',
            (new Envelope())->setType('ask')->setWaitResponse(1000)
        );

        $this->assertSame(504, $response->responseCode);
        $this->assertSame('no response from client within the given wait time', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testUnknownClientIsReturnedNotThrown(): void
    {
        $client = $this->createClient(new MockHandler([
            new Response(404, [], '{"error": "there is no WS connection for client id client1"}'),
        ]));

        $response = $client->sendRegisteredClientMessage('client1', 'm1', (new Envelope())->setType('ping'));

        $this->assertSame(404, $response->responseCode);
        $this->assertSame('there is no WS connection for client id client1', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testErrorWithoutJsonBodyUsesReasonPhrase(): void
    {
        $client = $this->createClient(new MockHandler([new Response(502, [], 'Bad Gateway page')]));

        $response = $client->sendRegisteredClientMessage('client1', 'm1', (new Envelope())->setType('ping'));

        $this->assertSame(502, $response->responseCode);
        $this->assertSame('Bad Gateway', $response->errorMessage);
        $this->assertNull($response->wsMessage);
    }

    public function testUnreachableGatewayThrows(): void
    {
        $client = $this->createClient(new MockHandler([
            new ConnectException('connection refused', new Request('POST', 'http://wshub:8585')),
        ]));

        $this->expectException(ConnectException::class);

        $client->sendRegisteredClientMessage('client1', 'm1', (new Envelope())->setType('ping'));
    }

    public function testRegisteredClientMessageRequest(): void
    {
        $client = $this->createClient(new MockHandler([new Response(200, [], '')]));

        $client->sendRegisteredClientMessage(
            'client1',
            'm1',
            (new Envelope())->setType('ask')->setData(['a' => 1])->setWaitResponse(1000)
        );

        $request = $this->requests[0]['request'];
        $this->assertSame('POST', $request->getMethod());
        $this->assertSame('/api/registered-client-message', $request->getUri()->getPath());
        $this->assertSame('clientId=client1&messageId=m1', $request->getUri()->getQuery());
        $this->assertSame(
            ['id' => '', 'type' => 'ask', 'data' => ['a' => 1], 'waitResponse' => 1000],
            json_decode((string)$request->getBody(), true)
        );
    }

    public function testClientMessageUsesItsOwnEndpointAndReturnsTheSameShape(): void
    {
        $client = $this->createClient(new MockHandler([new Response(200, [], '{"x": 1}')]));

        $response = $client->sendClientMessage(
            'client1',
            'm1',
            (new Envelope())->setType('ask')->setWaitResponse(1000)
        );

        $this->assertSame('/api/client-message', $this->requests[0]['request']->getUri()->getPath());
        $this->assertSame(200, $response->responseCode);
        $this->assertSame(['x' => 1], $response->wsMessage);
    }
}
