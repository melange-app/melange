//+build android

package packaging

import (
	"os"
	"path/filepath"

	"code.google.com/p/go-uuid/uuid"
)

func (p *Packager) getTempFile() (*os.File, error) {
	return os.OpenFile(
		filepath.Join(p.Plugin, uuid.New()),
		os.O_RDWR|os.O_CREATE,
		os.ModeTemporary,
	)
}
