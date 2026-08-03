<?php

namespace App\Http\Controllers;

use App\Services\ClientErrorLogger;
use App\Services\PlayerRegistrationService;
use App\Support\ApiData;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class PublicController extends Controller
{
    private const LEVEL = [
        'level_id' => 'levelId',
        'rows_count' => 'rows',
        'cols_count' => 'cols',
        'pair_count' => 'pairCount',
        'mode' => 'mode',
        'theme_id' => 'themeId',
        'initial_preview_ms' => 'initialPreviewMs',
        'flip_back_delay_ms' => 'flipBackDelayMs',
        'level_time_limit_seconds' => 'levelTimeLimitSeconds',
        'max_mismatch_count' => 'maxMismatchCount',
        'show_steps' => 'showSteps',
        'show_timer' => 'showTimer',
        'show_mismatch' => 'showMismatch',
        'hint_highlight_ms' => 'hintHighlightMs',
        'coin_reward_base' => 'coinRewardBase',
        'coin_reward_star1' => 'coinRewardStar1',
        'coin_reward_star2' => 'coinRewardStar2',
        'coin_reward_star3' => 'coinRewardStar3',
        'stamina_cost' => 'staminaCost',
        'excellent_step_threshold' => 'excellentStepThreshold',
        'normal_step_threshold' => 'normalStepThreshold',
        'excellent_time_threshold' => 'excellentTimeThreshold',
        'normal_time_threshold' => 'normalTimeThreshold',
        'time_limit_seconds' => 'timeLimitSeconds',
        'step_limit' => 'stepLimit',
        'version' => 'version',
        'updated_at' => 'updatedAt',
    ];

    public function login(
        Request $r,
        PlayerRegistrationService $players,
        ClientErrorLogger $errorLogger,
    ): JsonResponse {
        $d = $r->validate([
            'code' => 'required|string',
            'nickname' => 'nullable|string',
            'avatarUrl' => 'nullable|string',
        ]);
        try {
            $session = $this->code2Session(trim($d['code']));
            $openId = $session['openid'];
            $unionId = $session['unionid'] ?? null;
            [$p, $g] = $players->upsert(
                $openId,
                $unionId,
                $d['avatarUrl'] ?? '',
            );
        } catch (\Throwable $error) {
            $errorLogger->write('backend_auth_login', [
                'source' => 'backend',
                'exception' => $error::class,
                'message' => mb_substr($error->getMessage(), 0, 500),
                'ip' => $r->ip(),
                'user_agent' => mb_substr((string) $r->userAgent(), 0, 255),
            ]);
            throw $error;
        }

        return response()->json([
            'token' => $openId,
            'player' => ApiData::row($p, [
                'id' => 'id',
                'open_id' => 'openId',
                'union_id' => 'unionId',
                'nickname' => 'nickname',
                'avatar_url' => 'avatarUrl',
                'status' => 'status',
                'last_login_at' => 'lastLoginAt',
                'created_at' => 'createdAt',
                'updated_at' => 'updatedAt',
            ]),
            'progress' => ApiData::progress($g),
        ]);
    }

    /**
     * @return array{openid: string, session_key: string, unionid?: string}
     */
    private function code2Session(string $code): array
    {
        if ($code === '') {
            throw new RuntimeException('wx login code is required');
        }

        $appId = config('services.wechat.minigame.app_id');
        $secret = config('services.wechat.minigame.secret');

        if (! $appId || ! $secret) {
            throw new RuntimeException('wechat minigame credentials are not configured');
        }

        $response = Http::acceptJson()
            ->timeout(5)
            ->get('https://api.weixin.qq.com/sns/jscode2session', [
                'appid' => $appId,
                'secret' => $secret,
                'js_code' => $code,
                'grant_type' => 'authorization_code',
            ]);

        if (! $response->successful()) {
            throw new RuntimeException(
                'wechat jscode2session http status '.$response->status(),
            );
        }

        $session = $response->json();
        if (! is_array($session)) {
            throw new RuntimeException(
                'decode wechat jscode2session response failed',
            );
        }

        if (($session['errcode'] ?? 0) !== 0) {
            throw new RuntimeException(sprintf(
                'wechat jscode2session error %d: %s',
                $session['errcode'],
                $session['errmsg'] ?? 'unknown error',
            ));
        }

        if (empty($session['openid'])) {
            throw new RuntimeException(
                'wechat jscode2session returned empty openid',
            );
        }

        return $session;
    }

    public function levels(): JsonResponse
    {
        return response()->json([
            'levels' => DB::table('level_configs')
                ->where('enabled', 1)
                ->orderBy('level_id')
                ->get()
                ->map(fn ($x) => ApiData::row($x, self::LEVEL))
                ->all(),
        ]);
    }

    public function ads(): JsonResponse
    {
        $x = DB::table('ad_frequency_configs')->where('id', 1)->first();

        return response()->json([
            'noInterstitialBeforeLevel' => (int) $x->no_interstitial_before_level,
            'interstitialEveryLevels' => (int) $x->interstitial_every_levels,
            'maxInterstitialPerDay' => (int) $x->max_interstitial_per_day,
            'maxRevivePerLevel' => (int) $x->max_revive_per_level,
            'bannerEnabledScenes' => json_decode(
                $x->banner_enabled_scenes ?? '[]',
                true,
            ),
            'version' => (int) ($x->version ?? 1),
            'updatedAt' => (string) $x->updated_at,
        ]);
    }

    public function init(): JsonResponse
    {
        $controls = [];
        foreach (
            DB::table('system_controls')
                ->where('enabled', 1)
                ->where('is_public', 1)
                ->where(
                    fn ($q) => $q
                        ->whereNull('effective_from')
                        ->orWhere('effective_from', '<=', now()),
                )
                ->where(
                    fn ($q) => $q
                        ->whereNull('effective_until')
                        ->orWhere('effective_until', '>=', now()),
                )
                ->get() as $x
        ) {
            $controls[$x->control_key] =
                $x->value_type === 'json'
                    ? json_decode($x->value_json, true)
                    : ($x->value_type === 'bool'
                        ? in_array(strtolower($x->value_text), [
                            '1',
                            'true',
                            'yes',
                            'on',
                        ])
                        : ($x->value_type === 'int'
                            ? (int) $x->value_text
                            : $x->value_text));
        }

        return response()->json([
            'systemControls' => $controls,
            'levels' => json_decode($this->levels()->getContent(), true)[
                'levels'
            ],
            'ads' => json_decode($this->ads()->getContent(), true),
        ]);
    }

    public function leaderboard(Request $r): JsonResponse
    {
        $q = DB::table('leaderboard_entries as l')
            ->join('players as p', 'p.id', '=', 'l.player_id')
            ->selectRaw(
                'p.open_id as openId,l.nickname,l.level_id as levelId,l.stars,l.steps,l.elapsed_ms as elapsedMs,l.submitted_at as submittedAt',
            );
        if ((int) $r->query('levelId')) {
            $q->where('l.level_id', (int) $r->query('levelId'));
        }

        return response()->json([
            'entries' => $q
                ->orderByDesc('stars')
                ->orderBy('steps')
                ->orderBy('elapsed_ms')
                ->limit(min(100, max(1, (int) $r->query('limit', 50))))
                ->get(),
        ]);
    }
}
