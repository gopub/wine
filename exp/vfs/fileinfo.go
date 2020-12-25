package vfs

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gopub/log"
	"github.com/gopub/types"
)

type fileMetadata struct {
	UUID       string       `json:"uuid"`
	Name       string       `json:"name"`
	IsDir      bool         `json:"is_dir,omitempty"`
	MIMEType   string       `json:"mime_type,omitempty"`
	Pages      []string     `json:"pages,omitempty"`
	Size       int64        `json:"size,omitempty"`
	Duration   int          `json:"duration,omitempty"`
	Files      []*FileInfo  `json:"files,omitempty"`
	CreatedAt  int64        `json:"created_at"`
	ModifiedAt int64        `json:"modified_at"`
	Location   *types.Point `json:"location,omitempty"`
	Permission int          `json:"permission,omitempty"`
	Ext        types.M      `json:"ext,omitempty"`
}

type FileInfo struct {
	fileMetadata
	parent     *FileInfo
	busy       bool
	dirContent []byte
}

var _ os.FileInfo = (*FileInfo)(nil)
var _ encoding.BinaryMarshaler = (*FileInfo)(nil)
var _ encoding.BinaryUnmarshaler = (*FileInfo)(nil)
var _ encoding.TextMarshaler = (*FileInfo)(nil)
var _ encoding.TextUnmarshaler = (*FileInfo)(nil)

func newFileInfo(isDir bool, name string) *FileInfo {
	return &FileInfo{
		fileMetadata: fileMetadata{
			UUID:       uuid.New().String(),
			Name:       name,
			IsDir:      isDir,
			CreatedAt:  time.Now().Unix(),
			ModifiedAt: time.Now().Unix(),
		},
	}
}

func (f *FileInfo) Name() string {
	return f.fileMetadata.Name
}

func (f *FileInfo) Size() int64 {
	if f.IsDir() {
		return int64(len(f.dirContent))
	}
	return f.fileMetadata.Size
}

func (f *FileInfo) Mode() os.FileMode {
	return 0400
}

func (f *FileInfo) ModTime() time.Time {
	return time.Unix(f.fileMetadata.ModifiedAt, 0)
}

func (f *FileInfo) IsDir() bool {
	return f.fileMetadata.IsDir
}

func (f *FileInfo) Sys() interface{} {
	return nil
}

func (f *FileInfo) UnmarshalBinary(data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(&f.fileMetadata)
}

func (f *FileInfo) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer(nil)
	err = gob.NewEncoder(buf).Encode(f.fileMetadata)
	return buf.Bytes(), err
}

func (f *FileInfo) UnmarshalText(data []byte) error {
	return json.Unmarshal(data, &f.fileMetadata)
}

func (f *FileInfo) MarshalText() (data []byte, err error) {
	return json.Marshal(f.fileMetadata)
}

func (f *FileInfo) addPage(p string) {
	f.Pages = append(f.Pages, p)
	f.fileMetadata.ModifiedAt = time.Now().Unix()
}

func (f *FileInfo) truncate() {
	f.fileMetadata.Pages = f.fileMetadata.Pages[:]
	f.setSize(0)
}

