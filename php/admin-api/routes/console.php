<?php

use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

Artisan::command(
    'admin:create {username} {password} {--name=Owner}',
    function () {
        DB::table('admin_users')->updateOrInsert(
            ['username' => $this->argument('username')],
            [
                'password_hash' => Hash::make($this->argument('password')),
                'display_name' => $this->option('name'),
                'role' => 'owner',
                'status' => 'active',
                'permissions' => json_encode((object) []),
                'updated_at' => now(),
                'created_at' => now(),
            ],
        );
        $this->info('Owner account is ready.');
    },
)->purpose('Create or reset an owner account');
