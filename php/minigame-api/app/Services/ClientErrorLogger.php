<?php

namespace App\Services;

use Illuminate\Log\Logger;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;

class ClientErrorLogger
{
    public function write(string $stage, array $context): string
    {
        $requestId = (string) Str::uuid();
        $this->logger()->warning('minigame_error', [
            'request_id' => $requestId,
            'stage' => $stage,
            ...$context,
        ]);

        return $requestId;
    }

    private function logger(): Logger
    {
        return Log::build([
            'driver' => 'daily',
            'path' => storage_path('logs/minigame-client-errors.log'),
            'level' => 'warning',
            'days' => 14,
            'locking' => true,
        ]);
    }
}
