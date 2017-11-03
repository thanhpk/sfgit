#!/bin/sh

CGO_ENABLED=0 go build -i . &&
		docker build -qt thanhpk/sfgit . &&
		docker rm -f sfgittest &&
		docker run --name sfgittest -v /srv/sfgittest/git:/srv/git -it thanhpk/sfgit
