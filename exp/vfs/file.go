package vfs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gopub/errors"
	"github.com/gopub/wine/httpvalue"
)

type File struct {
	buf    *bytes.Buffer
	offset int64

	info *FileInfo
	vo   *FileSystem
	flag Flag
}

var _ http.File = (*File)(nil)

func newFile(vo *FileSystem, info *FileInfo, flag Flag) *File {
	if (flag&ReadOnly) != 0 && (flag&WriteOnly) != 0 {
		panic("invalid flag")
	}
	f := &File{
		vo:   vo,
		info: info,
		flag: flag,
		buf:  bytes.NewBuffer(nil),
	}
	if flag&WriteOnly != 0 {
		info.busy = true
	}
	return f
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.info.IsDir() {
		return 0, errors.New("cannot seek dir")
	}
	if f.flag&WriteOnly != 0 {
		return 0, errors.New("cannot seek in write mode")
	}
	if offset >= f.info.Size() {
		return f.offset, io.EOF
	}
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.offset + offset
	case io.SeekEnd:
		// abs = f.info.Size() - 1 + offset ?
		abs = f.info.Size() + offset
	default:
		return f.offset, fmt.Errorf("invalid whence: %d", whence)
	}
	if abs < 0 {
		return f.offset, fmt.Errorf("negative position: %d", abs)
	}

	// abs >= f.info.Size() ?
	if abs > f.info.Size() {
		return f.offset, fmt.Errorf("overflow: %d", abs)
	}
	f.buf.Reset()
	f.offset = abs
	return f.offset, nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	if !f.info.IsDir() {
		return nil, errors.New("not dir")
	}
	l := make([]os.FileInfo, len(f.info.Files))
	for i, fi := range f.info.Files {
		l[i] = fi
	}
	return l, nil
}

func (f *File) Stat() (os.FileInfo, error) {
	return f.info, nil
}

func (f *File) Read(p []byte) (int, error) {
	if f.flag&WriteOnly != 0 {
		return 0, os.ErrPermission
	}

	if f.info.IsDir() {
		var nr int
		if f.offset < f.info.Size() {
			nr = copy(p, f.info.DirContent()[f.offset:])
		}
		f.offset += int64(nr)
		if f.offset >= f.info.Size() {
			return nr, io.EOF
		}
		return nr, nil
	}

	nExpected := len(p)
	nRead := 0
	for nRead < nExpected {
		n, err := f.read(p[nRead:])
		if err != nil {
			return nRead, err
		}
		nRead += n
	}
	return nRead, nil
}

func (f *File) read(p []byte) (int, error) {
	if f.offset >= f.info.Size() {
		return 0, io.EOF
	}

	if f.buf.Len() == 0 {
		// load one page to buf
		pageIndex := f.offset / f.vo.pageSize
		page := f.info.Pages[pageIndex]
		data, err := f.vo.storage.Get(page)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return 0, io.EOF
			}
			return 0, fmt.Errorf("load page %s: %w", page, err)
		}

		if err := f.vo.DecryptPage(data); err != nil {
			return 0, fmt.Errorf("decrypt: %w", err)
		}
		nw, err := f.buf.Write(data)
		if err != nil {
			return 0, fmt.Errorf("write to buf: %w", err)
		}

		if nw != len(data) {
			return 0, errors.New("cannot write to buf")
		}

		f.buf.Grow(int(f.offset % f.vo.pageSize))
	}

	nr, err := f.buf.Read(p)
	f.offset += int64(nr)
	// EOF of reading buf is not an error
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return nr, err
}

func (f *File) Write(p []byte) (int, error) {
	if f.flag&WriteOnly == 0 {
		return 0, os.ErrPermission
	}
	_, err := f.buf.Write(p)
	if err != nil {
		return 0, err
	}
	err = f.flush(false)
	if err != nil {
		n := len(p) - f.buf.Len()
		if n < 0 {
			n = 0
		}
		return n, err
	}
	return len(p), nil
}

func (f *File) WriteThumbnail(b []byte) error {
	if f.flag&WriteOnly == 0 {
		return os.ErrPermission
	}
	if f.info.Thumbnail == "" {
		f.info.Thumbnail = uuid.New().String()
	}
	return f.vo.storage.Put(f.info.Thumbnail, b)
}

func (f *File) ReadThumbnail() ([]byte, error) {
	if f.flag&WriteOnly != 0 {
		return nil, os.ErrPermission
	}
	if f.info.Thumbnail == "" {
		return nil, os.ErrNotExist
	}
	return f.vo.storage.Get(f.info.Thumbnail)
}

func (f *File) Close() error {
	f.info.busy = false
	if f.flag&WriteOnly != 0 {
		return f.flush(true)
	}
	f.offset = 0
	return nil
}

func (f *File) flush(all bool) error {
	if f.buf.Len() > 0 && (f.offset == 0 || f.info.MIMEType() == "") {
		// detect at the beginning (offset==0)
		// or if prior detection failed (f.info.MIMEType()=="")
		f.info.SetMIMEType(httpvalue.DetectContentType(f.buf.Bytes()))
	}
	for all || int64(f.buf.Len()) >= f.vo.pageSize {
		b := make([]byte, f.vo.pageSize)
		n, err := f.buf.Read(b)
		// even err is io.EOF, n may be > 0
		if n > 0 {
			if f.offset == 0 {
				f.info.truncate()
			}
			f.offset += int64(n)
			page := uuid.New().String()
			data := b[:n]
			if er := f.vo.EncryptPage(data); er != nil {
				return fmt.Errorf("encrypt: %w", er)
			}

			if er := f.vo.storage.Put(page, data); er != nil {
				return fmt.Errorf("put: %w", er)
			}
			f.info.addPage(page)
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}

	if all {
		f.info.setSize(f.offset)
		if err := f.vo.SaveFileTree(); err != nil {
			return fmt.Errorf("save file info list: %w", err)
		}
	}
	return nil
}

func (f *File) Info() *FileInfo {
	return f.info
}
