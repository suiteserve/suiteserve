package database

import (
	"bytes"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"time"
)

type Attachment struct {
	*gridfs.DownloadStream

	Name     string `bson:"filename"`
	Length   int64  `bson:"length"`
	Metadata struct {
		ContentType string `bson:"contentType"`
	} `bson:"metadata"`
}

func (d *Database) SaveAttachment(name string, src io.Reader) (string, error) {
	oid := primitive.NewObjectID()

	// Sniff content type.
	var buf bytes.Buffer
	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		if _, err := io.CopyN(&buf, src, 512); err != nil {
			return "", fmt.Errorf("failed to copy source to buffer: %v", err)
		}
		contentType = http.DetectContentType(buf.Bytes())
	}

	opts := options.GridFSUpload().SetMetadata(bson.M{
		"contentType": contentType,
	})

	dst, err := d.mgoBucket.OpenUploadStreamWithID(oid, name, opts)
	if err != nil {
		return "", fmt.Errorf("failed to open GridFS upload stream: %v", err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("failed to close GridFS upload stream: %v\n", err)
		}
	}()

	if err := dst.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return "", fmt.Errorf("failed to set GridFS upload deadline: %v", err)
	}

	_, err = io.Copy(dst, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to copy buffer to GridFS: %v", err)
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("failed to copy source to GridFS: %v", err)
	}

	return oid.Hex(), nil
}

func (d *Database) GetAttachment(id string) (*Attachment, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrBadId
	}

	src, err := d.mgoBucket.OpenDownloadStream(oid)
	if err == gridfs.ErrFileNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to open GridFS download stream: %v", err)
	}

	cursor, err := d.mgoBucket.Find(bson.M{"_id": oid})
	if err != nil {
		return nil, fmt.Errorf("failed to find within GridFS: %v", err)
	}
	defer func() {
		if err := cursor.Close(newCtx()); err != nil {
			log.Printf("failed to close GridFS cursor: %v\n", err)
		}
	}()

	if ok := cursor.Next(newCtx()); !ok {
		return nil, fmt.Errorf("failed to traverse GridFS cursor: %v",
			cursor.Err())
	}

	attachment := &Attachment{
		DownloadStream: src,
	}
	if err := cursor.Decode(attachment); err != nil {
		return nil, fmt.Errorf("failed to decode GridFS cursor: %v", err)
	}

	return attachment, nil
}
