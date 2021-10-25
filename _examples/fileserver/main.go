package main

import (
	"flag"

	"github.com/gopub/log/v2"
	"github.com/gopub/wine"
)

func main() {
	pDir := flag.String("dir", ".", "directory")
	pAddr := flag.String("addr", ":8000", "server address")
	flag.Parse()
	s := wine.NewServer(nil)
	s.StaticDir("/", *pDir)
	log.Infof("BytesFile directory: %s", *pDir)
	s.Run(*pAddr)
}
