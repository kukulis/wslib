<?php

namespace App\Messages;

class ObjectMovedMessage
{
    public const string TYPE = 'object_moved';
    public string $id = "123";

    public float $latitude = 53;
    public float $longitude = 27;
}