<?php

use App\Http\Controllers\ClientErrorController;
use App\Http\Controllers\GameController;
use App\Http\Controllers\PublicController;
use Illuminate\Support\Facades\Route;

Route::get('/healthz', fn () => response()->json(['status' => 'ok']));
Route::prefix('/api/v1')->group(function () {
    Route::post('/auth/login', [PublicController::class, 'login']);
    Route::post('/client-errors', [ClientErrorController::class, 'store'])
        ->middleware('throttle:30,1');
    Route::get('/config/init', [PublicController::class, 'init']);
    Route::get('/config/levels', [PublicController::class, 'levels']);
    Route::get('/config/ads', [PublicController::class, 'ads']);
    Route::get('/leaderboard', [PublicController::class, 'leaderboard']);
    Route::middleware('player')->group(function () {
        Route::get('/player/progress', [GameController::class, 'progress']);
        Route::post('/levels/start', [GameController::class, 'start']);
        Route::post('/levels/results', [GameController::class, 'result']);
        Route::post('/tools/change', [GameController::class, 'changeTool']);
        Route::post('/tools/purchase', [GameController::class, 'purchaseTool']);
        Route::get('/shop/products', [GameController::class, 'products']);
        Route::post('/shop/purchase', [GameController::class, 'purchase']);
        Route::post('/events/batch', [GameController::class, 'events']);
    });
});
