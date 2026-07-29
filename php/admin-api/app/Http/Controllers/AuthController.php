<?php

namespace App\Http\Controllers;

use App\Support\Jwt;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class AuthController extends Controller
{
    public function login(Request $request): JsonResponse
    {
        $data = $request->validate(['username' => 'required|string', 'password' => 'required|string']);
        $admin = DB::table('admin_users')->where('username', trim($data['username']))->first();
        if (! $admin) {
            return response()->json(['error' => '用户名或密码错误'], 401);
        }
        if ($admin->status === 'disabled') {
            return response()->json(['error' => '账号已被禁用'], 403);
        }
        if ($admin->locked_until && now()->lt($admin->locked_until)) {
            return response()->json(['error' => '登录失败次数过多，请稍后再试'], 423);
        }
        if (! password_verify($data['password'], $admin->password_hash)) {
            $attempts = (int) $admin->failed_login_attempts + 1;
            DB::table('admin_users')->where('id', $admin->id)->update([
                'failed_login_attempts' => DB::raw('failed_login_attempts + 1'),
                'locked_until' => $attempts >= 5 ? now()->addMinutes(15) : null,
                'status' => $attempts >= 5 ? 'locked' : 'active',
            ]);

            return response()->json(['error' => '用户名或密码错误'], 401);
        }
        DB::table('admin_users')->where('id', $admin->id)->update([
            'failed_login_attempts' => 0, 'locked_until' => null, 'status' => 'active',
            'last_login_at' => now(), 'last_login_ip' => $request->ip(),
        ]);
        $claims = ['sub' => (int) $admin->id, 'username' => $admin->username,
            'displayName' => $admin->display_name, 'role' => $admin->role, 'exp' => time() + 28800];

        return response()->json(['token' => Jwt::sign($claims), 'expiresAt' => $claims['exp'], 'user' => $claims]);
    }

    public function me(Request $request): JsonResponse
    {
        return response()->json(['user' => $request->attributes->get('admin')]);
    }
}
