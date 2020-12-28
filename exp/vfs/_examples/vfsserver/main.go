package main

import (
	"flag"
	"net/url"
	"os"

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
	dir, err := os.Open(*pDir)
	if err != nil {
		log.Fatal(err)
	}
	names, err := dir.Readdirnames(100)
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range names {
		_, err = fs.Wrapper().CopyDiskFile("", name)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Debugf("http://127.0.0.1:9000/%s", url.PathEscape(name))
	}
	s.StaticFS("/", fs)
	s.Run(*pAddr)
}
