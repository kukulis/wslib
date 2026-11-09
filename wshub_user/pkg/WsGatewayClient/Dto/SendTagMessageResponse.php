<?php

namespace WsGatewayClient\Dto;

/**
 * The gateway's answer to a message sent to every client with a tag, in a fixed shape.
 */
class SendTagMessageResponse
{
    /**
     * @param int $responseCode the HTTP status code of the gateway answer
     * @param string|null $errorMessage the gateway's error text; null when the code is 2xx
     * @param int|null $delivered the number of connections the message was handed to; 0 for a tag
     *                            without logged in clients; null when the gateway answered with an error
     */
    public function __construct(
        public readonly int $responseCode,
        public readonly ?string $errorMessage,
        public readonly ?int $delivered,
    ) {
    }
}
