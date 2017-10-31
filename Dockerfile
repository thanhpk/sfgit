FROM nginx:1.13.5-alpine

LABEL maintainer="NGINX Docker Maintainers <docker-maint@nginx.com>"

# install git

RUN apk --update add git openssh fcgiwrap spawn-fcgi && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN ssh-keyscan bitbucket.org >> ~/.ssh/known_hosts

COPY git.conf /etc/nginx/conf.d/
RUN mkdir /app
RUN mkdir /srv/git
WORKDIR /app
COPY start.sh /app/

COPY config.yaml /app/config.yaml
COPY sfgit /app/

CMD ["/app/start.sh"]
