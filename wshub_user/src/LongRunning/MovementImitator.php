<?php

namespace App\LongRunning;

use App\Messages\ObjectMovedMessage;
use App\WsHubUser\Container;
use WsGatewayClient\Dto\Envelope;

class MovementImitator
{
    private const string CLIENT_ID = 'abc';
    private const string DATA_FILE = __DIR__ . '/../../long_running/movement_data.csv';

    public function run() {
        $client = (new Container())->getGatewayClient();
        $points = $this->loadPoints();

        for ($i = 0; $i < 1000000000; $i++) {
            [$latitude, $longitude] = $points[$i % count($points)];

            $message = new ObjectMovedMessage();
            $message->latitude = $latitude;
            $message->longitude = $longitude;

            try {
                $client->sendClientMessage(
                    self::CLIENT_ID,
                    '',
                    (new Envelope())
                        ->setType(ObjectMovedMessage::TYPE)
                        ->setData(get_object_vars($message))
                );
                echo "step $i: $latitude, $longitude\n";
            } catch (\Throwable $e) {
                echo "step $i failed: " . $e->getMessage() . "\n";
            }

            sleep(1);
        }
    }

    /**
     * @return array<int, array{0: float, 1: float}> [latitude, longitude] pairs
     */
    private function loadPoints(): array
    {
        $points = [];
        $file = new \SplFileObject(self::DATA_FILE);
        while (!$file->eof()) {
            $row = $file->fgetcsv();
            if (!is_array($row) || count($row) < 2) {
                continue;
            }
            $points[] = [(float)$row[0], (float)$row[1]];
        }

        if ($points === []) {
            throw new \RuntimeException('No points in ' . self::DATA_FILE);
        }

        return $points;
    }
}
