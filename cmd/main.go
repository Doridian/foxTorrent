package main

import (
	"log"

	"github.com/Doridian/foxTorrent/bencoding"
)

func main() {
	log.Printf("Hello")

	testStr := "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee"
	log.Printf("testStr: %v", testStr)
	res, err := bencoding.Decode(testStr)
	log.Printf("err: %v ; res: %v", err, res)
}
