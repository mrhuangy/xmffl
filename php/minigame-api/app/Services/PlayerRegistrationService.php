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
        $surnames = ['赵', '钱', '孙', '李', '周', '吴', '郑', '王', '冯', '陈', '褚', '卫', '蒋', '沈', '韩', '杨', '朱', '秦', '尤', '许', '何', '吕', '施', '张', '孔', '曹', '严', '华', '金', '魏', '陶', '姜', '戚', '谢', '邹', '喻', '柏', '水', '窦', '章', '云', '苏', '潘', '葛', '奚', '范', '彭', '郎', '鲁', '韦'];
        $given = ['安', '柏', '辰', '澄', '川', '岚', '宁', '清', '然', '若', '书', '思', '望', '微', '溪', '晓', '言', '一', '予', '知', '舟', '景', '星', '月', '云', '禾', '秋', '南', '北', '青', '夏', '冬', '明', '远', '初', '白', '锦', '乐', '可', '宜'];

        for ($attempt = 0; $attempt < 80; $attempt++) {
            $nickname = $surnames[array_rand($surnames)]
                .$given[array_rand($given)]
                .$given[array_rand($given)];
            if (! DB::table('players')->where('nickname', $nickname)->exists()) {
                return $nickname;
            }
        }

        return '玩家'.now()->format('ymdHis').random_int(100, 999);
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
