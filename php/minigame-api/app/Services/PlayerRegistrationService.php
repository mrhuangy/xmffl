<?php

namespace App\Services;

use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;

class PlayerRegistrationService
{
    private const INITIAL_COINS = 100;

    private const INITIAL_TOOL_CHARGES = 3;

    private const STAMINA_RECOVER_MINUTES = 2;

    /**
     * @return array{0: object, 1: object}
     */
    public function upsert(string $openId, ?string $unionId, string $avatarUrl): array
    {
        return DB::transaction(function () use ($openId, $unionId, $avatarUrl) {
            $player = DB::table('players')
                ->where('open_id', $openId)
                ->where('status', '!=', 'deleted')
                ->lockForUpdate()
                ->first();

            if (! $player) {
                $playerId = DB::table('players')->insertGetId([
                    'open_id' => $openId,
                    'union_id' => $unionId,
                    'nickname' => $this->uniqueNickname(),
                    'avatar_url' => $avatarUrl,
                    'status' => 'active',
                    'last_login_at' => now(),
                    'created_at' => now(),
                    'updated_at' => now(),
                ]);
            } else {
                $playerId = $player->id;
                $updates = ['last_login_at' => now(), 'updated_at' => now()];
                if ($unionId !== null && $unionId !== '') {
                    $updates['union_id'] = $unionId;
                }
                if ($avatarUrl !== '') {
                    $updates['avatar_url'] = $avatarUrl;
                }
                DB::table('players')->where('id', $playerId)->update($updates);
            }

            $player = DB::table('players')->where('id', $playerId)->first();
            $progress = DB::table('player_progress')
                ->where('player_id', $playerId)
                ->lockForUpdate()
                ->first();

            if (! $progress) {
                $this->createProgress($player);
            }

            return [$player, $this->settleStaminaRecovery($playerId)];
        });
    }

    private function uniqueNickname(): string
    {
        $prefixes = [
            '星夜', '云端', '闪电', '极光', '月影', '焰火', '霜雪', '幻梦',
            '银河', '青空', '暮光', '晨曦', '深海', '森林', '琥珀', '像素',
            '泡泡', '彩虹', '甜梦', '流星',
        ];
        $characters = [
            '团子', '松果', '布丁', '星尘', '旅人', '骑士', '精灵', '船长',
            '猎手', '守卫', '术士', '游侠', '鲸歌', '萤火', '喵球', '果冻',
            '蘑菇', '云朵', '雪球', '飞鱼',
        ];
        $titles = ['', '', '', '王', '酱', '侠', '号', '大王', '队长', '达人'];

        for ($attempt = 0; $attempt < 80; $attempt++) {
            $nickname = $prefixes[array_rand($prefixes)]
                .$characters[array_rand($characters)]
                .$titles[array_rand($titles)];
            if (! DB::table('players')->where('nickname', $nickname)->exists()) {
                return $nickname;
            }
        }

        return '星际'.strtoupper(substr(base_convert((string) hrtime(true), 10, 36), -4));
    }

    private function createProgress(object $player): void
    {
        DB::table('player_progress')->insert([
            'player_id' => $player->id,
            'current_level' => 1,
            'coins' => self::INITIAL_COINS,
            'stamina' => 5,
            'max_stamina' => 5,
            'hints' => self::INITIAL_TOOL_CHARGES,
            'preview_again_count' => self::INITIAL_TOOL_CHARGES,
            'remove_pair_count' => self::INITIAL_TOOL_CHARGES,
            'level_stars' => '{}',
            'completed_levels' => '[]',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        DB::table('coin_transactions')->insert([
            'transaction_no' => $this->makeId('coin'),
            'player_id' => $player->id,
            'change_amount' => self::INITIAL_COINS,
            'balance_after' => self::INITIAL_COINS,
            'reason' => 'activity_reward',
            'ref_type' => 'register',
            'ref_id' => $player->open_id,
            'note' => 'new player registration initial coins',
        ]);
        DB::table('reward_grants')->insert([
            'reward_id' => $this->makeId('reward'),
            'player_id' => $player->id,
            'source' => 'activity',
            'source_ref' => 'register',
            'reward_type' => 'coins',
            'amount' => self::INITIAL_COINS,
        ]);

        foreach (['hint', 'preview_again', 'remove_pair'] as $toolType) {
            DB::table('tool_transactions')->insert([
                'transaction_no' => $this->makeId('tool'),
                'player_id' => $player->id,
                'tool_type' => $toolType,
                'change_amount' => self::INITIAL_TOOL_CHARGES,
                'balance_after' => self::INITIAL_TOOL_CHARGES,
                'source' => 'register',
                'ref_type' => 'player',
                'ref_id' => $player->open_id,
                'note' => 'initial tool charges',
            ]);
        }
    }

    private function settleStaminaRecovery(int $playerId): object
    {
        $progress = DB::table('player_progress')
            ->where('player_id', $playerId)
            ->lockForUpdate()
            ->first();

        if ($progress->stamina >= $progress->max_stamina) {
            if ($progress->next_stamina_recover_at !== null) {
                DB::table('player_progress')->where('player_id', $playerId)->update([
                    'next_stamina_recover_at' => null,
                    'updated_at' => now(),
                ]);
            }

            return DB::table('player_progress')->where('player_id', $playerId)->first();
        }

        if ($progress->next_stamina_recover_at === null) {
            DB::table('player_progress')->where('player_id', $playerId)->update([
                'next_stamina_recover_at' => now()->addMinutes(self::STAMINA_RECOVER_MINUTES),
                'updated_at' => now(),
            ]);

            return DB::table('player_progress')->where('player_id', $playerId)->first();
        }

        $next = Carbon::parse($progress->next_stamina_recover_at);
        $now = now();
        if ($now->lt($next)) {
            return $progress;
        }

        $recoverCount = 1 + intdiv((int) $next->diffInSeconds($now), self::STAMINA_RECOVER_MINUTES * 60);
        $recoverCount = min($recoverCount, $progress->max_stamina - $progress->stamina);
        $stamina = $progress->stamina + $recoverCount;
        $nextRecovery = $stamina >= $progress->max_stamina
            ? null
            : $next->addMinutes($recoverCount * self::STAMINA_RECOVER_MINUTES);

        DB::table('player_progress')->where('player_id', $playerId)->update([
            'stamina' => $stamina,
            'next_stamina_recover_at' => $nextRecovery,
            'updated_at' => $now,
        ]);
        DB::table('stamina_transactions')->insert([
            'transaction_no' => $this->makeId('stamina'),
            'player_id' => $playerId,
            'change_amount' => $recoverCount,
            'balance_after' => $stamina,
            'reason' => 'auto_recover',
            'ref_type' => 'system',
            'ref_id' => 'natural_recovery',
            'note' => 'natural stamina recovery',
        ]);

        return DB::table('player_progress')->where('player_id', $playerId)->first();
    }

    private function makeId(string $prefix): string
    {
        return $prefix.'_'.str_replace('.', '', uniqid('', true));
    }
}
