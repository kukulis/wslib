<?php

namespace WsGatewayClient\Dto;

/**
 * The gateway's answer to a message sent to one client, in a fixed shape.
 */
class SendMessageResponse
{
    /**
     * @param int $responseCode the HTTP status code of the gateway answer
     * @param string|null $errorMessage the gateway's error text; null when the code is 2xx
     * @param Envelope|null $wsMessage the message the WS client replied with, wrapped in an envelope like
     *                                 every message; null when nothing was waited for, the reply is not an
     *                                 envelope, or the gateway answered with an error
     */
    public function __construct(
        public readonly int $responseCode,
        public readonly ?string $errorMessage,
        public readonly ?Envelope $wsMessage,
    ) {
    }
}
