package wine

import (
	"github.com/gopub/wine/internal/path"
)

func JoinURLPath(segment ...string) string {
	return path.Join(segment...)
}
