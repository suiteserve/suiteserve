package persist

import (
	"bytes"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
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

func (db *DB) GetAttachment(idHex string, dst http.ResponseWriter) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return err
	}

	cursor, err := db.bucket.Find(bson.M{"_id": id})
	if err != nil {
		return err
	}
	defer func() {
		if err := cursor.Close(newCtx()); err != nil {
			log.Println(err)
		}
	}()

	if ok := cursor.Next(newCtx()); !ok {
		return cursor.Err()
	}

	type gridfsFile struct {
		Name   string `bson:"filename"`
		Length int64  `bson:"length"`
	}
	var file gridfsFile
	if err := cursor.Decode(&file); err != nil {
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

	dst.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", strconv.Quote(file.Name)))
	dst.Header().Set("Content-Length", strconv.FormatInt(file.Length, 10))
	contentType := mime.TypeByExtension(path.Ext(file.Name))
	if contentType == "" {
		var b bytes.Buffer
		if _, err := io.CopyN(&b, src, 512); err != nil {
			return err
		}

		dst.Header().Set("Content-Type", http.DetectContentType(b.Bytes()))
		if _, err := io.Copy(dst, &b); err != nil {
			return err
		}
	} else {
		dst.Header().Set("Content-Type", contentType)
	}
	_, err = io.Copy(dst, src)
	return err
}
