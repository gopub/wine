package main

import (
	"flag"
	"net/http"

	"github.com/gopub/log"

	"github.com/gopub/wine"
	"github.com/gopub/wine/exp/storage"
)

func main() {
	pDir := flag.String("dir", ".", "directory")
	pAddr := flag.String("addr", ":8000", "server address")
	flag.Parse()
	s := wine.NewServer()
	s.Header().AllowOrigins("*")

	bucket, err := storage.NewDiskBucket(*pDir)
	if err != nil {
		log.Fatal(err)
	}
	s.Bind(http.MethodPost, "/upload", storage.NewFileWriter(bucket))
	s.Run(*pAddr)
}
