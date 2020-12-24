package vfs

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/gopub/types"
)

type gobFileInfo struct {
	Name       string
	IsDir      bool
	MIMEType   string
	Pages      []string
	Size       int64
	Files      []*FileInfo
	CreatedAt  int64
	ModifiedAt int64
	Location   *types.Point
}

type FileInfo struct {
	gobFileInfo
	parent *FileInfo
}

var _ os.FileInfo = (*FileInfo)(nil)
var _ encoding.BinaryMarshaler = (*FileInfo)(nil)
var _ encoding.BinaryUnmarshaler = (*FileInfo)(nil)

func newDirInfo(name string) *FileInfo {
	return &FileInfo{
		gobFileInfo: gobFileInfo{
			Name:       name,
			IsDir:      true,
			CreatedAt:  time.Now().Unix(),
			ModifiedAt: time.Now().Unix(),
		},
	}
}

func newFileInfo(name, mimeType string, location *types.Point) *FileInfo {
	return &FileInfo{
		gobFileInfo: gobFileInfo{
			Name:       name,
			IsDir:      false,
			MIMEType:   mimeType,
			Location:   location,
			CreatedAt:  time.Now().Unix(),
			ModifiedAt: time.Now().Unix(),
		},
	}
}

func (f *FileInfo) Name() string {
	return f.gobFileInfo.Name
}

func (f *FileInfo) Size() int64 {
	return f.gobFileInfo.Size
}

func (f *FileInfo) Mode() os.FileMode {
	return 0400
}

func (f *FileInfo) ModTime() time.Time {
	return time.Unix(f.gobFileInfo.ModifiedAt, 0)
}

func (f *FileInfo) IsDir() bool {
	return f.gobFileInfo.IsDir
}

func (f *FileInfo) Sys() interface{} {
	return nil
}

func (f *FileInfo) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(&f.gobFileInfo)
}

func (f *FileInfo) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(f.gobFileInfo)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f *FileInfo) addPage(p string) {
	f.Pages = append(f.Pages, p)
	f.ModifiedAt = time.Now().Unix()
}

func (f *FileInfo) setSize(size int64) {
	f.gobFileInfo.Size = size
	f.ModifiedAt = time.Now().Unix()
}

func (f *FileInfo) GetByPath(path string) *FileInfo {
	return f.getByPathList(splitPath(path))
}

func (f *FileInfo) getByPathList(pathList []string) *FileInfo {
	if len(pathList) == 0 {
		return f
	}
	for _, fi := range f.Files {
		if fi.Name() == pathList[0] {
			return fi.getByPathList(pathList[1:])
		}
	}
	return nil
}

func (f *FileInfo) Get(name string) *FileInfo {
	for _, fi := range f.Files {
		if fi.Name() == name {
			return fi
		}
	}
	return nil
}

func (f *FileInfo) Exists(name string) bool {
	for _, fi := range f.Files {
		if fi.Name() == name {
			return true
		}
	}
	return false
}

func (f *FileInfo) DistinctName(name string) string {
	i := 0
	s := name
	for f.Exists(name) {
		i++
		s = fmt.Sprintf("%s-%d", name, i)
	}
	return s
}

func (f *FileInfo) AddSub(sub *FileInfo) {
	if sub.parent != nil && sub.parent != f {
		sub.parent.RemoveSub(sub)
	}
	sub.parent = f
	f.gobFileInfo.Files = append(f.gobFileInfo.Files, sub)
	f.ModifiedAt = time.Now().Unix()
}

func (f *FileInfo) RemoveSub(sub *FileInfo) {
	for i, fi := range f.Files {
		if fi == sub {
			f.Files = append(f.Files[:i], f.Files[i+1:]...)
			break
		}
	}
	f.ModifiedAt = time.Now().Unix()
}

// FileInfoList is defined for gomobile
type FileInfoList struct {
	files []*FileInfo
}

func (l *FileInfoList) Len() int {
	return len(l.files)
}

func (l *FileInfoList) Get(i int) *FileInfo {
	return l.files[i]
}

func (l *FileInfoList) Swap(i, j int) {
	l.files[i], l.files[j] = l.files[j], l.files[i]
}

func (l *FileInfoList) Less(i, j int) bool {
	return l.files[i].CreatedAt >= l.files[j].CreatedAt
}
