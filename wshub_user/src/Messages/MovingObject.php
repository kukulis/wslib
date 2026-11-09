<?php

namespace App\Messages;

class MovingObject
{
    public string $id;
    public float $latitude;
    public float $longitude;

    public function setId(string $id): MovingObject
    {
        $this->id = $id;
        return $this;
    }

    public function setLatitude(float $latitude): MovingObject
    {
        $this->latitude = $latitude;
        return $this;
    }

    public function setLongitude(float $longitude): MovingObject
    {
        $this->longitude = $longitude;
        return $this;
    }
}
