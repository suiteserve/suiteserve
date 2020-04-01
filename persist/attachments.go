package persist

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"os"
	"time"
)

type Attachment struct {
	Name string
	File *os.File
}

func (db *DB) SaveAttachment(name string, src io.Reader) (string, error) {
	id := primitive.NewObjectID()
	dst, err := db.bucket.OpenUploadStreamWithID(id, name)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Println(err)
		}
	}()

	if err := dst.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return "", err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

func (db *DB) GetAttachment(idHex string, dst io.Writer) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return err
	}

	src, err := db.bucket.OpenDownloadStream(id)
	if err != nil {
		return err
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Println(err)
		}
	}()

	if err := src.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}
