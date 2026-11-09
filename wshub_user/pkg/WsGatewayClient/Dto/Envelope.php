<?php

namespace WsGatewayClient\Dto;

class Envelope
{
    public string $id ='';
    public string $type ='';
    public mixed $data = [];
    public int $waitResponse = 0;

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

    public function setData(mixed $data): Envelope
    {
        $this->data = $data;
        return $this;
    }

    public function setWaitResponse(int $waitResponse): Envelope
    {
        $this->waitResponse = $waitResponse;
        return $this;
    }
}