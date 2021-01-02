package vfs_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/types"
	"github.com/gopub/wine/exp/vfs"
	"github.com/gopub/wine/httpvalue"
	"github.com/stretchr/testify/require"
)

func TestFileSystemSync_Sync(t *testing.T) {
	fs1, err := vfs.NewFileSystem(vfs.NewMemoryStorage())
	require.NoError(t, err)
	fs2, err := vfs.NewFileSystem(vfs.NewMemoryStorage())
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err = fs1.KeyChain().Save(&vfs.SecKeyItem{Name: uuid.New().String(), Account: uuid.New().String(), Password: uuid.New().String()})
		require.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		var path = uuid.New().String()
		for j := 0; j < i; j++ {
			path = filepath.Join(path, uuid.New().String())
		}

		dir, err := fs1.MkdirAll(path)
		require.NoError(t, err)
		require.Equal(t, "/"+path, dir.Path(), i)
		for k := 0; k < i; k++ {
			name := uuid.New().String()
			f, err := fs1.Wrapper().Create(dir.UUID(), name)
			require.NoError(t, err)
			switch k % 3 {
			case 0:
				f.Info().SetMIMEType(httpvalue.Plain)
				_, err = f.Write([]byte(strings.Repeat(uuid.New().String(), 1024*(k+1))))
				require.NoError(t, err)
			case 1:
				f.Info().SetMIMEType(httpvalue.JSON)
				_, err = f.Write(conv.MustJSONBytes(types.M{uuid.New().String(): uuid.New().String()}))
				require.NoError(t, err)
			default:
				f.Info().SetMIMEType(httpvalue.OctetStream)
				_, err = f.Write([]byte(strings.Repeat(uuid.New().String(), 10240*(k+1))))
				require.NoError(t, err)
			}
		}
	}

	logC, errC := fs2.Sync(fs1)
LOOP:
	for {
		select {
		case log := <-logC:
			t.Log(log.Action, log.Source.Path())
			if log.Destination != nil {
				t.Log(log.Destination.Path())
			}
			break
		case err := <-errC:
			require.NoError(t, err)
			break LOOP
		}
	}
}
