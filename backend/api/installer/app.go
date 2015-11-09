package installer

import (
	"fmt"

	"getmelange.com/backend/api/router"
	"getmelange.com/backend/framework"
	"getmelange.com/updater"
)

// CheckUpdateController will check to see if there is a new Melange
// update available.
type CheckUpdateController struct{}

// Get will perform the check for update.
func (g *CheckUpdateController) Get(req *router.Request) framework.View {
	update, err := updater.CheckForUpdate(
		req.Environment.Version,
		req.Environment.Platform)
	if err.NoUpdate {
		return &framework.HTTPError{
			ErrorCode: 422,
			Message:   "No update",
		}
	} else if err.HasError() {
		fmt.Println("Got error checking for update", err.Error())
		return &framework.HTTPError{
			ErrorCode: 500,
			Message:   "Update error.",
		}
	}

	return &framework.JSONView{
		Content: update,
	}
}

var requestStatus chan chan float64
var errorChan chan error
var dirChan chan string

// DownloadUpdateController will actually download the melange update.
type DownloadUpdateController struct{}

// Post will initiate the download.
func (d *DownloadUpdateController) Post(req *router.Request) framework.View {
	u := &updater.Update{}
	err := req.JSON(u)
	if err != nil {
		fmt.Println("Got error decoding body", err)
		return framework.Error500
	}

	prog := make(chan float64)
	requestStatus = make(chan chan float64)

	errorChan = make(chan error)
	dirChan = make(chan string)
	quitChan := make(chan struct{})

	// Initiate the download in the background.
	go func() {
		defer func() {
			quitChan <- struct{}{}
		}()

		dir, status := updater.DownloadUpdate(u.Download, prog)
		if status.HasError() {
			errorChan <- status
			return
		}
		dirChan <- dir
	}()

	// Create a goroutine to keep the progress for the download.
	go func() {
		progress := 0.0
		for {
			select {
			case progress = <-prog:
			case r := <-requestStatus:
				r <- progress
			case <-quitChan:
				return
			}
		}
	}()

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}

// UpdateProgressController will return the progress of the download.
type UpdateProgressController struct{}

// Get will return the progress of the download.
func (d *UpdateProgressController) Get(req *router.Request) framework.View {
	select {
	case err := <-errorChan:
		dirChan = nil
		requestStatus = nil
		errorChan = nil
		fmt.Println("Got error downloading", err)
		return framework.Error500
	case dir := <-dirChan:
		dirChan = nil
		requestStatus = nil
		errorChan = nil
		return &framework.JSONView{
			Content: map[string]string{
				"dir": dir,
			},
		}
	default:
		res := make(chan float64)
		requestStatus <- res
		p := <-res
		return &framework.JSONView{
			Content: map[string]float64{
				"progress": p,
			},
		}
	}
}

// InstallUpdateController will actually install the update.
type InstallUpdateController struct{}

type installationRequest struct {
	Dir string `json:"dir"`
}

// Post will actually start installing the update.
func (i *InstallUpdateController) Post(req *router.Request) framework.View {
	fmt.Println("Going to install update")

	u := &installationRequest{}
	err := req.JSON(u)
	if err != nil {
		fmt.Println("Got error decoding body", err)
		return framework.Error500
	}

	status := updater.InstallUpdate(u.Dir, req.Environment.AppLocation)
	if status.HasError() {
		fmt.Println("Error installing update", status)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}
