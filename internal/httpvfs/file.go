package httpvfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type inMemFile struct {
	name    string
	data    []byte
	offset  int64
	size    int64
	modTime time.Time
}

func (f *inMemFile) Close() error {
	return nil
}

func (f *inMemFile) Read(p []byte) (n int, err error) {
	defer func() {
		fmt.Println(n, err)
	}()
	if f.offset >= f.size {
		return 0, io.EOF
	}
	for i := range p {
		p[i] = f.data[f.offset]
		f.offset++
		n++
		if f.offset >= f.size {
			return n, io.EOF
		}
	}
	return n, nil
}

func (f *inMemFile) Seek(offset int64, whence int) (int64, error) {
	if offset >= f.size {
		return f.offset, io.EOF
	}
	switch whence {
	case io.SeekStart:
		f.offset = offset
	case io.SeekCurrent:
		offset += f.offset
		if offset >= f.size {
			return f.offset, io.EOF
		}
		f.offset = offset
	case io.SeekEnd:
		f.offset = f.size - offset - 1
	}
	return f.offset, nil
}

func (f *inMemFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *inMemFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *inMemFile) Name() string {
	return f.name
}

func (f *inMemFile) Size() int64 {
	return f.size
}

func (f *inMemFile) Mode() os.FileMode {
	return 0444
}

func (f *inMemFile) ModTime() time.Time {
	return f.modTime
}

func (f *inMemFile) IsDir() bool {
	return false
}

func (f *inMemFile) Sys() interface{} {
	return nil
}

var _ http.File = (*inMemFile)(nil)
var _ os.FileInfo = (*inMemFile)(nil)

func NewFile(name string, data []byte) http.File {
	return &inMemFile{
		name:    name,
		data:    data,
		offset:  0,
		size:    int64(len(data)),
		modTime: time.Now(),
	}
}
