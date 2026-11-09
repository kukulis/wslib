<?php

namespace Tests\Mock;

use GuzzleHttp\Psr7\Response;
use Psr\Http\Message\ResponseInterface;
use WsGatewayClient\Dto\Envelope;
use WsGatewayClient\Dto\SendMessageResponse;
use WsGatewayClient\Dto\SendTagMessageResponse;
use WsGatewayClient\ClientContract;

class ClientMock implements ClientContract
{
    private ?\Closure $sendClientMessageClosure = null;
    private ?\Closure $sendRegisteredClientMessageClosure = null;
    private ?\Closure $sendTagMessageClosure = null;
    private ?\Closure $registerSessionClosure = null;

    public function sendClientMessage(string $clientId, string $messageId, Envelope $envelope): SendMessageResponse
    {
        if ($this->sendClientMessageClosure === null) {
            return new SendMessageResponse(200, null, null);
        }
        $sendClientMessageClosure = $this->sendClientMessageClosure;

        return $sendClientMessageClosure($clientId, $messageId, $envelope);
    }

    public function sendRegisteredClientMessage(
        string $clientId,
        string $messageId,
        Envelope $envelope
    ): SendMessageResponse {
        if ( $this->sendRegisteredClientMessageClosure === null ) {
            return new SendMessageResponse(200, null, null);
        }

        $sendRegisteredClientMessageClosure = $this->sendRegisteredClientMessageClosure;

        return $sendRegisteredClientMessageClosure($clientId, $messageId, $envelope);
    }

    public function sendTagMessage(string $tag, string $messageId, Envelope $envelope): SendTagMessageResponse
    {
        if ( $this->sendTagMessageClosure === null ) {
            return new SendTagMessageResponse(200, null, 0);
        }

        $sendTagMessageClosure = $this->sendTagMessageClosure;

        return $sendTagMessageClosure($tag, $messageId, $envelope);
    }

    public function registerSession(string $clientId, string $sessionToken, array $tags): ResponseInterface
    {
        if ( $this->registerSessionClosure === null ) {
            return new Response(200, [], 'TODO: Implement registerSession() method.');
        }

        $registerSessionClosure = $this->registerSessionClosure;

        return $registerSessionClosure($clientId, $sessionToken, $tags);
    }

    public function setSendClientMessageClosure(?\Closure $sendClientMessageClosure): static
    {
        $this->sendClientMessageClosure = $sendClientMessageClosure;
        return $this;
    }

    public function setSendRegisteredClientMessageClosure(?\Closure $sendRegisteredClientMessageClosure): static
    {
        $this->sendRegisteredClientMessageClosure = $sendRegisteredClientMessageClosure;
        return $this;
    }

    public function setSendTagMessageClosure(?\Closure $sendTagMessageClosure): static
    {
        $this->sendTagMessageClosure = $sendTagMessageClosure;
        return $this;
    }

    public function setRegisterSessionClosure(?\Closure $registerSessionClosure): static
    {
        $this->registerSessionClosure = $registerSessionClosure;
        return $this;
    }
}