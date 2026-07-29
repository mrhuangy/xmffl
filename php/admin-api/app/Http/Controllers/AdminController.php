<?php

namespace App\Http\Controllers;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class AdminController extends Controller
{
    private const USER_MAP = [
        'id' => 'id',
        'username' => 'username',
        'email' => 'email',
        'display_name' => 'displayName',
        'role' => 'role',
        'permissions' => 'permissions',
        'status' => 'status',
        'failed_login_attempts' => 'failedLoginAttempts',
        'locked_until' => 'lockedUntil',
        'password_changed_at' => 'passwordChangedAt',
        'last_login_at' => 'lastLoginAt',
        'last_login_ip' => 'lastLoginIp',
        'created_by' => 'createdBy',
        'created_at' => 'createdAt',
        'updated_at' => 'updatedAt',
    ];

    private const CONTROL_MAP = [
        'id' => 'id',
        'control_key' => 'controlKey',
        'control_group' => 'controlGroup',
        'name' => 'name',
        'description' => 'description',
        'value_type' => 'valueType',
        'value_text' => 'valueText',
        'value_json' => 'valueJson',
        'default_value_text' => 'defaultValueText',
        'default_value_json' => 'defaultValueJson',
        'enabled' => 'enabled',
        'is_public' => 'isPublic',
        'sort_order' => 'sortOrder',
        'version' => 'version',
        'effective_from' => 'effectiveFrom',
        'effective_until' => 'effectiveUntil',
        'created_by' => 'createdBy',
        'updated_by' => 'updatedBy',
        'created_at' => 'createdAt',
        'updated_at' => 'updatedAt',
    ];

    private function owner(Request $request): ?JsonResponse
    {
        return ($request->attributes->get('admin')['role'] ?? '') === 'owner'
            ? null
            : response()->json(['error' => 'owner role required'], 403);
    }

    private function map(object $row, array $map): array
    {
        $result = [];
        foreach ($map as $db => $json) {
            $value = $row->$db ?? null;
            if (
                in_array(
                    $db,
                    ['permissions', 'value_json', 'default_value_json'],
                    true,
                ) &&
                is_string($value)
            ) {
                $value = json_decode($value, true);
            }
            if (in_array($db, ['enabled', 'is_public'], true)) {
                $value = (bool) $value;
            }
            $result[$json] = $value;
        }

        return $result;
    }

    public function players(Request $request): JsonResponse
    {
        $size = min(100, max(1, (int) $request->query('pageSize', 20)));
        $page = max(1, (int) $request->query('page', 1));
        $query = DB::table('players as p')->leftJoin(
            'player_progress as g',
            'g.player_id',
            '=',
            'p.id',
        );
        $status = (string) $request->query('status', '');
        if ($status !== '' && ! in_array($status, ['active', 'blocked', 'deleted'], true)) {
            return response()->json(['error' => 'invalid player status'], 400);
        }
        if ($request->query('keyword')) {
            $query->where(
                fn ($q) => $q
                    ->where(
                        'p.nickname',
                        'like',
                        '%'.$request->query('keyword').'%',
                    )
                    ->orWhere(
                        'p.open_id',
                        'like',
                        '%'.$request->query('keyword').'%',
                    ),
            );
        }
        if ($status !== '') {
            $query->where('p.status', $status);
        }
        $total = (clone $query)->count();
        $map = [
            'id' => 'id',
            'open_id' => 'openId',
            'nickname' => 'nickname',
            'avatar_url' => 'avatarUrl',
            'status' => 'status',
            'last_login_at' => 'lastLoginAt',
            'created_at' => 'createdAt',
            'current_level' => 'currentLevel',
            'coins' => 'coins',
            'stamina' => 'stamina',
            'max_stamina' => 'maxStamina',
            'hints' => 'hints',
            'completed_count' => 'completedCount',
            'total_games' => 'totalGames',
            'successful_games' => 'successfulGames',
        ];
        $items = $query
            ->select(
                'p.*',
                'g.current_level',
                'g.coins',
                'g.stamina',
                'g.max_stamina',
                'g.hints',
                DB::raw('JSON_LENGTH(g.completed_levels) as completed_count'),
                DB::raw('(SELECT COUNT(*) FROM level_results lr WHERE lr.player_id=p.id) as total_games'),
                DB::raw('(SELECT COUNT(*) FROM level_results lr WHERE lr.player_id=p.id AND lr.success=1) as successful_games'),
            )
            ->orderByRaw('COALESCE(p.last_login_at, p.created_at) DESC')
            ->orderByDesc('p.id')
            ->offset(($page - 1) * $size)
            ->limit($size)
            ->get()
            ->map(fn ($row) => $this->map($row, $map));

        return response()->json([
            'players' => $items,
            'total' => $total,
            'page' => $page,
            'pageSize' => $size,
        ]);
    }

