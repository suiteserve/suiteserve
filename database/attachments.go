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

type AttachmentMetadata struct {
	ContentType string `json:"content_type" bson:"contentType"`
}

type Attachment struct {
	*gridfs.DownloadStream

	Id                 string `json:"id" bson:"_id,omitempty"`
	Name               string `json:"name" bson:"filename"`
	Size               int64  `json:"size" bson:"length"`
	AttachmentMetadata `bson:"metadata"`
}

func (d *Database) NewAttachment(name, contentType string, src io.Reader) (string, error) {
	oid := primitive.NewObjectID()

	// Sniff content type.
	var buf bytes.Buffer
	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(name))
	}
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

func (d *Database) Attachment(id string) (*Attachment, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrNotFound
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
		return nil, fmt.Errorf("failed to traverse GridFS cursor: %v", cursor.Err())
	}

	attachment := &Attachment{
		DownloadStream: src,
	}
	if err := cursor.Decode(attachment); err != nil {
		return nil, fmt.Errorf("failed to decode GridFS cursor: %v", err)
	}

	return attachment, nil
}

func (d *Database) AllAttachments() ([]Attachment, error) {
	cursor, err := d.mgoBucket.Find(bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find within GridFS: %v", err)
	}
	defer func() {
		if err := cursor.Close(newCtx()); err != nil {
			log.Printf("failed to close GridFS cursor: %v\n", err)
		}
	}()

	attachments := make([]Attachment, 0)
	if err := cursor.All(newCtx(), &attachments); err != nil {
		return nil, fmt.Errorf("failed to traverse and decode GridFS cursor: %v", err)
	}
	return attachments, nil
}

func (d *Database) DeleteAttachment(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNotFound
	}

	if err := d.mgoBucket.Delete(oid); err == gridfs.ErrFileNotFound {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("failed to delete within GridFS: %v", err)
	}
	return nil
}

func (d *Database) DeleteAllAttachments() error {
	if err := d.mgoBucket.Drop(); err != nil {
		return fmt.Errorf("failed to drop GridFS bucket: %v", err)
	}
	return nil
}
