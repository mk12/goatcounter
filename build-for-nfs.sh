#!/bin/bash

set -Eeufo pipefail
trap 'echo >&2 "$0:$LINENO [$?]: $BASH_COMMAND"' ERR

version=$(git log -n1 --format='%h_%cI')

GOOS=freebsd GOARCH=amd64 CGO_ENABLED=1 \
	CC="zig cc -target x86_64-freebsd.14.0" \
	go build -ldflags="-s -w -X zgo.at/goatcounter/v2.Version=$version" \
	-trimpath -o goatcounter-nfs \
	./cmd/goatcounter

cat <<EOF
To deploy:

    scp goatcounter-nfs nfs:/home/protected/goatcounter/goatcounter
EOF
