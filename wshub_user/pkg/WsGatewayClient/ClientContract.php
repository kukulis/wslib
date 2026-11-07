<?php

namespace WsGatewayClient;

use Psr\Http\Message\ResponseInterface;
use WsGatewayClient\Dto\Envelope;
use WsGatewayClient\Dto\SendMessageResponse;

interface ClientContract
{
    /**
     * Sends a message to a connection by its id; the connection id is chosen by the client itself.
     * The call waits for the reply of the WS client as long as $envelope->waitResponse milliseconds;
     * 0 means do not wait. The gateway answers are returned as a response, they are not thrown;
     * only a failure to reach the gateway at all throws.
     *
     * @param string $clientId receiving client id
     * @param Envelope $envelope is the message to send
     */
    public function sendClientMessage( string $clientId, string $messageId, Envelope $envelope) : SendMessageResponse;

    /**
     * Sends a message to a client that logged in with a session registered by registerSession.
     * The call waits for the reply of the WS client as long as $envelope->waitResponse milliseconds;
     * 0 means do not wait. The gateway answers are returned as a response, they are not thrown;
     * only a failure to reach the gateway at all throws.
     *
     * @param string $clientId id of the registered client
     * @param Envelope $envelope is the message to send
     */
    public function sendRegisteredClientMessage( string $clientId, string $messageId, Envelope $envelope) : SendMessageResponse;

    /**
     * TODO refactor to return more concrete type
     *
     * Sends a message to every logged-in client that carries the tag. The gateway answers
     * with {"delivered": N}, the number of connections the message was handed to; a tag
     * without clients is not an error. The envelope must not wait for a response.
     *
     * @param string $tag the tag of the receiving clients
     * @param Envelope $envelope is the message to send
     */
    public function sendTagMessage( string $tag, string $messageId, Envelope $envelope) : ResponseInterface;

    /**
     * TODO refactor to return more concrete type
     *
     * Registers a session, so a Web client can log in with the client id and the session token.
     *
     * @param string $clientId id of the client the session belongs to
     * @param string $sessionToken the token the Web client must present when it logs in
     * @param string[] $tags tags of the session; the gateway rejects an empty tag
     */
    public function registerSession( string $clientId, string $sessionToken, array $tags) : ResponseInterface;
}
