package wine

import (
	"github.com/gopub/wine/internal/path"
)

// JoinURLPath joins segments into a url path
func JoinURLPath(segment ...string) string {
	return path.Join(segment...)
}
