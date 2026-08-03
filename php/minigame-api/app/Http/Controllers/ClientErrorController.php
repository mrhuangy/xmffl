<?php

namespace App\Http\Controllers;

use App\Services\ClientErrorLogger;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class ClientErrorController extends Controller
{
    public function store(Request $request, ClientErrorLogger $logger): JsonResponse
    {
        $data = $request->validate([
            'stage' => 'required|string|max:64',
            'message' => 'required|string|max:500',
            'errMsg' => 'nullable|string|max:500',
            'statusCode' => 'nullable|integer|min:0|max:599',
            'requestUrl' => 'nullable|string|max:255',
            'occurredAt' => 'nullable|integer|min:0',
            'device' => 'nullable|array|max:12',
            'device.platform' => 'nullable|string|max:32',
            'device.model' => 'nullable|string|max:100',
            'device.system' => 'nullable|string|max:100',
            'device.version' => 'nullable|string|max:32',
            'device.SDKVersion' => 'nullable|string|max:32',
            'device.brand' => 'nullable|string|max:50',
            'device.language' => 'nullable|string|max:32',
        ]);
        $requestId = $logger->write($data['stage'], [
            'source' => 'wechat_minigame',
            'message' => $data['message'],
            'err_msg' => $data['errMsg'] ?? null,
            'status_code' => $data['statusCode'] ?? null,
            'request_url' => $data['requestUrl'] ?? null,
            'occurred_at' => $data['occurredAt'] ?? null,
            'device' => $data['device'] ?? null,
            'ip' => $request->ip(),
            'user_agent' => mb_substr((string) $request->userAgent(), 0, 255),
        ]);

        return response()->json(['status' => 'accepted', 'requestId' => $requestId], 202);
    }
}
