<?php

namespace App\Messages;

use WsGatewayClient\Dto\Envelope;

/**
 * The positions of the moving objects that moved. It may hold all the objects or only some of them.
 */
class MovingObjectsMessage
{
    public const string TYPE = 'moving_objects';

    /** @var MovingObject[] */
    public array $objects = [];

    public function addObject(MovingObject $object): MovingObjectsMessage
    {
        $this->objects[] = $object;
        return $this;
    }

    public function toEnvelope(): Envelope
    {
        return (new Envelope())
            ->setType(self::TYPE)
            ->setData(['objects' => array_map('get_object_vars', $this->objects)]);
    }
}