    public function player(int $id): JsonResponse
    {
        $player = DB::table('players as p')
            ->leftJoin('player_progress as g', 'g.player_id', '=', 'p.id')
            ->where('p.id', $id)
            ->first();

        if (! $player) {
            return response()->json(['error' => '用户不存在'], 404);
        }
        $list = $this->map($player, [
            'id' => 'id', 'open_id' => 'openId', 'nickname' => 'nickname', 'avatar_url' => 'avatarUrl',
            'status' => 'status', 'last_login_at' => 'lastLoginAt', 'created_at' => 'createdAt',
            'current_level' => 'currentLevel', 'coins' => 'coins', 'stamina' => 'stamina',
            'max_stamina' => 'maxStamina', 'hints' => 'hints',
        ]);
        $stars = json_decode($player->level_stars ?? '{}', true) ?? [];
        $completed = json_decode($player->completed_levels ?? '[]', true) ?? [];
        $list['completedCount'] = count($completed);
        $list['totalGames'] = DB::table('level_results')->where('player_id', $id)->count();
        $list['successfulGames'] = DB::table('level_results')->where('player_id', $id)->where('success', 1)->count();
        $progress = [
            'currentLevel' => (int) ($player->current_level ?? 1), 'coins' => (int) ($player->coins ?? 0),
            'stamina' => (int) ($player->stamina ?? 0), 'maxStamina' => (int) ($player->max_stamina ?? 5),
            'hints' => (int) ($player->hints ?? 0), 'previewAgainCount' => (int) ($player->preview_again_count ?? 0),
            'removePairCount' => (int) ($player->remove_pair_count ?? 0), 'levelStars' => $stars,
            'completedLevels' => $completed, 'nextStaminaRecoverAt' => $player->next_stamina_recover_at ?? null,
            'updatedAt' => $player->updated_at ?? null,
        ];
        $recent = DB::table('level_results')->where('player_id', $id)
            ->orderByDesc('completed_at')->orderByDesc('id')->limit(20)->get()->map(fn ($x) => $this->map($x, [
                'id' => 'id', 'level_id' => 'levelId', 'success' => 'success', 'reason' => 'reason',
                'steps' => 'steps', 'mismatch_count' => 'mismatchCount', 'elapsed_ms' => 'elapsedMs',
                'stars' => 'stars', 'coins_earned' => 'coinsEarned', 'completed_at' => 'completedAt',
            ]));

        return response()->json(['player' => $list, 'progress' => $progress, 'recentResults' => $recent]);
    }

    public function index(Request $request): JsonResponse
    {
        if ($error = $this->owner($request)) {
            return $error;
        }
        $users = $request->is('api/admin/users*');
        $rows = DB::table($users ? 'admin_users' : 'system_controls')
            ->when($users, fn ($q) => $q->orderBy('id'))
            ->when(! $users, fn ($q) => $q->orderBy('control_group')->orderBy('sort_order')->orderBy('id'))
            ->get()
            ->map(
                fn ($row) => $this->map(
                    $row,
                    $users ? self::USER_MAP : self::CONTROL_MAP,
                ),
            );

        return response()->json([$users ? 'admins' : 'controls' => $rows]);
    }

    public function show(Request $request, int $id): JsonResponse
    {
        if ($error = $this->owner($request)) {
            return $error;
        }
        $users = $request->is('api/admin/users*');
        $row = DB::table($users ? 'admin_users' : 'system_controls')
            ->where('id', $id)
            ->first();

        return $row
            ? response()->json(
                $this->map($row, $users ? self::USER_MAP : self::CONTROL_MAP),
            )
            : response()->json(['error' => 'record not found'], 404);
    }

    public function store(Request $request): JsonResponse
    {
        if ($error = $this->owner($request)) {
            return $error;
        }

        return $request->is('api/admin/users*')
            ? $this->saveUser($request)
            : $this->saveControl($request);
    }

    public function update(Request $request, int $id): JsonResponse
    {
        if ($error = $this->owner($request)) {
            return $error;
        }

        return $request->is('api/admin/users*')
            ? $this->saveUser($request, $id)
            : $this->saveControl($request, $id);
    }

    public function destroy(Request $request, int $id): JsonResponse
    {
        if ($error = $this->owner($request)) {
            return $error;
        }
        $table = $request->is('api/admin/users*')
            ? 'admin_users'
            : 'system_controls';
        if (
            $table === 'admin_users' &&
            (int) $request->attributes->get('admin')['sub'] === $id
        ) {
            return response()->json(
                ['error' => 'cannot delete current account'],
                400,
            );
        }

        return DB::transaction(function () use ($table, $id) {
            if ($table === 'admin_users') {
                DB::table('admin_users')->where('created_by', $id)->update(['created_by' => null]);
            }

            return DB::table($table)->where('id', $id)->delete();
        })
            ? response()->noContent()
            : response()->json(['error' => 'record not found'], 404);
    }

