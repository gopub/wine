package main

import (
	"flag"

	"github.com/gopub/log/v2"
	"github.com/gopub/wine/exp/vfs"
)

func main() {
	pAddr := flag.String("addr", ":9000", "server address")
	flag.Parse()
	fs, err := vfs.NewFileSystem(vfs.NewMemoryStorage())
	if err != nil {
		log.Fatal(err)
	}
	fs.Handler().RunServer(*pAddr)
}
