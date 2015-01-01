//+build android

package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func getTrackerURL(url string) string {
	requestURL := fmt.Sprintf(
		"http://www.getmelange.com/api/resolve/%s",
		url,
	)

	resp, err := http.Get(requestURL)
	if err != nil {
		fmt.Println("Got error resolving SRV record", err)
		return url
	} else if resp.StatusCode != 200 {
		fmt.Println("Got error status", resp.Status)
		return url
	}

	data, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()
	if err != nil {
		fmt.Println("Error reading body", err)
		return url
	}

	return string(data)
}
