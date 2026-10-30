<?php

namespace App\Messages;

class Envelope
{
    public string $id = '';
    public string $type = '';
    public mixed $data = '';
    public int $waitResponse = 0;
}