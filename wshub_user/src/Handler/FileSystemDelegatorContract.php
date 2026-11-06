<?php

namespace App\Handler;

interface FileSystemDelegatorContract
{
    public function filePathExists(string $filePath): bool;

    public function writeFile(string $filePath, string $content): bool;
}