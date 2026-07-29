<?php

use App\Http\Controllers\AdminController;
use App\Http\Controllers\AuthController;
use App\Http\Controllers\ConfigController;
use App\Http\Controllers\LegacyGameController;
use Illuminate\Support\Facades\Route;

Route::get('/healthz', fn () => response()->json(['status' => 'ok']));
Route::post('/api/auth/login', [AuthController::class, 'login']);
Route::middleware('admin')->group(function () {
    Route::get('/api/auth/me', [AuthController::class, 'me']);
    Route::put('/api/admin/levels/{id}', [
        ConfigController::class,
        'saveLevel',
    ]);
    Route::put('/api/admin/config/ads', [ConfigController::class, 'saveAds']);
    Route::get('/api/admin/players', [AdminController::class, 'players']);
    Route::get('/api/admin/players/{id}', [AdminController::class, 'player']);
    foreach (['users', 'system-controls'] as $resource) {
        Route::get("/api/admin/$resource", [AdminController::class, 'index']);
        Route::post("/api/admin/$resource", [AdminController::class, 'store']);
        Route::get("/api/admin/$resource/{id}", [AdminController::class, 'show']);
        Route::put("/api/admin/$resource/{id}", [AdminController::class, 'update']);
        Route::delete("/api/admin/$resource/{id}", [AdminController::class, 'destroy']);
    }
});
Route::get('/api/config/levels', [ConfigController::class, 'levels']);
Route::get('/api/config/ads', [ConfigController::class, 'ads']);
Route::get('/api/player/progress', [LegacyGameController::class, 'progress']);
Route::post('/api/player/progress', [
    LegacyGameController::class,
    'saveProgress',
]);
Route::post('/api/player/level-results', [
    LegacyGameController::class,
    'levelResult',
]);
Route::post('/api/leaderboard/submit', [
    LegacyGameController::class,
    'submitLeaderboard',
]);
Route::get('/api/leaderboard', [LegacyGameController::class, 'leaderboard']);
Route::post('/api/events/batch', [LegacyGameController::class, 'events']);
