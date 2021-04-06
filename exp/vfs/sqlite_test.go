package vfs_test

import (
	"github.com/google/uuid"
	"github.com/gopub/wine/exp/vfs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"sort"
	"testing"
)

func TestSQLiteStorage_ListKeysByLength(t *testing.T) {
	s, err := vfs.NewSQLiteStorage(filepath.Join(t.TempDir(), uuid.NewString()))
	require.NoError(t, err)
	var keys1 []string
	var keys2 []string
	for i := 0; i < 10; i++ {
		keys1 = append(keys1, uuid.NewString())
		err = s.Put(keys1[i], []byte(keys1[i]))
		require.NoError(t, err)
	}
	for i := 0; i < 15; i++ {
		keys2 = append(keys2, uuid.NewString()+uuid.NewString())
		err = s.Put(keys2[i], []byte(keys2[i]))
		require.NoError(t, err)
	}

	l, err := s.ListKeysByLength(0)
	require.NoError(t, err)
	require.Empty(t, l)
	l, err = s.ListKeysByLength(1)
	require.NoError(t, err)
	require.Empty(t, l)
	l, err = s.ListKeysByLength(len(keys1[0]))
	require.NoError(t, err)
	sort.Strings(keys1)
	sort.Strings(l)
	require.Equal(t, keys1, l)
	l, err = s.ListKeysByLength(len(keys2[0]))
	require.NoError(t, err)
	sort.Strings(keys2)
	sort.Strings(l)
	require.Equal(t, keys2, l)
}
