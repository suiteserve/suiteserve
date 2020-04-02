package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
	"log"
	"time"
)

type Attachment struct {
	*gridfs.DownloadStream

	Name   string `bson:"filename"`
	Length int64  `bson:"length"`
}

func (d *Database) SaveAttachment(name string, src io.Reader) (string, error) {
	oid := primitive.NewObjectID()
	dst, err := d.mgoBucket.OpenUploadStreamWithID(oid, name)
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

	//if err := src.SetReadDeadline(time.Now().Add(timeout)); err != nil {
	//	return nil, err
	//}

	return attachment, nil

	//dst.Header().Set("Content-Disposition",
	//	fmt.Sprintf("attachment; filename=%s", strconv.Quote(file.Name)))
	//dst.Header().Set("Content-Length", strconv.FormatInt(file.Length, 10))
	//contentType := mime.TypeByExtension(path.Ext(file.Name))
	//if contentType == "" {
	//	var b bytes.Buffer
	//	if _, err := io.CopyN(&b, src, 512); err != nil {
	//		return err
	//	}
	//
	//	dst.Header().Set("Content-Type", http.DetectContentType(b.Bytes()))
	//	if _, err := io.Copy(dst, &b); err != nil {
	//		return err
	//	}
	//} else {
	//	dst.Header().Set("Content-Type", contentType)
	//}
	//_, err = io.Copy(dst, src)
	//return err
}
