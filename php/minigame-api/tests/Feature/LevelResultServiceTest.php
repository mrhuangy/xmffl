<?php

namespace Tests\Feature;

use App\Services\LevelResultService;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Tests\TestCase;

class LevelResultServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        config()->set('database.connections.level_result_test', [
            'driver' => 'sqlite',
            'database' => ':memory:',
            'prefix' => '',
            'foreign_key_constraints' => true,
        ]);
        DB::purge('level_result_test');
        DB::setDefaultConnection('level_result_test');
        $this->createSchema();
        $this->seedState();
    }

    public function test_first_three_star_completion_grants_stamina_and_saves_all_records(): void
    {
        [$progress, $rewards] = app(LevelResultService::class)->submit($this->player(), $this->resultData());

        $this->assertSame(['stamina' => 1], $rewards);
        $this->assertSame(5, $progress->stamina);
        $this->assertSame(130, $progress->coins);
        $this->assertSame(2, $progress->current_level);
        $this->assertSame(['1' => 3], json_decode($progress->level_stars, true));
        $this->assertSame([1], json_decode($progress->completed_levels, true));
        $this->assertNull($progress->next_stamina_recover_at);
        $this->assertDatabaseHas('level_results', ['success' => 1, 'stars' => 3, 'coins_earned' => 30]);
        $this->assertDatabaseHas('coin_transactions', ['reason' => 'level_complete', 'change_amount' => 30, 'balance_after' => 130]);
        $this->assertDatabaseHas('stamina_transactions', ['reason' => 'activity_reward', 'change_amount' => 1, 'balance_after' => 5, 'note' => 'first_3_star']);
        $this->assertSame(2, DB::table('reward_grants')->count());
        $this->assertSame(1, DB::table('leaderboard_entries')->count());

        [, $secondRewards] = app(LevelResultService::class)->submit($this->player(), $this->resultData());

        $this->assertSame(['stamina' => 0], $secondRewards);
        $this->assertSame(1, DB::table('stamina_transactions')->count());
        $this->assertSame(3, DB::table('reward_grants')->count());
    }

    public function test_non_completed_reason_is_saved_as_failure_without_changing_progress(): void
    {
        $result = $this->resultData();
        $result['reason'] = 'quit';
        $result['levelId'] = 2;

        [$progress, $rewards] = app(LevelResultService::class)->submit($this->player(), $result);

        $this->assertSame(['stamina' => 0], $rewards);
        $this->assertSame(1, $progress->current_level);
        $this->assertSame(100, $progress->coins);
        $this->assertDatabaseHas('level_results', ['level_id' => 2, 'success' => 0, 'stars' => 0, 'coins_earned' => 0]);
        $this->assertSame(0, DB::table('leaderboard_entries')->count());
        $this->assertSame(0, DB::table('reward_grants')->count());
    }

    /** @return array<string, mixed> */
    private function resultData(): array
    {
        return ['levelId' => 1, 'success' => true, 'reason' => 'completed', 'steps' => 10, 'mismatchCount' => 1, 'elapsedMs' => 29_001, 'usedHints' => 0];
    }

    private function player(): object
    {
        return (object) ['id' => 1, 'nickname' => '赵安宁'];
    }

    private function createSchema(): void
    {
        Schema::create('player_progress', function (Blueprint $table) {
            $table->unsignedBigInteger('player_id')->primary();
            $table->unsignedInteger('current_level');
            $table->integer('coins');
            $table->integer('stamina');
            $table->unsignedInteger('max_stamina');
            $table->json('level_stars');
            $table->json('completed_levels');
            $table->dateTime('next_stamina_recover_at')->nullable();
            $table->timestamps();
        });
        Schema::create('level_configs', function (Blueprint $table) {
            $table->id();
            $table->unsignedInteger('level_id')->unique();
            $table->boolean('enabled');
            $table->integer('excellent_step_threshold');
            $table->integer('normal_step_threshold');
            $table->integer('excellent_time_threshold')->nullable();
            $table->integer('normal_time_threshold')->nullable();
            $table->integer('coin_reward_base');
            $table->integer('coin_reward_star1');
            $table->integer('coin_reward_star2');
            $table->integer('coin_reward_star3');
        });
        Schema::create('level_results', function (Blueprint $table) {
            $table->id();
            $table->unsignedBigInteger('player_id');
            $table->unsignedInteger('level_id');
            $table->boolean('success');
            $table->string('reason');
            $table->integer('steps');
            $table->integer('mismatch_count');
            $table->integer('elapsed_ms');
            $table->integer('stars');
            $table->integer('coins_earned');
            $table->integer('used_hints');
            $table->dateTime('completed_at');
        });
        foreach (['coin_transactions', 'stamina_transactions'] as $name) {
            Schema::create($name, function (Blueprint $table) {
                $table->id();
                $table->string('transaction_no')->unique();
                $table->unsignedBigInteger('player_id');
                $table->integer('change_amount');
                $table->integer('balance_after');
                $table->string('reason');
                $table->string('ref_type');
                $table->string('ref_id');
                $table->string('note')->default('');
            });
        }
        Schema::create('reward_grants', function (Blueprint $table) {
            $table->id();
            $table->string('reward_id')->unique();
            $table->unsignedBigInteger('player_id');
            $table->string('source');
            $table->string('source_ref');
            $table->string('reward_type');
            $table->integer('amount');
            $table->unsignedInteger('level_id')->nullable();
        });
        Schema::create('leaderboard_entries', function (Blueprint $table) {
            $table->id();
            $table->unsignedBigInteger('player_id');
            $table->unsignedInteger('level_id');
            $table->string('nickname');
            $table->integer('stars');
            $table->integer('steps');
            $table->integer('elapsed_ms');
            $table->dateTime('submitted_at');
        });
    }

    private function seedState(): void
    {
        DB::table('player_progress')->insert([
            'player_id' => 1, 'current_level' => 1, 'coins' => 100, 'stamina' => 4, 'max_stamina' => 5,
            'level_stars' => '{}', 'completed_levels' => '[]', 'next_stamina_recover_at' => now()->addMinute(),
            'created_at' => now(), 'updated_at' => now(),
        ]);
        foreach ([1, 2] as $levelId) {
            DB::table('level_configs')->insert([
                'level_id' => $levelId, 'enabled' => true, 'excellent_step_threshold' => 10,
                'normal_step_threshold' => 20, 'excellent_time_threshold' => 30, 'normal_time_threshold' => 60,
                'coin_reward_base' => 10, 'coin_reward_star1' => 10, 'coin_reward_star2' => 20, 'coin_reward_star3' => 30,
            ]);
        }
    }
}