    private function saveUser(Request $request, ?int $id = null): JsonResponse
    {
        $data = $request->validate([
            'username' => 'required|string|max:64',
            'email' => 'nullable|email',
            'displayName' => 'required|string|max:64',
            'role' => 'required|in:owner,operator,viewer',
            'status' => 'required|in:active,disabled',
            'permissions' => 'nullable|array',
            'password' => ($id ? 'nullable' : 'required').'|string|min:10',
        ]);
        $row = [
            'username' => $data['username'],
            'email' => $data['email'] ?? null,
            'display_name' => $data['displayName'],
            'role' => $data['role'],
            'status' => $data['status'],
            'permissions' => json_encode($data['permissions'] ?? (object) []),
            'updated_at' => now(),
        ];
        if (! empty($data['password'])) {
            $row['password_hash'] = Hash::make($data['password']);
        }
        if ($id) {
            DB::table('admin_users')->where('id', $id)->update($row);
            $code = 200;
        } else {
            $row['created_by'] = $request->attributes->get('admin')['sub'];
            $row['created_at'] = now();
            $id = DB::table('admin_users')->insertGetId($row);
            $code = 201;
        }

        return response()->json(
            $this->map(
                DB::table('admin_users')->where('id', $id)->first(),
                self::USER_MAP,
            ),
            $code,
        );
    }

    private function saveControl(
        Request $request,
        ?int $id = null,
    ): JsonResponse {
        $data = $request->validate([
            'controlKey' => 'required|string|max:128',
            'controlGroup' => 'required|string|max:64',
            'name' => 'required|string|max:128',
            'description' => 'nullable|string|max:512',
            'valueType' => 'required|in:bool,int,decimal,string,json',
            'valueText' => 'nullable|string',
            'valueJson' => 'nullable',
            'defaultValueText' => 'nullable|string',
            'defaultValueJson' => 'nullable',
            'enabled' => 'boolean',
            'isPublic' => 'boolean',
            'sortOrder' => 'integer',
            'effectiveFrom' => 'nullable|date',
            'effectiveUntil' => 'nullable|date',
        ]);
        $json = $data['valueType'] === 'json';
        if ($json && ! array_key_exists('valueJson', $data)) {
            return response()->json(['error' => 'valueJson is required for json type'], 400);
        }
        if (! $json && ! array_key_exists('valueText', $data)) {
            return response()->json(['error' => 'valueText is required'], 400);
        }
        if ($data['valueType'] === 'bool' && ! in_array($data['valueText'] ?? null, ['true', 'false'], true)) {
            return response()->json(['error' => 'bool value must be true or false'], 400);
        }
        if ($data['valueType'] === 'int' && filter_var($data['valueText'], FILTER_VALIDATE_INT) === false) {
            return response()->json(['error' => 'invalid integer value'], 400);
        }
        if ($data['valueType'] === 'decimal' && ! is_numeric($data['valueText'])) {
            return response()->json(['error' => 'invalid decimal value'], 400);
        }
        if (! empty($data['effectiveFrom']) && ! empty($data['effectiveUntil'])
            && strtotime($data['effectiveFrom']) > strtotime($data['effectiveUntil'])) {
            return response()->json(['error' => 'effectiveFrom must be before effectiveUntil'], 400);
        }
        $row = [
            'control_key' => $data['controlKey'],
            'control_group' => $data['controlGroup'],
            'name' => $data['name'],
            'description' => $data['description'] ?? '',
            'value_type' => $data['valueType'],
            'value_text' => $json ? null : $data['valueText'],
            'value_json' => $json
                ? json_encode($data['valueJson'])
                : null,
            'default_value_text' => $json
                ? null
                : ($data['defaultValueText'] ?? null),
            'default_value_json' => $json
                ? (array_key_exists('defaultValueJson', $data) ? json_encode($data['defaultValueJson']) : null)
                : null,
            'enabled' => $data['enabled'] ?? true,
            'is_public' => $data['isPublic'] ?? false,
            'sort_order' => $data['sortOrder'] ?? 0,
            'effective_from' => $data['effectiveFrom'] ?? null,
            'effective_until' => $data['effectiveUntil'] ?? null,
            'updated_by' => $request->attributes->get('admin')['sub'],
            'updated_at' => now(),
        ];
        if ($id) {
            $row['version'] = DB::raw('version + 1');
            DB::table('system_controls')->where('id', $id)->update($row);
            $code = 200;
        } else {
            $row['created_by'] = $request->attributes->get('admin')['sub'];
            $row['created_at'] = now();
            $id = DB::table('system_controls')->insertGetId($row);
            $code = 201;
        }

        return response()->json(
            $this->map(
                DB::table('system_controls')->where('id', $id)->first(),
                self::CONTROL_MAP,
            ),
            $code,
        );
    }
}
