package vfs_test

import (
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
	"github.com/gopub/wine/exp/vfs"
	"github.com/stretchr/testify/require"
)

func TestFileSystemWrapper_ImportDiskFile(t *testing.T) {
	fs, err := vfs.NewFileSystem(vfs.NewMemoryStorage())
	require.NoError(t, err)
	fs.SetPassword(uuid.New().String())
	fi, err := fs.Wrapper().ImportDiskFile("", "vfs.go")
	require.NoError(t, err)
	require.Equal(t, "vfs.go", fi.Name())

	data, err := fs.Wrapper().Read(fi.UUID())
	require.NoError(t, err)

	origin, err := ioutil.ReadFile("vfs.go")
	require.NoError(t, err)

	require.Equal(t, origin, data)
}
