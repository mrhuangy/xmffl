<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class Cors
{
    public function handle(Request $request, Closure $next): Response
    {
        $headers = [
            'Access-Control-Allow-Origin' => env('ALLOW_ORIGIN', '*'),
            'Access-Control-Allow-Headers' => 'Content-Type, Authorization, X-Openid',
            'Access-Control-Allow-Methods' => 'GET, POST, PUT, DELETE, OPTIONS',
        ];
        if ($request->isMethod('OPTIONS')) {
            return response('', 204, $headers);
        }
        $response = $next($request);
        foreach ($headers as $key => $value) {
            $response->headers->set($key, $value);
        }

        return $response;
    }
}
