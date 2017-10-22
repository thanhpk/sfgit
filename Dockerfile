FROM nginx:1.13.5-alpine

LABEL maintainer="NGINX Docker Maintainers <docker-maint@nginx.com>"

# install git

RUN apk --update add git openssh fcgiwrap spawn-fcgi && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

COPY git.conf /etc/nginx/conf.d/
RUN mkdir /app
RUN mkdir /srv/git
WORKDIR /app
COPY sfgit /app/
COPY start.sh /app/

COPY config.yaml /app/config.yaml
CMD ["/app/start.sh"]
