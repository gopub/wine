package vfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/gopub/types"
	"github.com/gopub/wine/exp/vfs"
	"github.com/gopub/wine/httpvalue"
	"github.com/stretchr/testify/require"
)

func setupFS(t *testing.T) *vfs.FileSystem {
	ms := vfs.NewMemoryStorage()
	password := uuid.New().String()
	fs, err := vfs.NewEncryptedFileSystem(ms, password)
	require.NoError(t, err)
	require.NotEmpty(t, fs)
	return fs
}

func TestNewEncryptedFileSystem(t *testing.T) {
	ms := vfs.NewMemoryStorage()
	password := uuid.New().String()
	fs, err := vfs.NewEncryptedFileSystem(ms, password)
	require.NoError(t, err)
	require.NotEmpty(t, fs)

	_, err = vfs.NewFileSystem(ms)
	require.Error(t, err)

	_, err = vfs.NewEncryptedFileSystem(ms, "incorrectpassword")
	require.Error(t, err)

	fs2, err := vfs.NewEncryptedFileSystem(ms, password)
	require.NoError(t, err)
	require.NotEmpty(t, fs2)

}

func TestFileSystem_CreateDir(t *testing.T) {
	fs := setupFS(t)
	dirName := uuid.New().String()
	dir, err := fs.CreateDir(nil, dirName)
	require.NoError(t, err)
	require.NotEmpty(t, dir)
	require.Equal(t, dirName, dir.Info().Name())
	require.Equal(t, true, dir.Info().IsDir())

	subDirName := uuid.New().String()
	subDir, err := fs.CreateDir(dir.Info(), subDirName)
	require.NoError(t, err)
	require.NotEmpty(t, subDir)

	f, err := fs.OpenFile(filepath.Join(dirName, subDirName), true)
	require.NoError(t, err)
	require.NotEmpty(t, f)

	f, err = fs.OpenFile(filepath.Join(dirName, uuid.New().String()), true)
	require.Error(t, err)
	require.Empty(t, f)
}

func TestFileSystem_CreateFile(t *testing.T) {
	fs := setupFS(t)

	t.Run("CreateFileInHome", func(t *testing.T) {
		fileName := uuid.New().String()
		f, err := fs.CreateFile(nil, fileName)
		require.NoError(t, err)
		require.NotEmpty(t, f)
		require.Equal(t, fileName, f.Info().Name())
		require.Empty(t, f.Info().Size())
		require.NotEmpty(t, f.Info().CreatedAt)
		require.NotEmpty(t, f.Info().ModifiedAt)

		of, err := fs.OpenFile(fileName, false)
		require.NoError(t, err)
		require.NotEmpty(t, of)
		require.Equal(t, f.Info(), of.Info())
	})

	t.Run("CreateMIMEFile", func(t *testing.T) {
		fileName := uuid.New().String()
		f, err := fs.CreateMIMEFile(nil, fileName, httpvalue.MPEG, types.NewPoint(23.1, 90.2))
		require.NoError(t, err)
		require.NotEmpty(t, f)
		require.Equal(t, fileName, f.Info().Name())
		require.Equal(t, httpvalue.MPEG, f.Info().MIMEType)
		require.NotEmpty(t, f.Info().Location)
		require.Empty(t, f.Info().Size())
		require.NotEmpty(t, f.Info().CreatedAt)
		require.NotEmpty(t, f.Info().ModifiedAt)

		of, err := fs.OpenFile(fileName, false)
		require.NoError(t, err)
		require.NotEmpty(t, of)
		require.Equal(t, f.Info(), of.Info())
	})

	t.Run("CreateFileInDir", func(t *testing.T) {
		dir, err := fs.CreateDir(nil, uuid.New().String())
		require.NoError(t, err)
		fileName := uuid.New().String()
		f, err := fs.CreateFile(dir.Info(), fileName)
		require.NoError(t, err)
		require.NotEmpty(t, f)
		require.Equal(t, fileName, f.Info().Name())
		require.Empty(t, f.Info().Size())
		require.NotEmpty(t, f.Info().CreatedAt)
		require.NotEmpty(t, f.Info().ModifiedAt)

		of, err := fs.OpenFile(fileName, false)
		require.Error(t, err)
		require.Empty(t, of)

		of, err = fs.OpenFile(dir.Info().Name()+"/"+fileName, false)
		require.NoError(t, err)
		require.NotEmpty(t, of)
		require.Equal(t, f.Info(), of.Info())
	})
}

func TestFileSystem_Delete(t *testing.T) {
	fs := setupFS(t)

	t.Run("DeleteExisted", func(t *testing.T) {
		fileName := uuid.New().String()
		f, err := fs.CreateFile(nil, fileName)
		require.NoError(t, err)

		err = fs.Delete(f.Info())
		require.NoError(t, err)

		_, err = fs.OpenFile(fileName, false)
		require.Error(t, os.ErrNotExist)
	})

	t.Run("DeleteNotExisted", func(t *testing.T) {
		fileName := uuid.New().String()
		f, err := fs.CreateFile(nil, fileName)
		require.NoError(t, err)

		err = fs.Delete(f.Info())
		require.NoError(t, err)

		err = fs.Delete(f.Info())
		require.NoError(t, err)
	})
}

func TestFileSystem_Move(t *testing.T) {
	fs := setupFS(t)

	fileName := uuid.New().String()
	f, err := fs.CreateFile(nil, fileName)
	require.NoError(t, err)

	dir, err := fs.CreateDir(nil, uuid.New().String())
	require.NoError(t, err)

	err = fs.Move(f.Info(), dir.Info())
	require.NoError(t, err)

	_, err = fs.OpenFile(dir.Info().Name()+"/"+f.Info().Name(), false)
	require.NoError(t, err)
}
