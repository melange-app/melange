package controllers

import (
	"fmt"
	"net/http"

	"getmelange.com/app/framework"
	"getmelange.com/updater"
)

type CheckUpdateController struct {
	Version  string
	Platform string
}

func (g *CheckUpdateController) Handle(req *http.Request) framework.View {
	update, err := updater.CheckForUpdate(g.Version, g.Platform)
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

type DownloadUpdateController struct{}

func (d *DownloadUpdateController) Handle(req *http.Request) framework.View {
	u := &updater.Update{}
	err := DecodeJSONBody(req, u)
	if err != nil {
		fmt.Println("Got error decoding body", err)
		return framework.Error500
	}

	prog := make(chan float64)
	requestStatus = make(chan chan float64)

	errorChan = make(chan error)
	dirChan = make(chan string)
	quitChan := make(chan struct{})

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

type UpdateProgressController struct{}

func (d *UpdateProgressController) Handle(req *http.Request) framework.View {
	select {
	case err := <-errorChan:
		// Cleaup
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

type InstallUpdateController struct {
	AppLocation string
}

func (i *InstallUpdateController) Handle(req *http.Request) framework.View {
	fmt.Println("Going to install update")

	u := make(map[string]string)
	err := DecodeJSONBody(req, &u)
	if err != nil {
		fmt.Println("Got error decoding body", err)
		return framework.Error500
	}

	dir, ok := u["dir"]
	if !ok {
		return &framework.HTTPError{
			ErrorCode: 400,
			Message:   "Need dir.",
		}
	}

	status := updater.InstallUpdate(dir, i.AppLocation)
	if status.HasError() {
		fmt.Println("Error installing update", status)
		return framework.Error500
	}

	return &framework.HTTPError{
		ErrorCode: 200,
		Message:   "OK",
	}
}
