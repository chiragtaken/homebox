FROM tobi312/rpi-nginx 

MAINTAINER Chirag Tayal <chiragtayal@gmail.com>

COPY default /etc/nginx/sites-available

COPY html /var/www

RUN mkdir -p /etc/nginx/ssl
COPY ssl/nginx.key /etc/nginx/ssl
COPY ssl/nginx.crt /etc/nginx/ssl

COPY hosts /etc

