<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Symfony\Component\HttpFoundation\Response;

class PlayerAuth
{
    public function handle(Request $r, Closure $next): Response
    {
        $token = $r->bearerToken() ?: $r->header('X-Openid');
        if (! $token) {
            return response()->json(['error' => 'missing token'], 401);
        }
        $p = DB::table('players')
            ->where('open_id', $token)
            ->where('status', '!=', 'deleted')
            ->first();
        if (! $p) {
            return response()->json(['error' => 'invalid token'], 401);
        }
        $r->attributes->set('player', $p);

        return $next($r);
    }
}
