FROM php:8.3-apache
RUN apt-get update \
    && apt-get install -y --no-install-recommends libpq-dev unzip \
    && docker-php-ext-install pdo_pgsql opcache \
    && rm -rf /var/lib/apt/lists/* \
    && mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini"
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer
ENV APP_ENV=prod APP_DEBUG=0 COMPOSER_ALLOW_SUPERUSER=1 PORT=8080
WORKDIR /var/www/app
COPY composer.json composer.lock ./
RUN composer install --no-dev --no-scripts --no-autoloader --no-interaction --prefer-dist
COPY . .
RUN composer dump-autoload --no-dev --optimize --classmap-authoritative \
    && rm -rf var && mkdir var && chown www-data:www-data var
RUN sed -ri 's!/var/www/html!/var/www/app/public!g' /etc/apache2/sites-available/*.conf \
    && sed -ri 's!^Listen 80$!Listen 8080!' /etc/apache2/ports.conf \
    && sed -ri 's!<VirtualHost \*:80>!<VirtualHost *:8080>!' /etc/apache2/sites-available/000-default.conf \
    && printf '<Directory /var/www/app/public>\n    AllowOverride None\n    Require all granted\n    FallbackResource /index.php\n</Directory>\nPassEnv APP_ENV APP_DEBUG DATABASE_URL\n' > /etc/apache2/conf-enabled/app.conf
EXPOSE 8080
