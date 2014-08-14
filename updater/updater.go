package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type update struct {
	Version   string
	Changelog string
	Download  *url.URL
}

// Environmental Variables Required
//
// MLGAPP  - the directory where the application is located
// MLGPORT - the port where the control server is located
func main() {
	err := getUpdate("0.0.1", "mac")
	fmt.Println(err)
}

type updateStatus struct {
	msg      string
	noUpdate bool
	updated  bool
	err      error
}

func (u updateStatus) Error() string {
	return fmt.Sprintf("%s %s", u.msg, u.err)
}

func getUpdate(currentVersion, platform string) updateStatus {
	// Check the updates site
	updateURL := fmt.Sprintf("http://getmelange.com/updates?version=%s&platform=%s", currentVersion, platform)
	resp, err := http.Get(updateURL)
	if err != nil {
		return updateStatus{
			msg: "Couldn't get updates, got error",
			err: err,
		}
	}
	defer resp.Body.Close()

	// Check for update code
	if resp.StatusCode == 422 {
		return updateStatus{
			msg:      "No update available.",
			noUpdate: true,
		}
	}
	if resp.StatusCode != 200 {
		return updateStatus{
			msg: fmt.Sprintf("Error getting update feed. %d %s", resp.StatusCode, resp.Body),
			err: nil,
		}
	}

	// Decode the response
	update := &update{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(update)
	if err != nil {
		return updateStatus{
			msg: "Couldn't decode update stream, got error",
			err: err,
		}
	}

	// Download the update
	updateData, err := http.Get(update.Download.String())
	if err != nil {
		return updateStatus{
			msg: "Couldn't download update",
			err: err,
		}
	}

	file, err := ioutil.TempFile("", "melange_update")
	if err != nil {
		return updateStatus{
			msg: "Couldn't create temporary file, got error",
			err: err,
		}
	}
	defer file.Close()

	n, err := io.Copy(file, updateData.Body)
	if err != nil {
		return updateStatus{
			msg: "Error downloading data",
			err: err,
		}
	}

	// Unzip the update
	dir, err := ioutil.TempDir("", "melange_update_extract")
	if err != nil {
		return updateStatus{
			msg: "Couldn't get extraction temp dir, got err",
			err: err,
		}
	}

	// Go back to the beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return updateStatus{
			msg: "Couldn't seek file.",
			err: err,
		}
	}

	// Unzip the update
	err = extractZip(file, n, dir)
	if err != nil {
		return updateStatus{
			msg: "Couldn't unzip file.",
			err: err,
		}
	}

	// Tell everyone to go down for update...
	updateResp, err := http.Get("http://localhost:" + os.Getenv("MLGPORT") + "/update")
	if err != nil || updateResp.StatusCode != 200 {
		return updateStatus{
			msg: "Can't update now",
			err: err,
		}
	}
	time.Sleep(1 * time.Second)

	// Remove old Melange
	err = os.RemoveAll(os.Getenv("MLGAPP"))
	if err != nil {
		return updateStatus{
			msg: "Can't remove the old melange",
			err: err,
		}
	}

	// REname new Melange
	err = os.Rename(dir, os.Getenv("MLGAPP"))
	if err != nil {
		return updateStatus{
			msg: "Can't move in new melange",
			err: err,
		}
	}

	return updateStatus{
		updated: true,
	}
}
