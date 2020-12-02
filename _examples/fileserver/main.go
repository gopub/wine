package main

import (
	"flag"
	"github.com/gopub/log"
	"github.com/gopub/wine"
)

func main() {
	pDir := flag.String("dir", ".", "directory")
	pAddr := flag.String("addr", ":8000", "server address")
	flag.Parse()
	s := wine.NewServer()
	s.StaticDir("/", *pDir)
	log.Infof("File directory: %s", *pDir)
	s.Run(*pAddr)
}
