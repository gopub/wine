package vfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/gopub/wine/exp/vfs"
	"github.com/stretchr/testify/require"
)

func setupFS(t *testing.T) *vfs.FileSystem {
	ms := vfs.NewMemoryStorage()
	password := uuid.NewString()
	fs, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.NotEmpty(t, fs)
	err = fs.SetPassword(password)
	require.NoError(t, err)
	return fs
}

func TestNewEncryptedFileSystem(t *testing.T) {
	ms := vfs.NewMemoryStorage()
	password := uuid.NewString()
	fs, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.NotEmpty(t, fs)
	err = fs.SetPassword(password)
	require.NoError(t, err)

	fs1, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.False(t, fs1.Auth(""))

	fs2, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.False(t, fs2.Auth("123"))

	fs3, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.NotEmpty(t, fs3)
	require.True(t, fs3.Auth(password))
}

func TestFileSystem_CreateDir(t *testing.T) {
	fs := setupFS(t)
	dirName := uuid.NewString()
	dir, err := fs.Mkdir(dirName)
	require.NoError(t, err)
	require.NotEmpty(t, dir)
	require.Equal(t, dirName, dir.Name())
	require.Equal(t, true, dir.IsDir())

	subDirName := uuid.NewString()
	subDir, err := fs.Wrapper().Create(dir.UUID(), subDirName)
	require.NoError(t, err)
	require.NotEmpty(t, subDir)
	subDir.Close()

	f, err := fs.Stat(filepath.Join(dirName, subDirName))
	require.NoError(t, err)
	require.NotEmpty(t, f)

	f, err = fs.Stat(filepath.Join(dirName, uuid.NewString()))
	require.Error(t, err)
	require.Empty(t, f)

	routes := dir.Route()
	require.Equal(t, 2, len(routes))
}

func TestFileSystem_CreateFile(t *testing.T) {
	fs := setupFS(t)

	t.Run("CreateFileInHome", func(t *testing.T) {
		fileName := uuid.NewString()
		f, err := fs.Create(fileName)
		require.NoError(t, err)
		require.NotEmpty(t, f)
		require.Equal(t, fileName, f.Info().Name())
		require.Empty(t, f.Info().Size())
		require.NotEmpty(t, f.Info().CreatedAt)
		require.NotEmpty(t, f.Info().ModifiedAt)
		f.Close()

		of, err := fs.OpenFile(fileName, vfs.ReadOnly)
		require.NoError(t, err)
		require.NotEmpty(t, of)
		require.Equal(t, f.Info(), of.Info())
	})

	t.Run("CreateFileInDir", func(t *testing.T) {
		dir, err := fs.Mkdir(uuid.NewString())
		require.NoError(t, err)
		fileName := uuid.NewString()
		f, err := fs.Wrapper().Create(dir.UUID(), fileName)
		require.NoError(t, err)
		require.NotEmpty(t, f)
		require.Equal(t, fileName, f.Info().Name())
		require.Empty(t, f.Info().Size())
		require.NotEmpty(t, f.Info().CreatedAt)
		require.NotEmpty(t, f.Info().ModifiedAt)
		f.Close()

		of, err := fs.OpenFile(fileName, vfs.ReadOnly)
		require.Error(t, err)
		require.Empty(t, of)

		of, err = fs.OpenFile(dir.Name()+"/"+fileName, vfs.ReadOnly)
		require.NoError(t, err)
		require.NotEmpty(t, of)
		require.Equal(t, f.Info(), of.Info())
	})
}

func TestFileSystem_Delete(t *testing.T) {
	fs := setupFS(t)

	t.Run("DeleteExisted", func(t *testing.T) {
		fileName := uuid.NewString()
		f, err := fs.Create(fileName)
		require.NoError(t, err)

		err = fs.Wrapper().Remove(f.Info().UUID())
		require.NoError(t, err)

		_, err = fs.OpenFile(fileName, vfs.ReadOnly)
		require.Error(t, os.ErrNotExist)
	})

	t.Run("DeleteNotExisted", func(t *testing.T) {
		fileName := uuid.NewString()
		f, err := fs.Create(fileName)
		require.NoError(t, err)

		err = fs.Wrapper().Remove(f.Info().UUID())
		require.NoError(t, err)

		err = fs.Wrapper().Remove(f.Info().UUID())
		require.NoError(t, err)
	})
}

func TestFileSystem_Move(t *testing.T) {
	fs := setupFS(t)

	fileName := uuid.NewString()
	f, err := fs.Create(fileName)
	require.NoError(t, err)
	f.Close()

	dir, err := fs.Mkdir(uuid.NewString())
	require.NoError(t, err)

	err = fs.Wrapper().Move(f.Info().UUID(), dir.UUID())
	require.NoError(t, err)

	_, err = fs.OpenFile(dir.Name()+"/"+f.Info().Name(), vfs.ReadOnly)
	require.NoError(t, err)
}

func TestFileSystem_Mount(t *testing.T) {
	ms := vfs.NewMemoryStorage()
	password := uuid.NewString()
	fs, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.NoError(t, fs.SetPassword(password))
	f, err := fs.Create(uuid.NewString())
	require.NoError(t, err)
	f.Close()

	fs2, err := vfs.NewFileSystem(ms)
	require.NoError(t, err)
	require.True(t, fs2.Auth(password))
	f2, err := fs2.OpenFile(f.Info().Name(), vfs.ReadOnly)
	require.NoError(t, err)
	f2.Close()
}
