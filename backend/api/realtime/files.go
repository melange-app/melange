package realtime

import (
	"fmt"
	"os"
	"path/filepath"

	"airdispat.ch/identity"

	mIdentity "getmelange.com/backend/models/identity"
)

const (
	fileUploadType   = "uploadFile"
	fileProgressType = "uploadProgress"
	fileDoneType     = "uploadedFile"
	fileErrorType    = "uploadError"
)

// fileUploadRequest represents an object that contains all the necessary
// information to upload a file.
type fileUploadRequest struct {
	ID       string   `json:"id"`
	Filename string   `json:"filename"`
	Type     string   `json:"type"`
	To       []string `json:"to"`
	Name     string   `json:"name"`

	progress  chan float64
	finalName chan string
}

type fileUploadResponse struct {
	ID string `json:"id"`

	// Objects associated with a successful upload.
	Name string `json:"name,omitempty"`
	User string `json:"user,omitempty"`
	URL  string `json:"url,omitempty"`

	// Objects associated with upload progress.
	Progress float64 `json:"progress,omitempty"`
}

func getDataURL(alias, name string) string {
	return fmt.Sprintf("http://data.local.getmelange.com:7776/%s/%s", alias, name)
}

// FileResponder is an object that handles uploading of objects into
// Melange.
type FileResponder struct{}

func (m *FileResponder) checkStatusLoop(req *fileUploadRequest, response chan *Message, alias *mIdentity.Alias) {
	for {
		data, ok := <-req.progress

		// Exit on Close
		if !ok {
			name := <-req.finalName

			response <- mustCreateMessage(fileDoneType, fileUploadResponse{
				ID:   req.ID,
				Name: name,
				User: alias.String(),
				URL:  getDataURL(alias.String(), name),
			})

			return
		}

		response <- mustCreateMessage(fileProgressType, fileUploadResponse{
			ID:       req.ID,
			Progress: data,
		})
	}
}

// Handle will perform the file upload.
func (m *FileResponder) Handle(req *Request) bool {
	// Ignore messages that aren't meant for us.
	if req.Message.Type != fileUploadType {
		return false
	}

	go m.asyncHandle(req)
	return true
}

func (m *FileResponder) asyncHandle(req *Request) {
	msg := &fileUploadRequest{}
	err := req.Message.Unmarshal(&msg)
	if err != nil {
		logError("[RLT-FILE]", "Received error unmarshalling file.", err)
		req.Response <- createError(fileErrorType, "invalid request")
		return
	}

	// Open the file that we are uploading.
	file, err := os.Open(msg.Filename)
	if err != nil {
		logError("[RLT-FILE]", "Received error opening file.", err)
		req.Response <- createError(fileErrorType, "file does not exist")
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		logError("[RLT-FILE]", "Received error getting file info.", err)
		req.Response <- createError(fileErrorType, "file does not exist")
		return
	}

	alias := &mIdentity.Alias{}
	err = req.Environment.Manager.Identity.Aliases.Limit(1).One(req.Environment.Store, alias)
	if err != nil {
		logError("[RLT-FILE]", "Received error loading user alias.", err)
		req.Response <- createError(fileErrorType, "cannot get alias")
		return
	}

	// Create channels to mark progress
	msg.finalName = make(chan string, 1)
	msg.progress = make(chan float64)

	// Get upload status and send it to web socket
	go m.checkStatusLoop(msg, req.Response, alias)

	// Construct the uploader
	uploader := &fileProgressReader{
		total:    info.Size(),
		notifier: msg.progress,
		File:     file,
	}

	// Convert the addresses into AirDispatch addresses. Since
	// Melange doesn't support private file upload yet, this may
	// not work.
	var toAddresses []*identity.Address
	for _, v := range msg.To {
		toAddresses = append(toAddresses, identity.CreateAddressFromString(v))
	}

	// Upload the message
	name, err := req.Environment.Manager.Client.PublishDataMessage(
		uploader,
		toAddresses,
		msg.Type, msg.Name,
		filepath.Base(msg.Filename))

	if err == nil {
		msg.finalName <- name
	} else {
		close(msg.finalName)
		close(msg.progress)

		logError("[RLT-FILE]", "Received error uploading file.", err)
		req.Response <- createError(fileErrorType, "could not upload file")
	}
}

// fileProgressReader controls how we are reading the file data. It keeps
// track of the progress that we are making reading the file.
type fileProgressReader struct {
	total     int64
	read      int64
	timesRead int

	notifier chan<- float64

	*os.File
}

func (r *fileProgressReader) Read(p []byte) (int, error) {
	n, err := r.File.Read(p)
	r.read += int64(n)

	if r.timesRead > 1 && r.notifier != nil {
		select {
		case r.notifier <- (float64(r.read) / float64(r.total)):
		default:
			// If we are unable to push the notification, that is okay.
		}
	}

	if r.read >= r.total {
		r.read = 0
		r.timesRead++

		if r.timesRead == 3 {
			close(r.notifier)
			r.notifier = nil
		}
	}

	return n, err
}
