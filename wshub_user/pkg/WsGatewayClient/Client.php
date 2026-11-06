<?php

namespace WsGatewayClient;

use GuzzleHttp\Client as GuzzleClient;
use Psr\Http\Message\ResponseInterface;
use WsGatewayClient\Dto\Envelope;

class Client implements ClientContract
{
    /**
     * @param string $baseUri the gateway address without a path, for example http://wshub:8585
     */
    public function __construct(
        private GuzzleClient $client,
        private string $baseUri,
    ) {
    }

    /**
     * Sends a message to a connection by its id; the connection id is chosen by the client itself.
     *
     * @param string $clientId receiving client id
     * @param Envelope $envelope is the message to send
     */
    public function sendClientMessage( string $clientId, string $messageId, Envelope $envelope) : ResponseInterface {
        return $this->client->post($this->uri('/api/client-message'), [
            'json' => $envelope,
            'query' => ['clientId' => $clientId, 'messageId' => $messageId,],
        ]);
    }

    /**
     * Sends a message to a client that logged in with a session registered by registerSession.
     *
     * @param string $clientId id of the registered client
     * @param Envelope $envelope is the message to send
     */
    public function sendRegisteredClientMessage( string $clientId, string $messageId, Envelope $envelope) : ResponseInterface {
        return $this->client->post($this->uri('/api/registered-client-message'), [
            'json' => $envelope,
            'query' => ['clientId' => $clientId, 'messageId' => $messageId,],
        ]);
    }

    /**
     * Sends a message to every logged in client that carries the tag. The gateway answers
     * with {"delivered": N}, the number of connections the message was handed to; a tag
     * without clients is not an error. The envelope must not wait for a response.
     *
     * @param string $tag the tag of the receiving clients
     * @param Envelope $envelope is the message to send
     */
    public function sendTagMessage( string $tag, string $messageId, Envelope $envelope) : ResponseInterface {
        return $this->client->post($this->uri('/api/tag-message'), [
            'json' => $envelope,
            'query' => ['tag' => $tag, 'messageId' => $messageId,],
        ]);
    }

    /**
     * Registers a session, so a Web client can log in with the client id and the session token.
     *
     * @param string $clientId id of the client the session belongs to
     * @param string $sessionToken the token the Web client must present when it logs in
     * @param string[] $tags tags of the session; the gateway rejects an empty tag
     */
    public function registerSession( string $clientId, string $sessionToken, array $tags) : ResponseInterface {
        return $this->client->post($this->uri('/api/register-session'), [
            'query' => ['clientId' => $clientId, 'sessionToken' => $sessionToken, 'tags' => implode(',', $tags),],
        ]);
    }

    private function uri(string $path) : string {
        return rtrim($this->baseUri, '/') . $path;
    }
}
