#!/bin/sh
set -e
go build -o ./fox-torrent ./cmd
exec ./fox-torrent "$@"
