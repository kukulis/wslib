<?php

namespace Tests\Mock;

use App\Handler\FileSystemDelegatorContract;

class FileSystemDelegatorMock implements FileSystemDelegatorContract
{
    private ?\Closure $filePathExistsClosure = null;
    private ?\Closure $writeFileClosure = null;

    public function filePathExists(string $filePath): bool
    {
        if ($this->filePathExistsClosure === null) {
            return false;
        }

        $filePathExistsClosure = $this->filePathExistsClosure;

        return $filePathExistsClosure($filePath);
    }

    public function writeFile(string $filePath, string $content): bool
    {
        if ($this->writeFileClosure === null) {
            return true;
        }

        $writeFileClosure = $this->writeFileClosure;

        return $writeFileClosure($filePath, $content);
    }

    public function setFilePathExistsClosure(?\Closure $filePathExistsClosure): static
    {
        $this->filePathExistsClosure = $filePathExistsClosure;
        return $this;
    }

    public function setWriteFileClosure(?\Closure $writeFileClosure): static
    {
        $this->writeFileClosure = $writeFileClosure;
        return $this;
    }
}
