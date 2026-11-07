<?php

namespace App\Messages;

class MovingObject
{
    public function __construct(
        public readonly string $id,
        public readonly float $latitude,
        public readonly float $longitude,
    ) {
    }
}
