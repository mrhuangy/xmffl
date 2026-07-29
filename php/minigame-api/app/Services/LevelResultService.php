<?php

namespace App\Services;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class LevelResultService
{
    /**
     * @param  array<string, mixed>  $result
     * @return array{0: object, 1: array{stamina: int}}
     */
    public function submit(object $player, array $result): array
    {
        return DB::transaction(function () use ($player, $result) {
            $progress = DB::table('player_progress')
                ->where('player_id', $player->id)
                ->lockForUpdate()
                ->first();
            $level = DB::table('level_configs')
                ->where('level_id', $result['levelId'])
                ->where('enabled', 1)
                ->first();

            abort_unless($progress, 404, 'player progress unavailable');
            abort_unless($level, 404, 'level unavailable');

            $reason = $this->normalizeReason($result['reason']);
            $success = ($result['success'] ?? false) && $reason === 'completed';
            $stars = $success ? $this->calculateStars($result, $level) : 0;
            $coins = $success ? $this->coinReward($level, $stars) : 0;
            $levelStars = json_decode($progress->level_stars, true) ?? [];
            $completedLevels = json_decode($progress->completed_levels, true) ?? [];
            $previousBestStars = (int) ($levelStars[(string) $result['levelId']] ?? 0);

            $levelResultId = DB::table('level_results')->insertGetId([
                'player_id' => $player->id,
                'level_id' => $result['levelId'],
                'success' => $success,
                'reason' => $reason,
                'steps' => $result['steps'] ?? 0,
                'mismatch_count' => $result['mismatchCount'] ?? 0,
                'elapsed_ms' => $result['elapsedMs'] ?? 0,
                'stars' => $stars,
                'coins_earned' => $coins,
                'used_hints' => $result['usedHints'] ?? 0,
                'completed_at' => now(),
            ]);

            $rewards = ['stamina' => 0];
            if (! $success) {
                return [$progress, $rewards];
            }

            $levelStars[(string) $result['levelId']] = max($previousBestStars, $stars);
            if (! in_array($result['levelId'], $completedLevels, true)) {
                $completedLevels[] = $result['levelId'];
            }

            $newCoins = $progress->coins + $coins;
            if ($coins > 0) {
                $this->recordCoinReward($player->id, $result['levelId'], $levelResultId, $coins, $newCoins);
            }

            $stamina = $progress->stamina;
            $nextRecovery = $progress->next_stamina_recover_at;
            if ($stars === 3 && $previousBestStars < 3) {
                $rewards['stamina'] = 1;
                $stamina++;
                if ($stamina >= $progress->max_stamina) {
                    $nextRecovery = null;
                }
                $this->recordStaminaReward($player->id, $result['levelId'], $levelResultId, $stamina);
            }

            DB::table('player_progress')->where('player_id', $player->id)->update([
                'current_level' => max($progress->current_level, $result['levelId'] + 1),
                'coins' => $newCoins,
                'stamina' => $stamina,
                'next_stamina_recover_at' => $nextRecovery,
                'level_stars' => json_encode($levelStars),
                'completed_levels' => json_encode($completedLevels),
                'updated_at' => now(),
            ]);
            DB::table('leaderboard_entries')->insert([
                'player_id' => $player->id,
                'level_id' => $result['levelId'],
                'nickname' => $player->nickname,
                'stars' => $stars,
                'steps' => $result['steps'] ?? 0,
                'elapsed_ms' => $result['elapsedMs'] ?? 0,
                'submitted_at' => now(),
            ]);

            return [
                DB::table('player_progress')->where('player_id', $player->id)->first(),
                $rewards,
            ];
        });
    }

    /** @param array<string, mixed> $result */
    private function calculateStars(array $result, object $level): int
    {
        $steps = (int) ($result['steps'] ?? 0);
        $elapsedSeconds = (int) ceil(((int) ($result['elapsedMs'] ?? 0)) / 1000);

        if ($steps <= $level->excellent_step_threshold
            && $this->withinThreshold($elapsedSeconds, $level->excellent_time_threshold)) {
            return 3;
        }
        if ($steps <= $level->normal_step_threshold
            && $this->withinThreshold($elapsedSeconds, $level->normal_time_threshold)) {
            return 2;
        }

        return 1;
    }

    private function withinThreshold(int $value, mixed $threshold): bool
    {
        return $threshold === null || $value <= (int) $threshold;
    }

    private function coinReward(object $level, int $stars): int
    {
        $configured = (int) ($level->{'coin_reward_star'.$stars} ?? 0);

        return $configured > 0 ? $configured : $stars * ((int) $level->coin_reward_base ?: 10);
    }

    private function normalizeReason(string $reason): string
    {
        return in_array($reason, ['completed', 'time_out', 'mismatch_limit', 'quit'], true)
            ? $reason
            : 'unknown';
    }

    private function recordCoinReward(int $playerId, int $levelId, int $resultId, int $amount, int $balance): void
    {
        DB::table('coin_transactions')->insert([
            'transaction_no' => $this->makeId('coin'),
            'player_id' => $playerId,
            'change_amount' => $amount,
            'balance_after' => $balance,
            'reason' => 'level_complete',
            'ref_type' => 'level_result',
            'ref_id' => (string) $resultId,
        ]);
        DB::table('reward_grants')->insert([
            'reward_id' => $this->makeId('reward'),
            'player_id' => $playerId,
            'source' => 'level_complete',
            'source_ref' => (string) $resultId,
            'reward_type' => 'coins',
            'amount' => $amount,
            'level_id' => $levelId,
        ]);
    }

    private function recordStaminaReward(int $playerId, int $levelId, int $resultId, int $balance): void
    {
        DB::table('stamina_transactions')->insert([
            'transaction_no' => $this->makeId('stamina'),
            'player_id' => $playerId,
            'change_amount' => 1,
            'balance_after' => $balance,
            'reason' => 'activity_reward',
            'ref_type' => 'level_result',
            'ref_id' => (string) $resultId,
            'note' => 'first_3_star',
        ]);
        DB::table('reward_grants')->insert([
            'reward_id' => $this->makeId('reward'),
            'player_id' => $playerId,
            'source' => 'activity',
            'source_ref' => (string) $resultId,
            'reward_type' => 'stamina',
            'amount' => 1,
            'level_id' => $levelId,
        ]);
    }

    private function makeId(string $prefix): string
    {
        return $prefix.'_'.Str::uuid();
    }
}
