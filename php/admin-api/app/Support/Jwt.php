<?php

namespace App\Support;

use RuntimeException;

final class Jwt
{
    private static function enc(string $value): string
    {
        return rtrim(strtr(base64_encode($value), '+/', '-_'), '=');
    }

    private static function dec(string $value): string|false
    {
        return base64_decode(strtr($value, '-_', '+/'), true);
    }

    public static function sign(array $claims): string
    {
        $head = self::enc(
            json_encode(
                ['alg' => 'HS256', 'typ' => 'JWT'],
                JSON_UNESCAPED_SLASHES,
            ),
        );
        $body = self::enc(
            json_encode(
                $claims,
                JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES,
            ),
        );

        return "$head.$body.".
            self::enc(
                hash_hmac(
                    'sha256',
                    "$head.$body",
                    env('JWT_SECRET', 'change-me'),
                    true,
                ),
            );
    }

    public static function verify(string $token): array
    {
        $parts = explode('.', $token);
        if (count($parts) !== 3) {
            throw new RuntimeException('invalid token');
        }
        $expected = self::enc(
            hash_hmac(
                'sha256',
                "$parts[0].$parts[1]",
                env('JWT_SECRET', 'change-me'),
                true,
            ),
        );
        if (! hash_equals($expected, $parts[2])) {
            throw new RuntimeException('invalid signature');
        }
        $claims = json_decode(self::dec($parts[1]) ?: '', true);
        if (! is_array($claims) || ($claims['exp'] ?? 0) <= time()) {
            throw new RuntimeException('expired token');
        }

        return $claims;
    }
}
