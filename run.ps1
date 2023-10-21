Remove-Item ./fox-torrent.exe
go build -o ./fox-torrent.exe ./cmd
./fox-torrent.exe testfiles/ubuntu-23.10-live-server-amd64.iso.torrent
