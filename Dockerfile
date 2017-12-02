FROM thanhpk/alpine-nginx-git:1.0.0

COPY git.conf /etc/nginx/conf.d/
RUN mkdir /app
RUN mkdir /srv/git
WORKDIR /app
COPY start.sh /app/

COPY config.yaml /app/config.yaml
COPY sfgit /app/

CMD ["/app/start.sh"]
