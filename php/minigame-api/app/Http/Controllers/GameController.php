<?php

namespace App\Http\Controllers;

use App\Services\LevelResultService;
use App\Support\ApiData;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class GameController extends Controller
{
    private function id(Request $r): int
    {
        return (int) $r->attributes->get('player')->id;
    }

    private function get(int $id): object
    {
        return DB::table('player_progress')->where('player_id', $id)->first();
    }

    public function progress(Request $r): JsonResponse
    {
        return response()->json(ApiData::progress($this->get($this->id($r))));
    }

    public function start(Request $r): JsonResponse
    {
        $level = (int) $r->validate(['levelId' => 'required|integer|min:1'])[
            'levelId'
        ];
        $p = DB::transaction(function () use ($r, $level) {
            $id = $this->id($r);
            $p = DB::table('player_progress')
                ->where('player_id', $id)
                ->lockForUpdate()
                ->first();
            $cfg = DB::table('level_configs')
                ->where('level_id', $level)
                ->where('enabled', 1)
                ->first();
            if (! $cfg) {
                abort(404, 'level unavailable');
            }
            $unlimited = DB::table('system_controls')
                ->where('control_key', 'game.unlimited_stamina')
                ->where('enabled', 1)
                ->value('value_text');
            $cost = (int) ($cfg->stamina_cost ?? 1);
            if (
                ! in_array(strtolower((string) $unlimited), [
                    '1',
                    'true',
                    'yes',
                    'on',
                ])
            ) {
                if ($p->stamina < $cost) {
                    return null;
                }
                DB::table('player_progress')
                    ->where('player_id', $id)
                    ->update([
                        'stamina' => $p->stamina - $cost,
                        'next_stamina_recover_at' => $p->next_stamina_recover_at ?:
                            now()->addMinutes(30),
                        'updated_at' => now(),
                    ]);
            }

            return $this->get($id);
        });
        if (! $p) {
            return response()->json(['error' => 'insufficient stamina'], 409);
        }

        return response()->json(['progress' => ApiData::progress($p)]);
    }

    public function result(Request $r, LevelResultService $results): JsonResponse
    {
        $d = $r->validate([
            'levelId' => 'required|integer',
            'success' => 'boolean',
            'reason' => 'required|string',
            'steps' => 'integer',
            'mismatchCount' => 'integer',
            'elapsedMs' => 'integer',
            'usedHints' => 'integer',
        ]);
        [$p, $reward] = $results->submit($r->attributes->get('player'), $d);

        return response()->json(
            ['progress' => ApiData::progress($p), 'rewards' => $reward],
            201,
        );
    }

    public function changeTool(Request $r): JsonResponse
    {
        $d = $r->validate([
            'toolType' => 'required|string',
            'delta' => 'required|integer|not_in:0',
        ]);

        return $this->tool(
            $r,
            $d['toolType'],
            $d['delta'],
            $d['delta'] < 0 ? 'use' : 'ad_reward',
        );
    }

    public function purchaseTool(Request $r): JsonResponse
    {
        $d = $r->validate(['toolType' => 'required|string']);
        $type =
            ['previewAgain' => 'preview_again', 'removePair' => 'remove_pair'][
                $d['toolType']
            ] ?? $d['toolType'];
        $result = DB::transaction(function () use ($r) {
            $id = $this->id($r);
            $p = DB::table('player_progress')
                ->where('player_id', $id)
                ->lockForUpdate()
                ->first();
            if ($p->coins < 300) {
                return null;
            }
            DB::table('player_progress')
                ->where('player_id', $id)
                ->decrement('coins', 300);

            return true;
        });
        if (! $result) {
            return response()->json(['error' => 'insufficient coins'], 409);
        }

        return $this->tool($r, $type, 1, 'shop_purchase');
    }

    private function tool(
        Request $r,
        string $type,
        int $delta,
        string $source,
    ): JsonResponse {
        $col =
            [
                'hint' => 'hints',
                'preview_again' => 'preview_again_count',
                'previewAgain' => 'preview_again_count',
                'remove_pair' => 'remove_pair_count',
                'removePair' => 'remove_pair_count',
            ][$type] ?? null;
        if (! $col) {
            return response()->json(['error' => 'invalid tool type'], 400);
        }
        $p = DB::transaction(function () use ($r, $col, $delta) {
            $id = $this->id($r);
            $p = DB::table('player_progress')
                ->where('player_id', $id)
                ->lockForUpdate()
                ->first();
            if ($p->$col + $delta < 0) {
                return null;
            }
            DB::table('player_progress')
                ->where('player_id', $id)
                ->update([$col => $p->$col + $delta, 'updated_at' => now()]);

            return $this->get($id);
        });
        if (! $p) {
            return response()->json(
                ['error' => 'insufficient tool charges'],
                409,
            );
        }

        return response()->json(['progress' => ApiData::progress($p)]);
    }

    public function products(): JsonResponse
    {
        return response()->json([
            'products' => DB::table('shop_products')
                ->where('enabled', 1)
                ->orderBy('sort_order')
                ->get()
                ->map(
                    fn ($x) => ApiData::row($x, [
                        'id' => 'id',
                        'product_key' => 'productKey',
                        'name' => 'name',
                        'product_type' => 'productType',
                        'currency_type' => 'currencyType',
                        'currency_amount' => 'currencyAmount',
                        'grant_type' => 'grantType',
                        'grant_amount' => 'grantAmount',
                        'daily_buy_limit' => 'dailyBuyLimit',
                        'sort_order' => 'sortOrder',
                        'updated_at' => 'updatedAt',
                    ]),
                ),
        ]);
    }

    public function purchase(Request $r): JsonResponse
    {
        $key = $r->validate(['productKey' => 'required|string'])['productKey'];
        $x = DB::transaction(function () use ($r, $key) {
            $id = $this->id($r);
            $product = DB::table('shop_products')
                ->where('product_key', $key)
                ->where('enabled', 1)
                ->lockForUpdate()
                ->first();
            if (! $product) {
                return ['error' => 'product unavailable', 'code' => 404];
            }
            $p = DB::table('player_progress')
                ->where('player_id', $id)
                ->lockForUpdate()
                ->first();
            if (
                $product->currency_type !== 'coins' ||
                $p->coins < $product->currency_amount
            ) {
                return ['error' => 'insufficient coins', 'code' => 409];
            }
            $order = 'order_'.Str::uuid();
            DB::table('purchase_orders')->insert([
                'order_no' => $order,
                'player_id' => $id,
                'product_id' => $product->id,
                'product_key' => $product->product_key,
                'product_name' => $product->name,
                'currency_type' => $product->currency_type,
                'currency_amount' => $product->currency_amount,
                'grant_type' => $product->grant_type,
                'grant_amount' => $product->grant_amount,
                'status' => 'fulfilled',
                'paid_at' => now(),
                'fulfilled_at' => now(),
            ]);
            $updates = [
                'coins' => $p->coins - $product->currency_amount,
                'updated_at' => now(),
            ];
            $col =
                [
                    'stamina' => 'stamina',
                    'hints' => 'hints',
                    'coins' => 'coins',
                ][$product->grant_type] ?? null;
            if (! $col) {
                return ['error' => 'product unavailable', 'code' => 404];
            }
            $updates[$col] =
                ($updates[$col] ?? $p->$col) + $product->grant_amount;
            DB::table('player_progress')
                ->where('player_id', $id)
                ->update($updates);

            return [
                'orderNo' => $order,
                'progress' => ApiData::progress($this->get($id)),
            ];
        });
        if (isset($x['error'])) {
            return response()->json(['error' => $x['error']], $x['code']);
        }

        return response()->json($x, 201);
    }

    public function events(Request $r): JsonResponse
    {
        $events = $r->validate(['events' => 'required|array'])['events'];
        foreach ($events as $e) {
            DB::table('game_events')->insertOrIgnore([
                'event_id' => $e['eventId'],
                'player_id' => $this->id($r),
                'event_type' => $e['eventType'],
                'level_id' => $e['levelId'] ?? null,
                'payload' => json_encode($e['payload'] ?? (object) []),
                'created_at' => now(),
            ]);
        }

        return response()->json(['accepted' => count($events)], 201);
    }
}
