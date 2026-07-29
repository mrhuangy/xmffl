<?php

namespace Tests\Feature;

use App\Services\PlayerRegistrationService;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Tests\TestCase;

class PlayerRegistrationServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        config()->set('database.connections.registration_test', [
            'driver' => 'sqlite',
            'database' => ':memory:',
            'prefix' => '',
            'foreign_key_constraints' => true,
        ]);
        DB::purge('registration_test');
        DB::setDefaultConnection('registration_test');

        Schema::create('players', function (Blueprint $table) {
            $table->id();
            $table->string('open_id')->unique();
            $table->string('union_id')->nullable();
            $table->string('nickname');
            $table->string('avatar_url')->default('');
            $table->string('status')->default('active');
            $table->dateTime('last_login_at')->nullable();
            $table->timestamps();
        });
        Schema::create('player_progress', function (Blueprint $table) {
            $table->unsignedBigInteger('player_id')->primary();
            $table->unsignedInteger('current_level');
            $table->integer('coins');
            $table->integer('stamina');
            $table->unsignedInteger('max_stamina');
            $table->integer('hints');
            $table->integer('preview_again_count');
            $table->integer('remove_pair_count');
            $table->json('level_stars');
            $table->json('completed_levels');
            $table->dateTime('next_stamina_recover_at')->nullable();
            $table->timestamps();
        });
        $this->createTransactionTable('coin_transactions', 'transaction_no', function (Blueprint $table) {
            $table->integer('change_amount');
            $table->integer('balance_after');
            $table->string('reason');
            $table->string('ref_type');
            $table->string('ref_id');
            $table->string('note')->default('');
        });
        $this->createTransactionTable('stamina_transactions', 'transaction_no', function (Blueprint $table) {
            $table->integer('change_amount');
            $table->integer('balance_after');
            $table->string('reason');
            $table->string('ref_type');
            $table->string('ref_id');
            $table->string('note')->default('');
        });
        $this->createTransactionTable('tool_transactions', 'transaction_no', function (Blueprint $table) {
            $table->string('tool_type');
            $table->integer('change_amount');
            $table->integer('balance_after');
            $table->string('source');
            $table->string('ref_type');
            $table->string('ref_id');
            $table->string('note')->default('');
        });
        Schema::create('reward_grants', function (Blueprint $table) {
            $table->id();
            $table->string('reward_id')->unique();
            $table->unsignedBigInteger('player_id');
            $table->string('source');
            $table->string('source_ref');
            $table->string('reward_type');
            $table->integer('amount');
        });
    }

    public function test_registration_matches_go_initialization_and_is_idempotent(): void
    {
        $service = app(PlayerRegistrationService::class);

        [$player, $progress] = $service->upsert('openid-1', 'union-1', 'avatar-1');

        $this->assertMatchesRegularExpression('/^.{3}$/u', $player->nickname);
        $this->assertSame(100, $progress->coins);
        $this->assertSame(3, $progress->hints);
        $this->assertSame(3, $progress->preview_again_count);
        $this->assertSame(3, $progress->remove_pair_count);
        $this->assertSame(1, DB::table('coin_transactions')->count());
        $this->assertSame(1, DB::table('reward_grants')->count());
        $this->assertSame(3, DB::table('tool_transactions')->count());

        [$secondPlayer] = $service->upsert('openid-1', null, '');

        $this->assertSame($player->nickname, $secondPlayer->nickname);
        $this->assertSame('union-1', $secondPlayer->union_id);
        $this->assertSame('avatar-1', $secondPlayer->avatar_url);
        $this->assertSame(1, DB::table('coin_transactions')->count());
        $this->assertSame(3, DB::table('tool_transactions')->count());
    }

    public function test_login_settles_natural_stamina_recovery(): void
    {
        $service = app(PlayerRegistrationService::class);
        [$player] = $service->upsert('openid-2', null, '');
        DB::table('player_progress')->where('player_id', $player->id)->update([
            'stamina' => 2,
            'next_stamina_recover_at' => now()->subMinutes(5),
        ]);

        [, $progress] = $service->upsert('openid-2', null, '');

        $this->assertSame(5, $progress->stamina);
        $this->assertNull($progress->next_stamina_recover_at);
        $this->assertDatabaseHas('stamina_transactions', [
            'player_id' => $player->id,
            'change_amount' => 3,
            'balance_after' => 5,
            'reason' => 'auto_recover',
            'ref_type' => 'system',
            'ref_id' => 'natural_recovery',
        ]);
    }

    private function createTransactionTable(string $name, string $numberColumn, callable $columns): void
    {
        Schema::create($name, function (Blueprint $table) use ($numberColumn, $columns) {
            $table->id();
            $table->string($numberColumn)->unique();
            $table->unsignedBigInteger('player_id');
            $columns($table);
            $table->timestamp('created_at')->useCurrent();
        });
    }
}
