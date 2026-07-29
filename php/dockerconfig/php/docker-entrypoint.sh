#!/bin/bash
set -e

prepare_laravel_dir() {
    local dir="$1"
    if [ -d "${dir}" ]; then
        mkdir -p "${dir}/bootstrap/cache" "${dir}/storage/framework/cache" \
            "${dir}/storage/framework/sessions" "${dir}/storage/framework/views" \
            "${dir}/storage/logs" "${dir}/vendor"
        chown -R www:www "${dir}/bootstrap/cache" "${dir}/storage" "${dir}/vendor"
    fi
}

composer_install() {
    local dir="$1"
    if [ -f "${dir}/composer.json" ] \
        && { [ ! -f "${dir}/vendor/autoload.php" ] || [ "${dir}/composer.lock" -nt "${dir}/vendor/autoload.php" ]; }; then
        echo "[entrypoint] composer install -> ${dir}"
        cd "${dir}"
        su-exec www composer install --no-interaction --no-progress --prefer-dist --optimize-autoloader
    fi
}

prepare_laravel_dir /var/www/admin-api
prepare_laravel_dir /var/www/minigame-api

composer_install /var/www/admin-api
composer_install /var/www/minigame-api

if [ "${1#-}" != "$1" ]; then
    set -- php-fpm "$@"
fi

if [ "$1" = "php-fpm" ] || [ "$1" = "php-fpm8" ]; then
    exec "$@"
fi

exec su-exec www "$@"
