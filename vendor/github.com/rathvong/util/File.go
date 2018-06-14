package util

import (

	"os"

)

// Mark -> Check file directory to see if the directory exists.
func  DirectoryExists(filePath string) (exists bool) {
	exists = true

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	return
}


