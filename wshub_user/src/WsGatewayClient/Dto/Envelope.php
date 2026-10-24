<?php

namespace WsGatewayClient\Dto;

class Envelope
{
    public string $id ='';
    public string $type ='';
    public array $data = [];

    public function setId(string $id): Envelope
    {
        $this->id = $id;
        return $this;
    }

    public function setType(string $type): Envelope
    {
        $this->type = $type;
        return $this;
    }

    public function setData(array $data): Envelope
    {
        $this->data = $data;
        return $this;
    }
}