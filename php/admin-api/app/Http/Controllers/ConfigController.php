<?php

namespace App\Http\Controllers;

use App\Support\Jwt;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;

class ConfigController extends Controller
{
    private const LEVEL_MAP = [
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
        'stamina_cost' => 'staminaCost',
        'excellent_step_threshold' => 'excellentStepThreshold',
        'normal_step_threshold' => 'normalStepThreshold',
        'excellent_time_threshold' => 'excellentTimeThreshold',
        'normal_time_threshold' => 'normalTimeThreshold',
        'time_limit_seconds' => 'timeLimitSeconds',
        'step_limit' => 'stepLimit',
        'enabled' => 'enabled',
        'version' => 'version',
        'updated_at' => 'updatedAt',
    ];

    private function map(object $row, array $map): array
    {
        $out = [];
        foreach ($map as $db => $json) {
            if (property_exists($row, $db)) {
                $value = $row->$db;
                if (in_array($db, ['show_steps', 'show_timer', 'show_mismatch', 'enabled'], true)) {
                    $value = (bool) $value;
                }
                if ($db === 'updated_at' && $value) {
                    $value = Carbon::parse($value)->toRfc3339String();
                }
                $out[$json] = $value;
            }
        }

        return $out;
    }

    public function levels(Request $r): JsonResponse
    {
        $include = $r->query('includeDisabled') === 'true';
        if ($include) {
            try {
                Jwt::verify($r->bearerToken() ?? '');
            } catch (\Throwable) {
                return response()->json(
                    ['error' => '登录状态已失效，请重新登录'],
                    401,
                );
            }
        }
        $q = DB::table('level_configs')->orderBy('level_id');
        if (! $include) {
            $q->where('enabled', 1);
        }

        return response()->json([
            'levels' => $q
                ->get()
                ->map(fn ($x) => $this->map($x, self::LEVEL_MAP))
                ->all(),
        ]);
    }

    public function saveLevel(Request $r, int $id): JsonResponse
    {
        $d = $r->all();
        if (
            $id < 1 ||
            ($d['rows'] ?? 0) < 1 ||
            ($d['cols'] ?? 0) < 1 ||
            ($d['pairCount'] ?? 0) < 1
        ) {
            return response()->json(['error' => 'invalid level config'], 400);
        }
        $slots = (int) $d['rows'] * (int) $d['cols'];
        $cards = (int) $d['pairCount'] * 2;
        if (($d['mode'] ?? '') === '' || ($d['themeId'] ?? '') === ''
            || ($slots !== $cards && ! ($slots === $cards + 1 && $slots % 2 === 1))) {
            return response()->json(['error' => 'invalid level config'], 400);
        }
        $row = [];
        foreach (self::LEVEL_MAP as $db => $json) {
            if (
                array_key_exists($json, $d) &&
                ! in_array($db, ['level_id', 'updated_at'])
            ) {
                $row[$db] = $d[$json];
            }
        }
        $row['level_id'] = $id;
        $exists = DB::table('level_configs')->where('level_id', $id)->exists();
        $row['updated_at'] = now();
        if ($exists) {
            $row['version'] = DB::raw('version + 1');
            unset($row['level_id']);
            DB::table('level_configs')->where('level_id', $id)->update($row);
        } else {
            $row['version'] = max(1, (int) ($d['version'] ?? 1));
            DB::table('level_configs')->insert($row);
        }

        return response()->json(['status' => 'saved']);
    }

    public function ads(): JsonResponse
    {
        $x = DB::table('ad_frequency_configs')->where('id', 1)->first();

        return response()->json([
            'noInterstitialBeforeLevel' => (int) $x->no_interstitial_before_level,
            'interstitialEveryLevels' => (int) $x->interstitial_every_levels,
            'maxInterstitialPerDay' => (int) $x->max_interstitial_per_day,
            'maxRevivePerLevel' => (int) $x->max_revive_per_level,
            'bannerEnabledScenes' => json_decode($x->banner_enabled_scenes ?? '[]', true),
            'version' => (int) ($x->version ?? 1),
            'updatedAt' => Carbon::parse($x->updated_at)->toRfc3339String(),
        ]);
    }

    public function saveAds(Request $r): JsonResponse
    {
        $d = $r->validate([
            'noInterstitialBeforeLevel' => 'required|integer|min:0',
            'interstitialEveryLevels' => 'required|integer|min:1',
            'maxInterstitialPerDay' => 'required|integer|min:0',
            'maxRevivePerLevel' => 'required|integer|min:0',
            'bannerEnabledScenes' => 'required|array',
        ]);
        $row = [
            'no_interstitial_before_level' => $d['noInterstitialBeforeLevel'],
            'interstitial_every_levels' => $d['interstitialEveryLevels'],
            'max_interstitial_per_day' => $d['maxInterstitialPerDay'],
            'max_revive_per_level' => $d['maxRevivePerLevel'],
            'banner_enabled_scenes' => json_encode($d['bannerEnabledScenes']),
            'updated_at' => now(),
        ];
        if (DB::table('ad_frequency_configs')->where('id', 1)->exists()) {
            $row['version'] = DB::raw('version + 1');
            DB::table('ad_frequency_configs')->where('id', 1)->update($row);
        } else {
            $row['id'] = 1;
            $row['version'] = 1;
            DB::table('ad_frequency_configs')->insert($row);
        }

        return response()->json(['status' => 'saved']);
    }
}
