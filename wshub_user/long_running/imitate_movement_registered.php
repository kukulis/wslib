<?php

// Usage: php long_running/imitate_movement_registered.php [tag]
// The movement goes to every logged in client carrying the tag; map2.php registers its pages with the tag "map".

require_once __DIR__ . '/../vendor/autoload.php';

$tag = $argv[1] ?? 'map';

$movementImitator = new \App\LongRunning\RegisteredMovementImitator($tag);

$movementImitator->run();
