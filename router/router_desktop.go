//+build !arm

package router

import (
	"airdispat.ch/tracker"
)

func getTrackerURL(url string) string {
	return tracker.GetTrackingServerLocationFromURL(url)
}
