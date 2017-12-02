#!/bin/sh

CGO_ENABLED=0 go build -i . &&
		docker build -t thanhpk/sfgit . &&
		(docker rm -f sfgit || true) &&
		docker volume create sfgit &&
		docker run --name sfgit -v sfgit:/srv/git -p 10292:80 -it thanhpk/sfgit
