package vfs_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/gopub/wine/exp/vfs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFile_Write(t *testing.T) {
	fs := setupFS(t)
	t.Run("WriteLargeAmount", func(t *testing.T) {
		f, err := fs.Create(uuid.NewString())
		require.NoError(t, err)
		data := []byte(strings.Repeat(uuid.NewString(), 12345))
		n, err := f.Write(data)
		require.Equal(t, len(data), n)
		require.NoError(t, err)
		err = f.Close()
		require.NoError(t, err)

		rf, err := fs.OpenFile(f.Info().Name(), vfs.ReadOnly)
		require.NoError(t, err)
		require.NotEmpty(t, rf)
		buf := bytes.NewBuffer(nil)
		var b [1000]byte
		nr, err := rf.Read(b[:])
		require.Equal(t, true, err == nil || err == io.EOF, err)
		buf.Write(b[:nr])
		for err == nil {
			nr, err = rf.Read(b[:])
			require.Equal(t, true, err == nil || err == io.EOF, err)
			buf.Write(b[:nr])
		}
		require.Equal(t, len(data), buf.Len())
		require.Equal(t, data, buf.Bytes())
	})

	t.Run("WriteSmallAmount", func(t *testing.T) {
		f, err := fs.Create(uuid.NewString())
		require.NoError(t, err)
		data := []byte(strings.Repeat(uuid.NewString(), 2))
		n, err := f.Write(data)
		require.Equal(t, len(data), n)
		require.NoError(t, err)
		err = f.Close()
		require.NoError(t, err)

		rf, err := fs.OpenFile(f.Info().Name(), vfs.ReadOnly)
		require.NoError(t, err)
		require.NotEmpty(t, rf)
		buf := bytes.NewBuffer(nil)
		var b [1000]byte
		nr, err := rf.Read(b[:])
		require.Equal(t, true, err == nil || err == io.EOF)
		buf.Write(b[:nr])
		for err == nil {
			nr, err = rf.Read(b[:])
			require.Equal(t, true, err == nil || err == io.EOF)
			buf.Write(b[:nr])
		}
		require.Equal(t, len(data), buf.Len())
		require.Equal(t, data, buf.Bytes())
	})

	t.Run("WriteTwice", func(t *testing.T) {
		f, err := fs.Create(uuid.NewString())
		require.NoError(t, err)
		data := []byte(strings.Repeat(uuid.NewString(), 2))
		n, err := f.Write(data)
		require.Equal(t, len(data), n)
		require.NoError(t, err)
		err = f.Close()
		require.NoError(t, err)

		f, err = fs.Create(uuid.NewString())
		data = []byte(strings.Repeat(uuid.NewString(), 10))
		n, err = f.Write(data)
		require.Equal(t, len(data), n)
		f.Close()
		require.Equal(t, (int64(len(data))+fs.PageSize()-1)/fs.PageSize(), int64(len(f.Info().Pages)))

		rf, err := fs.OpenFile(f.Info().Name(), vfs.ReadOnly)
		require.NoError(t, err)
		require.NotEmpty(t, rf)
		buf := bytes.NewBuffer(nil)
		var b [1000]byte
		nr, err := rf.Read(b[:])
		require.Equal(t, true, err == nil || err == io.EOF)
		buf.Write(b[:nr])
		for err == nil {
			nr, err = rf.Read(b[:])
			require.Equal(t, true, err == nil || err == io.EOF)
			buf.Write(b[:nr])
		}
		require.Equal(t, len(data), buf.Len())
		require.Equal(t, data, buf.Bytes())
	})
}
