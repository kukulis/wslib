<?php

namespace WsGatewayClient;

use GuzzleHttp\Client as GuzzleClient;
use Psr\Http\Message\ResponseInterface;
use WsGatewayClient\Dto\Envelope;
use WsGatewayClient\Dto\SendMessageResponse;
use WsGatewayClient\Dto\SendTagMessageResponse;

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
     * The call waits for the reply of the WS client as long as $envelope->waitResponse milliseconds;
     * 0 means do not wait. The gateway answers are returned as a response, they are not thrown;
     * only a failure to reach the gateway at all throws.
     *
     * @param string $clientId receiving client id
     * @param Envelope $envelope is the message to send
     */
    public function sendClientMessage( string $clientId, string $messageId, Envelope $envelope) : SendMessageResponse {
        $response = $this->client->post($this->uri('/api/client-message'), [
            'json' => $envelope,
            'query' => ['clientId' => $clientId, 'messageId' => $messageId,],
            'http_errors' => false,
        ]);

        return $this->createSendMessageResponse($response);
    }

    /**
     * Sends a message to a client that logged in with a session registered by registerSession.
     * The call waits for the reply of the WS client as long as $envelope->waitResponse milliseconds;
     * 0 means do not wait. The gateway answers are returned as a response, they are not thrown;
     * only a failure to reach the gateway at all throws.
     *
     * @param string $clientId id of the registered client
     * @param Envelope $envelope is the message to send
     */
    public function sendRegisteredClientMessage( string $clientId, string $messageId, Envelope $envelope) : SendMessageResponse {
        $response = $this->client->post($this->uri('/api/registered-client-message'), [
            'json' => $envelope,
            'query' => ['clientId' => $clientId, 'messageId' => $messageId,],
            'http_errors' => false,
        ]);

        return $this->createSendMessageResponse($response);
    }

    /**
     * Sends a message to every logged in client that carries the tag. The gateway answers
     * with {"delivered": N}, the number of connections the message was handed to; a tag
     * without clients is not an error. The envelope must not wait for a response.
     *
     * @param string $tag the tag of the receiving clients
     * @param Envelope $envelope is the message to send
     */
    public function sendTagMessage( string $tag, string $messageId, Envelope $envelope) : SendTagMessageResponse {

        $body = json_encode($envelope);
        $response = $this->client->post($this->uri('/api/tag-message'), [
            'body' => $body,
            'query' => ['tag' => $tag, 'messageId' => $messageId,],
        ]);

        return $this->createSendTagMessageResponse($response);
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

    public function createSendMessageResponse(ResponseInterface $response) : SendMessageResponse {
        $responseCode = $response->getStatusCode();
        $body = json_decode((string)$response->getBody(), true);

        if ($responseCode >= 200 && $responseCode < 300) {
            // a reply is an envelope; a body without a type is not a reply
            $wsMessage = is_array($body) && is_string($body['type'] ?? null) ? Envelope::fromArray($body) : null;

            return new SendMessageResponse($responseCode, null, $wsMessage);
        }

        $errorMessage = is_array($body) && is_string($body['error'] ?? null)
            ? $body['error']
            : $response->getReasonPhrase();

        return new SendMessageResponse($responseCode, $errorMessage, null);
    }

    public function createSendTagMessageResponse(ResponseInterface $response) : SendTagMessageResponse {
        $responseCode = $response->getStatusCode();
        $body = json_decode((string)$response->getBody(), true);

        if ($responseCode >= 200 && $responseCode < 300) {
            $delivered = is_array($body) && is_int($body['delivered'] ?? null) ? $body['delivered'] : null;

            return new SendTagMessageResponse($responseCode, null, $delivered);
        }

        $errorMessage = is_array($body) && is_string($body['error'] ?? null)
            ? $body['error']
            : $response->getReasonPhrase();

        return new SendTagMessageResponse($responseCode, $errorMessage, null);
    }

    private function uri(string $path) : string {
        return rtrim($this->baseUri, '/') . $path;
    }
}
