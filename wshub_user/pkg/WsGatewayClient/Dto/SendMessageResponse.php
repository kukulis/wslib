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
     * @param array|null $wsMessage the message the WS client replied with; null when nothing
     *                              was waited for, or the gateway answered with an error
     */
    public function __construct(
        public readonly int $responseCode,
        public readonly ?string $errorMessage,
        public readonly ?array $wsMessage,
    ) {
    }
}
