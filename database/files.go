package database

import (
	"io"
	"log"
	"os"
)

// -rw-------
const filePerm = 0600

func createFile(filename string, src io.Reader) (int64, error) {
	// Create a new non-existent file as write-only.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, filePerm)
	if err != nil {
		return 0, err
	}
	size, err := io.Copy(file, src)
	// Always close the file. If there was an error during `Copy()`, still
	// return that, otherwise return the `Close()` error.
	if closeErr := file.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	// If there was an error anywhere, delete the file. Always return the other
	// error; we'll just print out the `deleteFile()` error for information.
	if err != nil {
		if deleteErr := deleteFile(filename); deleteErr != nil {
			log.Printf("%v\n", deleteErr)
		}
	}
	return size, err
}

func openFile(filename string) (io.ReadCloser, error) {
	// Open the file as read-only.
	return os.OpenFile(filename, os.O_RDONLY, filePerm)
}

func deleteFile(filename string) error {
	return os.Remove(filename)
}

func deleteAllFiles(filenames []string) error {
	for _, filename := range filenames {
		if err := deleteFile(filename); err != nil {
			return err
		}
	}
	return nil
}
