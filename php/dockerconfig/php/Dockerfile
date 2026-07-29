FROM php:8.2-fpm-alpine

LABEL maintainer="dietitian-team"
LABEL description="PHP 8.2 FPM for Laravel 10"

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tencent.com/g' /etc/apk/repositories

RUN apk add --no-cache \
    bash git curl unzip su-exec \
    libpng-dev libjpeg-turbo-dev libwebp-dev freetype-dev \
    libzip-dev oniguruma-dev icu-dev

RUN apk add --no-cache --virtual .build-deps $PHPIZE_DEPS linux-headers \
    && docker-php-ext-configure gd --with-freetype --with-jpeg --with-webp \
    && docker-php-ext-install -j$(nproc) \
        pdo_mysql mbstring exif pcntl bcmath gd zip intl opcache \
    && apk del .build-deps

ENV PHPREDIS_VERSION=6.0.2
RUN apk add --no-cache --virtual .redis-build-deps $PHPIZE_DEPS linux-headers \
    && curl -sSL "https://gitee.com/mirrors/phpredis/repository/archive/${PHPREDIS_VERSION}.tar.gz" \
        -o /tmp/phpredis.tar.gz \
    && mkdir -p /tmp/phpredis \
    && tar -xzf /tmp/phpredis.tar.gz -C /tmp/phpredis --strip-components=1 \
    && cd /tmp/phpredis \
    && phpize && ./configure && make -j$(nproc) && make install \
    && docker-php-ext-enable redis \
    && apk del .redis-build-deps \
    && rm -rf /tmp/phpredis /tmp/phpredis.tar.gz

RUN curl -sSL https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer \
    && composer --version

RUN addgroup -g 1000 -S www \
    && adduser -u 1000 -S www -G www \
    && mkdir -p /var/www/admin-api /var/www/minigame-api /var/log/php-fpm \
    && chown -R www:www /var/www /var/log/php-fpm

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh \
    && chown www:www /usr/local/bin/docker-entrypoint.sh

WORKDIR /var/www/admin-api

EXPOSE 9000

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["php-fpm"]
