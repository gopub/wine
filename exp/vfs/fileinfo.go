package vfs

import (
	"encoding"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gopub/wine/httpvalue"

	"github.com/google/uuid"
	"github.com/gopub/log"
	"github.com/gopub/types"
)

type fileMetadata struct {
	UUID       string       `json:"uuid"`
	Name       string       `json:"name"`
	IsDir      bool         `json:"is_dir"`
	MIMEType   string       `json:"mime_type,omitempty"`
	Pages      []string     `json:"pages,omitempty"`
	Size       int64        `json:"size"`
	Duration   int          `json:"duration"`
	Files      []*FileInfo  `json:"files,omitempty"`
	CreatedAt  int64        `json:"created_at"`
	ModifiedAt int64        `json:"modified_at"`
	Location   *types.Point `json:"location,omitempty"`
	Permission int          `json:"permission,omitempty"`
	Extra      types.M      `json:"extra,omitempty"`
	Thumbnail  string       `json:"thumbnail,omitempty"`
	Version    int          `json:"version,omitempty"`
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

type rootFileInfo struct {
	*FileInfo
}

func newRootFile() *rootFileInfo {
	r := new(rootFileInfo)
	r.FileInfo = &FileInfo{
		fileMetadata: fileMetadata{
			UUID:       "",
			Name:       "",
			IsDir:      true,
			CreatedAt:  time.Now().Unix(),
			ModifiedAt: time.Now().Unix(),
		},
	}
	return r
}

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
	return json.Unmarshal(data, &f.fileMetadata)
}

func (f *FileInfo) MarshalBinary() (data []byte, err error) {
	return json.Marshal(f.fileMetadata)
}

func (f *FileInfo) addPage(p string) {
	f.Pages = append(f.Pages, p)
	f.fileMetadata.ModifiedAt = time.Now().Unix()
	f.invalidateDirContent()
}

func (f *FileInfo) truncate() {
	f.fileMetadata.Pages = f.fileMetadata.Pages[:0]
	f.setSize(0)
	f.invalidateDirContent()
}

func (f *FileInfo) setSize(size int64) {
	f.fileMetadata.Size = size
	f.fileMetadata.ModifiedAt = time.Now().Unix()
	f.invalidateDirContent()
}

func (f *FileInfo) SetPermission(p int) {
	f.Permission = p
	for _, sub := range f.Files {
		sub.SetPermission(p)
	}
	f.invalidateDirContent()
}

func (f *FileInfo) FindByPath(path string) *FileInfo {
	return f.find(splitPath(path)...)
}

func (f *FileInfo) find(segments ...string) *FileInfo {
	if len(segments) == 0 {
		return f
	}
	for _, fi := range f.Files {
		if fi.Name() == segments[0] {
			return fi.find(segments[1:]...)
		}
	}
	return nil
}

func (f *FileInfo) Find(baseName string) *FileInfo {
	return f.find(baseName)
}

func (f *FileInfo) FindByUUID(id string) *FileInfo {
	for _, fi := range f.Files {
		if fi.fileMetadata.UUID == id {
			return fi
		}
		if found := fi.FindByUUID(id); found != nil {
			return found
		}
	}
	return nil
}

func (f *FileInfo) Exists(baseName string) bool {
	for _, fi := range f.Files {
		if fi.Name() == baseName {
			return true
		}
	}
	return false
}

