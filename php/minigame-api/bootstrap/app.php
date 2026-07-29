<?php

use App\Http\Middleware\Cors;
use App\Http\Middleware\PlayerAuth;
use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        api: __DIR__.'/../routes/api.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
        apiPrefix: '',
    )
    ->withMiddleware(function (Middleware $m): void {
        $m->append(Cors::class);
        $m->alias(['player' => PlayerAuth::class]);
    })
    ->withExceptions(fn (Exceptions $e) => null)
    ->create();
