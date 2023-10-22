#!/bin/sh
set -e
rm -f ./foxTorrent
go build -o ./foxTorrent ./cmd/foxTorrent
exec ./foxTorrent ./testfiles/ubuntu-23.10-live-server-amd64.iso.torrent
