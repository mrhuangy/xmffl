<?php

namespace Tests\Feature;

use App\Services\PlayerRegistrationService;
use Illuminate\Http\Client\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use Mockery;
use Tests\TestCase;

class AuthLoginTest extends TestCase
{
    public function test_login_exchanges_code_for_real_wechat_open_id(): void
    {
        config()->set('services.wechat.minigame.app_id', 'wx-test-app-id');
        config()->set('services.wechat.minigame.secret', 'test-secret');

        Http::fake([
            'api.weixin.qq.com/sns/jscode2session*' => Http::response([
                'openid' => 'wechat-open-id',
                'session_key' => 'wechat-session-key',
                'unionid' => 'wechat-union-id',
            ]),
        ]);

        $player = $this->player();
        $progress = $this->progress();

        $service = Mockery::mock(PlayerRegistrationService::class);
        $service->shouldReceive('upsert')
            ->once()
            ->with('wechat-open-id', 'wechat-union-id', '')
            ->andReturn([$player, $progress]);
        $this->app->instance(PlayerRegistrationService::class, $service);

        $this->postJson('/api/v1/auth/login', [
            'code' => 'wx-login-code',
            'nickname' => 'Test',
            'avatarUrl' => '',
        ])->assertOk()
            ->assertJsonPath('token', 'wechat-open-id')
            ->assertJsonPath('player.openId', 'wechat-open-id')
            ->assertJsonPath('player.unionId', 'wechat-union-id');

        Http::assertSent(
            fn (Request $request) => $request['appid'] === 'wx-test-app-id'
            && $request['secret'] === 'test-secret'
            && $request['js_code'] === 'wx-login-code'
            && $request['grant_type'] === 'authorization_code'
        );
    }

    public function test_login_rejects_wechat_errors_without_creating_player(): void
    {
        config()->set('services.wechat.minigame.app_id', 'wx-test-app-id');
        config()->set('services.wechat.minigame.secret', 'test-secret');

        Http::fake([
            'api.weixin.qq.com/sns/jscode2session*' => Http::response([
                'errcode' => 40029,
                'errmsg' => 'invalid code',
            ]),
        ]);

        DB::shouldReceive('transaction')->never();

        $this->postJson('/api/v1/auth/login', [
            'code' => 'invalid-code',
        ])->assertInternalServerError();
    }

    private function player(): object
    {
        return (object) [
            'id' => 1,
            'open_id' => 'wechat-open-id',
            'union_id' => 'wechat-union-id',
            'nickname' => 'Test',
            'avatar_url' => '',
            'status' => 'active',
            'last_login_at' => '2026-07-28 12:00:00',
            'created_at' => '2026-07-28 12:00:00',
            'updated_at' => '2026-07-28 12:00:00',
        ];
    }

    private function progress(): object
    {
        return (object) [
            'current_level' => 1,
            'coins' => 100,
            'stamina' => 5,
            'max_stamina' => 5,
            'hints' => 3,
            'preview_again_count' => 3,
            'remove_pair_count' => 3,
            'level_stars' => '{}',
            'completed_levels' => '[]',
            'next_stamina_recover_at' => null,
            'updated_at' => '2026-07-28 12:00:00',
        ];
    }
}
