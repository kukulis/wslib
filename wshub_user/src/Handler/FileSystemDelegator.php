<?php

namespace App\Handler;

class FileSystemDelegator implements FileSystemDelegatorContract
{
    public function filePathExists(string $filePath): bool
    {
        return file_exists($filePath);
    }

    public function writeFile(string $filePath, string $content): bool
    {
        $rez = file_put_contents($filePath, $content);

        return is_int($rez);
    }
}