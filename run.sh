#!/bin/sh
set -e
go build -o ./fox-torrent ./cmd
exec ./fox-torrent ./testfiles/ubuntu-23.10-live-server-amd64.iso.torrent
