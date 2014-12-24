// +build darwin

package updater

var updateFile = "Melange.app"

var executables = []string{}

var unzipper = []string{
	"ditto",
	"-xk",
}

var postUpdate = "updater"
