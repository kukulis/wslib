<?php

require_once __DIR__ . '/../vendor/autoload.php';

$movementImitator = new \App\LongRunning\MovementImitator();

$movementImitator->run();