func (f *FileInfo) DistinctName(baseName string) string {
	i := 0
	ext := filepath.Ext(baseName)
	base := baseName[:len(baseName)-len(ext)]
	distinct := baseName
	for f.Exists(distinct) {
		i++
		distinct = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
	return distinct
}

func (f *FileInfo) AddSub(sub *FileInfo) {
	if sub.parent == f {
		return
	}
	if sub.parent != nil {
		sub.parent.RemoveSub(sub)
	}
	sub.parent = f
	sub.SetPermission(f.Permission)
	f.fileMetadata.Files = append(f.fileMetadata.Files, sub)
	f.fileMetadata.ModifiedAt = time.Now().Unix()
	f.invalidateDirContent()
}

func (f *FileInfo) RemoveSub(sub *FileInfo) {
	for i, fi := range f.Files {
		if fi == sub {
			f.Files = append(f.Files[:i], f.Files[i+1:]...)
			break
		}
	}
	sub.Permission = 0
	f.fileMetadata.ModifiedAt = time.Now().Unix()
	f.invalidateDirContent()
}

func (f *FileInfo) Rename(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	if strings.Contains(name, "/") {
		return false
	}

	if f.Name() == name {
		return true
	}

	// This is root dir
	if f.parent == nil {
		return false
	}

	if f.parent.Exists(name) {
		return false
	}

	f.fileMetadata.Name = name
	f.invalidateDirContent()
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

func (f *FileInfo) invalidateDirContent() {
	f.dirContent = nil
	if f.parent != nil {
		f.parent.invalidateDirContent()
	}
}

func (f *FileInfo) MIMEType() string {
	return f.fileMetadata.MIMEType
}

func (f *FileInfo) SetMIMEType(t string) {
	f.fileMetadata.MIMEType = t
	f.invalidateDirContent()
}

func (f *FileInfo) Location() *types.Point {
	return f.fileMetadata.Location
}

func (f *FileInfo) SetLocation(p *types.Point) {
	f.fileMetadata.Location = p
	f.invalidateDirContent()
}

func (f *FileInfo) CreatedAt() int64 {
	return f.fileMetadata.CreatedAt
}

// SetCreatedAt is for migrating use
func (f *FileInfo) SetCreatedAt(t int64) {
	f.fileMetadata.CreatedAt = t
	f.fileMetadata.ModifiedAt = t
	f.invalidateDirContent()
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
	f.invalidateDirContent()
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
	f.invalidateDirContent()
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

func (f *FileInfo) Route() []*FileInfo {
	if f.parent == nil {
		return []*FileInfo{f}
	}
	return append(f.parent.Route(), f)
}

func (f *FileInfo) ListByPermission(p int) []*FileInfo {
	var l []*FileInfo
	for _, fi := range f.Files {
		if fi.Permission&p != 0 {
			l = append(l, fi)
		} else {
			l = append(l, fi.ListByPermission(p)...)
		}
	}
	return l
}

func (f *FileInfo) SubFilter(fn func(info *FileInfo) bool) []*FileInfo {
	var l []*FileInfo
	for _, fi := range f.Files {
		if fn(fi) {
			l = append(l, fi)
		}
	}
	return l
}

func (f *FileInfo) GetSubDirs() []*FileInfo {
	return f.SubFilter(func(info *FileInfo) bool {
		return info.IsDir()
	})
}

func (f *FileInfo) GetSubFiles() []*FileInfo {
	return f.SubFilter(func(info *FileInfo) bool {
		return !info.IsDir()
	})
}

func (f *FileInfo) FindByName(name string) []*FileInfo {
	var res []*FileInfo
	name = strings.ToLower(name)
	for _, file := range f.Files {
		if strings.Index(strings.ToLower(file.Name()), name) >= 0 {
			res = append(res, file)
		}
		res = append(res, file.FindByName(name)...)
	}
	return res
}

func (f *FileInfo) IsText() bool {
	return httpvalue.IsMIMETextType(f.MIMEType())
}

func (f *FileInfo) IsImage() bool {
	return strings.HasPrefix(f.MIMEType(), "image")
}

func (f *FileInfo) IsVideo() bool {
	return strings.HasPrefix(f.MIMEType(), "video")
}

func (f *FileInfo) IsAudio() bool {
	return strings.HasPrefix(f.MIMEType(), "audio")
}

func (f *FileInfo) totalSize() int64 {
	n := f.Size()
	if f.IsDir() {
		for _, sub := range f.Files {
			n += sub.totalSize()
		}
	}
	return n
}

func (f *FileInfo) makeDoubleLinked() {
	for _, sub := range f.Files {
		sub.parent = f
		sub.makeDoubleLinked()
	}
}

func (f *FileInfo) HasInheritedPermission(p int) bool {
	if f.Permission == p {
		return true
	}

	if f.parent == nil {
		return false
	}

	return f.parent.HasInheritedPermission(p)
}

func (f *FileInfo) ListAllPages() []string {
	pages := make([]string, len(f.Pages))
	copy(pages, f.Pages)
	if f.Thumbnail != "" {
		pages = append(pages, f.Thumbnail)
	}
	if f.IsDir() {
		for _, sub := range f.Files {
			pages = append(pages, sub.ListAllPages()...)
		}
	}
	return pages
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
