<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class Cors
{
    public function handle(Request $r, Closure $next): Response
    {
        $h = [
            'Access-Control-Allow-Origin' => env('ALLOW_ORIGIN', '*'),
            'Access-Control-Allow-Headers' => 'Content-Type, Authorization, X-Openid',
            'Access-Control-Allow-Methods' => 'GET, POST, OPTIONS',
        ];
        if ($r->isMethod('OPTIONS')) {
            return response('', 204, $h);
        }
        $x = $next($r);
        foreach ($h as $k => $v) {
            $x->headers->set($k, $v);
        }

        return $x;
    }
}
