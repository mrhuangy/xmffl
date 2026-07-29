<?php

namespace App\Http\Controllers;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class LegacyGameController extends Controller
{
    private function openId(Request $request): string
    {
        return (string) ($request->header('X-Openid') ?: $request->query('openId', ''));
    }

    private function ensurePlayer(string $openId): int
    {
        $id = DB::table('players')->where('open_id', $openId)->value('id');
        if ($id) {
            return (int) $id;
        }

        return DB::table('players')->insertGetId(['open_id' => $openId, 'last_login_at' => now()]);
    }

    public function progress(Request $request): JsonResponse
    {
        $openId = $this->openId($request);
        if ($openId === '') {
            return response()->json(['error' => 'openId is required'], 400);
        }
        $row = DB::table('players as p')->join('player_progress as pp', 'pp.player_id', '=', 'p.id')
            ->where('p.open_id', $openId)->select('p.id as player_id', 'p.open_id', 'pp.*')->first();

        return response()->json($row ? $this->mapProgress($row) : [
            'playerId' => 0, 'openId' => $openId, 'currentLevel' => 1, 'coins' => 0,
            'stamina' => 5, 'maxStamina' => 5, 'hints' => 3, 'previewAgainCount' => 3,
            'removePairCount' => 3, 'levelStars' => (object) [], 'completedLevels' => [],
            'updatedAt' => now()->toISOString(),
        ]);
    }

    public function saveProgress(Request $request): JsonResponse
    {
        $data = $request->all();
        $openId = (string) ($data['openId'] ?? $this->openId($request));
        if ($openId === '') {
            return response()->json(['error' => 'openId is required'], 400);
        }
        $id = $this->ensurePlayer($openId);
        DB::table('player_progress')->updateOrInsert(['player_id' => $id], [
            'current_level' => $data['currentLevel'] ?? 0, 'coins' => $data['coins'] ?? 0,
            'stamina' => $data['stamina'] ?? 0, 'max_stamina' => $data['maxStamina'] ?? 0,
            'hints' => $data['hints'] ?? 0, 'preview_again_count' => $data['previewAgainCount'] ?? 0,
            'remove_pair_count' => $data['removePairCount'] ?? 0,
            'level_stars' => json_encode($data['levelStars'] ?? (object) []),
            'completed_levels' => json_encode($data['completedLevels'] ?? []),
            'next_stamina_recover_at' => $data['nextStaminaRecoverAt'] ?? null, 'updated_at' => now(),
        ]);

        return response()->json(['status' => 'saved']);
    }

    public function levelResult(Request $request): JsonResponse
    {
        $data = $request->all();
        $openId = (string) ($data['openId'] ?? $this->openId($request));
        if ($openId === '' || (int) ($data['levelId'] ?? 0) <= 0) {
            return response()->json(['error' => 'openId and levelId are required'], 400);
        }
        DB::table('level_results')->insert([
            'player_id' => $this->ensurePlayer($openId), 'level_id' => $data['levelId'],
            'success' => (bool) ($data['success'] ?? false), 'reason' => $data['reason'] ?? '',
            'steps' => $data['steps'] ?? 0, 'mismatch_count' => $data['mismatchCount'] ?? 0,
            'elapsed_ms' => $data['elapsedMs'] ?? 0, 'stars' => $data['stars'] ?? 0,
            'coins_earned' => $data['coinsEarned'] ?? 0, 'used_hints' => $data['usedHints'] ?? 0,
        ]);

        return response()->json(['status' => 'created'], 201);
    }

    public function submitLeaderboard(Request $request): JsonResponse
    {
        $data = $request->all();
        $openId = (string) ($data['openId'] ?? $this->openId($request));
        if ($openId === '' || (int) ($data['levelId'] ?? 0) <= 0) {
            return response()->json(['error' => 'openId and levelId are required'], 400);
        }
        DB::table('leaderboard_entries')->insert([
            'player_id' => $this->ensurePlayer($openId), 'nickname' => $data['nickname'] ?? '',
            'level_id' => $data['levelId'], 'stars' => $data['stars'] ?? 0,
            'steps' => $data['steps'] ?? 0, 'elapsed_ms' => $data['elapsedMs'] ?? 0,
        ]);

        return response()->json(['status' => 'created'], 201);
    }

    public function leaderboard(Request $request): JsonResponse
    {
        $limit = (int) $request->query('limit', 50);
        $limit = $limit <= 0 || $limit > 100 ? 50 : $limit;
        $query = DB::table('leaderboard_entries as le')->join('players as p', 'p.id', '=', 'le.player_id')
            ->selectRaw('p.open_id as openId,le.nickname,le.level_id as levelId,le.stars,le.steps,le.elapsed_ms as elapsedMs,le.submitted_at as submittedAt');
        if ((int) $request->query('levelId', 0) !== 0) {
            $query->where('le.level_id', (int) $request->query('levelId'));
        }

        return response()->json(['entries' => $query->orderByDesc('stars')->orderBy('steps')
            ->orderBy('elapsed_ms')->orderBy('submitted_at')->limit($limit)->get()]);
    }

    public function events(Request $request): JsonResponse
    {
        $events = $request->input('events', []);
        foreach ($events as $event) {
            if (empty($event['eventId']) || empty($event['eventType'])) {
                return response()->json(['error' => 'eventId and eventType are required'], 400);
            }
        }
        foreach ($events as $event) {
            $openId = (string) ($event['openId'] ?? $this->openId($request));
            DB::table('game_events')->insert([
                'event_id' => $event['eventId'], 'player_id' => $openId === '' ? null : $this->ensurePlayer($openId),
                'event_type' => $event['eventType'], 'level_id' => $event['levelId'] ?? null,
                'payload' => json_encode($event['payload'] ?? (object) []), 'created_at' => $event['createdAt'] ?? now(),
            ]);
        }

        return response()->json(['accepted' => count($events)], 201);
    }

    private function mapProgress(object $row): array
    {
        return ['playerId' => (int) $row->player_id, 'openId' => $row->open_id,
            'currentLevel' => (int) $row->current_level, 'coins' => (int) $row->coins,
            'stamina' => (int) $row->stamina, 'maxStamina' => (int) $row->max_stamina,
            'hints' => (int) $row->hints, 'previewAgainCount' => (int) $row->preview_again_count,
            'removePairCount' => (int) $row->remove_pair_count,
            'levelStars' => json_decode($row->level_stars, true) ?? (object) [],
            'completedLevels' => json_decode($row->completed_levels, true) ?? [],
            'nextStaminaRecoverAt' => $row->next_stamina_recover_at, 'updatedAt' => $row->updated_at];
    }
}
