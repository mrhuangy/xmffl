<?php

namespace App\Http\Middleware;

use App\Support\Jwt;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class AdminAuth
{
    public function handle(Request $request, Closure $next): Response
    {
        try {
            $claims = Jwt::verify($request->bearerToken() ?? '');
        } catch (\Throwable) {
            return response()->json(
                ['error' => '登录状态已失效，请重新登录'],
                401,
            );
        }
        $request->attributes->set('admin', $claims);

        return $next($request);
    }
}
