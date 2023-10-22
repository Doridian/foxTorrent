Remove-Item ./foxTorrent.exe
go build -o ./foxTorrent.exe ./cmd/foxTorrent
./foxTorrent.exe testfiles/ubuntu-23.10-live-server-amd64.iso.torrent
