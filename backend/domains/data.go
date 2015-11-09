package domains

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"

	adErrors "airdispat.ch/errors"
)

// TODO: Refactor file so that it doesn't handle the fetching of
// AirDispatch Data messages. This should be included somewhere in the
// regular API so that clients can access it directly.

// HandleData will resolve URLs that occur on the Data domain.
//
// This involves more work than the other non-API domains since it
// must fetch the AirDispatch data message and return it to the
// client.
func HandleData(
	res http.ResponseWriter,
	req *http.Request,
	env *info.Environment,
) {
	path := req.URL.Path
	components := strings.Split(path, "/")[1:]

	// Really need at least two
	if len(components) < 2 {
		return
	}

	user := components[0]
	name := strings.Join(components[1:], "/")

	dataMessage, conn, err := env.Manager.Client.GetDataMessage(name, user)
	if aerr, ok := err.(*adErrors.Error); ok {
		if aerr.Code == 5 {
			framework.WriteView(framework.Error404, res)
			return
		}
	}

	if err != nil {
		fmt.Println("[DATA] Received error getting data message", name, user)
		fmt.Println("[DATA]", err)
		framework.WriteView(framework.Error500, res)
		return
	}

	defer conn.Close()

	if dataMessage.Filename != "" {
		cd := fmt.Sprintf(`inline; filename="%s"`, dataMessage.Filename)
		res.Header().Add("Content-Disposition", cd)
	}

	res.Header().Add("Content-Length", strconv.Itoa(int(dataMessage.TrueLength())))

	decrypted, err := dataMessage.DecryptReader(conn)
	if err != nil {
		fmt.Println("Error decrypting data", err)
		framework.WriteView(framework.Error500, res)
		return
	}

	_, err = io.Copy(res, decrypted)
	if err != nil {
		fmt.Println("Error writing data", err)
		framework.WriteView(framework.Error500, res)
		return
	}

	// Well, now we're screwed
	if !dataMessage.VerifyPayload() {
		fmt.Println("Wait! That was totally not real!")
	}
}
