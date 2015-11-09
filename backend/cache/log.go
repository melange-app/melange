package cache

import "fmt"

func logError(prefix, msg string, err error) {
	fmt.Println(prefix, msg)
	fmt.Println(prefix, err)
}
