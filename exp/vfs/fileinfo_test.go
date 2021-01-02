package vfs

import (
	"testing"

	"github.com/google/uuid"
)

func TestFileInfo_Path(t *testing.T) {
	home := newFileInfo(true, "")
	dir1 := newFileInfo(true, uuid.New().String())
	home.AddSub(dir1)
	dir2 := newFileInfo(true, uuid.New().String())
	dir1.AddSub(dir2)
	t.Log(dir2.Path())
}
