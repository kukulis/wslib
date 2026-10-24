FROM php:8.3-fpm-bookworm

WORKDIR /var/www/html

RUN apt-get update
RUN apt-get -y install less vim git iputils-ping zip libzip-dev libpng-dev libxml2-dev mariadb-client

RUN #docker-php-ext-install zip gd bcmath exif gd intl opcache pcntl mysqli pdo_mysql
RUN docker-php-ext-install zip exif intl pcntl mysqli pdo_mysql

RUN pecl install xdebug
RUN docker-php-ext-enable xdebug


RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - && apt-get install -y nodejs

RUN addgroup --gid=1000 app
RUN addgroup --gid=1002 developer

RUN #adduser --disabled-password --quiet --uid=1000 --gid=1000 --no-create-home --gecos GECOS app
RUN adduser --disabled-password --quiet --uid=1000 --gid=1000 --gecos GECOS app