<?php

namespace App\LongRunning;

use App\Messages\MovingObject;
use App\Messages\MovingObjectsMessage;
use App\WsHubUser\Container;
use WsGatewayClient\Dto\Envelope;

class RegisteredMovementImitator
{
    private const string DATA_FILE = __DIR__ . '/../../long_running/movement_data.csv';
    private const int STEP = 20;

    private array $movingObjects;

    /**
     * @param string $tag the movement goes to every logged in client carrying this tag
     */
    public function __construct(private string $tag)
    {
        $this->movingObjects = [
            (new MovingObject())->setId('aaa'),
            (new MovingObject())->setId('bbb'),
            (new MovingObject())->setId('ccc'),
            (new MovingObject())->setId('ddd'),
            (new MovingObject())->setId('eee'),
            (new MovingObject())->setId('fff'),
            (new MovingObject())->setId('ggg'),
            (new MovingObject())->setId('hhh'),
            (new MovingObject())->setId('iii'),
            (new MovingObject())->setId('jjj'),
        ];
    }

    public function run()
    {
        $client = (new Container())->getGatewayClient();
        $points = $this->loadPoints();

        for ($i = 0; $i < 1000000000; $i++) {
            $message = new MovingObjectsMessage();
            $j = 0;
            foreach ($this->movingObjects as $object) {
                [$latitude, $longitude] = $points[($i + $j * self::STEP) % count($points)];
                $message->addObject($object->setLatitude($latitude)->setLongitude($longitude));
                $j++;
            }

            try {
                $response = $client->sendTagMessage(
                    $this->tag,
                    '',
                    (new Envelope())
                        ->setType(MovingObjectsMessage::TYPE)
                        ->setData($message)
                );

                $delivered = $response->delivered ?? '?';
                echo "step $i: " . $message->objects[0]->latitude . ',' . $message->objects[0]->longitude . ", delivered to $delivered client(s)\n";
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
