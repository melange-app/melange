package realtime

import "fmt"

type linkError struct {
	Error string `json:"error"`
}

func createError(typ, err string) *Message {
	return mustCreateMessage(typ, &linkError{
		Error: err,
	})
}

func logError(place interface{}, msgs ...interface{}) {
	for _, msg := range msgs {
		fmt.Println(place, msg)
	}
}