func (f *FileInfo) setSize(size int64) {
	f.fileMetadata.Size = size
	f.fileMetadata.ModifiedAt = time.Now().Unix()
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

func (f *FileInfo) GetByUUID(id string) *FileInfo {
	for _, fi := range f.Files {
		if fi.fileMetadata.UUID == id {
			return fi
		}
		if found := fi.GetByUUID(id); found != nil {
			return found
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
	f.fileMetadata.Files = append(f.fileMetadata.Files, sub)
	f.fileMetadata.ModifiedAt = time.Now().Unix()
	f.dirContent = nil
	f.DirContent()
}

func (f *FileInfo) RemoveSub(sub *FileInfo) {
	for i, fi := range f.Files {
		if fi == sub {
			f.Files = append(f.Files[:i], f.Files[i+1:]...)
			break
		}
	}
	f.fileMetadata.ModifiedAt = time.Now().Unix()
	f.dirContent = nil
	f.DirContent()
}

func (f *FileInfo) Rename(name string) bool {
	name = strings.TrimSpace(name)
	if !validateFileName(name) {
		return false
	}
	if f.Name() == name {
		return true
	}
	// This is home dir
	if f.parent == nil {
		return false
	}
	if f.parent.Exists(name) {
		return false
	}
	f.fileMetadata.Name = name
	f.dirContent = nil
	f.DirContent()
	return true
}

func (f *FileInfo) DirContent() []byte {
	if !f.IsDir() {
		return nil
	}

	if f.dirContent == nil {
		b, err := json.Marshal(f.fileMetadata)
		if err != nil {
			log.Errorf("Marshal: %v", err)
		}
		f.dirContent = b
	}
	return f.dirContent
}

func (f *FileInfo) MIMEType() string {
	return f.fileMetadata.MIMEType
}

func (f *FileInfo) SetMIMEType(t string) {
	f.fileMetadata.MIMEType = t
}

func (f *FileInfo) Location() *types.Point {
	return f.fileMetadata.Location
}

func (f *FileInfo) SetLocation(p *types.Point) {
	f.fileMetadata.Location = p
}

func (f *FileInfo) CreatedAt() int64 {
	return f.fileMetadata.CreatedAt
}

// SetCreatedAt is for migrating use
func (f *FileInfo) SetCreatedAt(t int64) {
	f.fileMetadata.CreatedAt = t
	f.fileMetadata.ModifiedAt = t
	f.dirContent = nil
	f.DirContent()
}

func (f *FileInfo) ModifiedAt() int64 {
	return f.fileMetadata.ModifiedAt
}

func (f *FileInfo) UUID() string {
	return f.fileMetadata.UUID
}

// SetUUID is for migrating use only
func (f *FileInfo) SetUUID(id string) {
	f.fileMetadata.UUID = id
}

func (f *FileInfo) ParentUUID() string {
	if f.parent != nil {
		return f.parent.UUID()
	}
	return ""
}

func (f *FileInfo) Duration() int {
	return f.fileMetadata.Duration
}

// SetUUID is for migrating use only
func (f *FileInfo) SetDuration(seconds int) {
	f.fileMetadata.Duration = seconds
}

func (f *FileInfo) Sort(order int) {
	l := &fileInfoList{
		files: f.Files,
		order: order,
	}
	sort.Sort(l)
}

func (f *FileInfo) Path() string {
	path := "/"
	if f.parent != nil {
		path = f.parent.Path()
	}
	return filepath.Join(path, f.Name())
}

const (
	OrderByCreatedTimeAsc = iota
	OrderByCreatedTimeDesc
	OrderByModTimeAsc
	OrderByModTimeDesc
	OrderBySizeAsc
	OrderBySizeDesc
	OrderByName
	OrderByMIMEType
)

type fileInfoList struct {
	files []*FileInfo
	order int
}

func (l *fileInfoList) Len() int {
	return len(l.files)
}

func (l *fileInfoList) Get(i int) *FileInfo {
	return l.files[i]
}

func (l *fileInfoList) Swap(i, j int) {
	l.files[i], l.files[j] = l.files[j], l.files[i]
}

func (l *fileInfoList) Less(i, j int) bool {
	switch l.order {
	case OrderByCreatedTimeAsc:
		return l.files[i].CreatedAt() <= l.files[j].CreatedAt()
	case OrderByCreatedTimeDesc:
		return l.files[i].CreatedAt() >= l.files[j].CreatedAt()
	case OrderByModTimeAsc:
		return l.files[i].ModifiedAt() <= l.files[j].ModifiedAt()
	case OrderByModTimeDesc:
		return l.files[i].ModifiedAt() >= l.files[j].ModifiedAt()
	case OrderBySizeAsc:
		return l.files[i].Size() <= l.files[j].Size()
	case OrderBySizeDesc:
		return l.files[i].Size() >= l.files[j].Size()
	case OrderByName:
		return l.files[i].Name() <= l.files[j].Name()
	case OrderByMIMEType:
		return l.files[i].MIMEType() <= l.files[j].MIMEType()
	default:
		return true
	}
}
