//+build !android

package packaging

import (
	"io/ioutil"
	"os"
)

func (p *Packager) getTempFile() (*os.File, error) {
	return ioutil.TempFile("", "melange_plugin_download_")
}
