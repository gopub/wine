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
	password := uuid.New().String()
	fs, err := vfs.NewFileSystem(ms, 0, password)
	require.NoError(t, err)
	require.NotEmpty(t, fs)
	return fs
}

func TestNewEncryptedFileSystem(t *testing.T) {
	ms := vfs.NewMemoryStorage()
	password := uuid.New().String()
	fs, err := vfs.NewFileSystem(ms, 0, password)
	require.NoError(t, err)
	require.NotEmpty(t, fs)

	_, err = vfs.NewFileSystem(ms, 0, "")
	require.Error(t, err)

	_, err = vfs.NewFileSystem(ms, 0, "incorrectpassword")
	require.Error(t, err)

	fs2, err := vfs.NewFileSystem(ms, 0, password)
	require.NoError(t, err)
	require.NotEmpty(t, fs2)

}

func TestFileSystem_CreateDir(t *testing.T) {
	fs := setupFS(t)
	dirName := uuid.New().String()
	dir, err := fs.Mkdir(dirName)
	require.NoError(t, err)
	require.NotEmpty(t, dir)
	require.Equal(t, dirName, dir.Name())
	require.Equal(t, true, dir.IsDir())

	subDirName := uuid.New().String()
	subDir, err := fs.Wrapper().Create(dir.UUID(), subDirName)
	require.NoError(t, err)
	require.NotEmpty(t, subDir)
	subDir.Close()

	f, err := fs.Stat(filepath.Join(dirName, subDirName))
	require.NoError(t, err)
	require.NotEmpty(t, f)

	f, err = fs.Stat(filepath.Join(dirName, uuid.New().String()))
	require.Error(t, err)
	require.Empty(t, f)
}

func TestFileSystem_CreateFile(t *testing.T) {
	fs := setupFS(t)

	t.Run("CreateFileInHome", func(t *testing.T) {
		fileName := uuid.New().String()
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
		dir, err := fs.Mkdir(uuid.New().String())
		require.NoError(t, err)
		fileName := uuid.New().String()
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
		fileName := uuid.New().String()
		f, err := fs.Create(fileName)
		require.NoError(t, err)

		err = fs.Wrapper().Remove(f.Info().UUID())
		require.NoError(t, err)

		_, err = fs.OpenFile(fileName, vfs.ReadOnly)
		require.Error(t, os.ErrNotExist)
	})

	t.Run("DeleteNotExisted", func(t *testing.T) {
		fileName := uuid.New().String()
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

	fileName := uuid.New().String()
	f, err := fs.Create(fileName)
	require.NoError(t, err)
	f.Close()

	dir, err := fs.Mkdir(uuid.New().String())
	require.NoError(t, err)

	err = fs.Wrapper().Move(f.Info().UUID(), dir.UUID())
	require.NoError(t, err)

	_, err = fs.OpenFile(dir.Name()+"/"+f.Info().Name(), vfs.ReadOnly)
	require.NoError(t, err)
}

func TestFileSystem_Mount(t *testing.T) {
	ms := vfs.NewMemoryStorage()
	password := uuid.New().String()
	fs, err := vfs.NewFileSystem(ms, 0, password)
	require.NoError(t, err)
	f, err := fs.Create(uuid.New().String())
	require.NoError(t, err)
	f.Close()

	fs2, err := vfs.NewFileSystem(ms, 0, password)
	require.NoError(t, err)
	f2, err := fs2.OpenFile(f.Info().Name(), vfs.ReadOnly)
	require.NoError(t, err)
	f2.Close()
}
