package main

import (
	"flag"

	"github.com/gopub/log"
	"github.com/gopub/wine"
	"github.com/gopub/wine/exp/vfs"
)

func main() {
	pDir := flag.String("dir", ".", "directory")
	pAddr := flag.String("addr", ":9000", "server address")
	flag.Parse()
	s := wine.NewServer()
	fs, err := vfs.NewFileSystem(vfs.NewMemoryStorage())
	if err != nil {
		log.Fatal(err)
	}
	_, err = fs.Wrapper().ImportDiskFile("", *pDir)
	if err != nil {
		log.Fatal(err)
	}
	s.StaticFS("/", fs)
	s.Run(*pAddr)
}
