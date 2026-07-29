<?php

namespace App\Support;

final class ApiData
{
    public static function progress(object $p): array
    {
        $x = [
            'version' => 1,
            'currentLevel' => (int) $p->current_level,
            'coins' => (int) $p->coins,
            'hints' => (int) $p->hints,
            'previewAgainCount' => (int) $p->preview_again_count,
            'removePairCount' => (int) $p->remove_pair_count,
            'stamina' => (int) $p->stamina,
            'maxStamina' => (int) $p->max_stamina,
            'levelStars' => json_decode($p->level_stars, true) ?? (object) [],
            'completedLevels' => json_decode($p->completed_levels, true) ?? [],
            'updatedAt' => strtotime($p->updated_at) * 1000,
        ];
        if ($p->next_stamina_recover_at) {
            $x['nextStaminaRecoverAt'] =
                strtotime($p->next_stamina_recover_at) * 1000;
        }

        return $x;
    }

    public static function row(object $r, array $map): array
    {
        $o = [];
        foreach ($map as $db => $json) {
            if (property_exists($r, $db)) {
                $o[$json] = $r->$db;
            }
        }

        return $o;
    }
}
