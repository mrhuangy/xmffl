<?php

namespace Tests\Feature;

// use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class ExampleTest extends TestCase
{
    public function test_health_endpoint_returns_ok(): void
    {
        $this->getJson('/healthz')
            ->assertOk()
            ->assertExactJson(['status' => 'ok']);
    }
}
