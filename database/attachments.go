package database

import (
	"bytes"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"path/filepath"
)

type Attachment struct {
	Id          interface{} `json:"id" bson:"_id,omitempty"`
	Filename    string      `json:"filename"`
	Size        int64       `json:"size"`
	ContentType string      `json:"content_type" bson:"content_type"`
}

func (a *Attachment) OpenFile() (io.ReadCloser, error) {
	return openFile(a.savedFilename())
}

func (a *Attachment) saveFile(src io.Reader) error {
	var err error
	a.Size, err = createFile(a.savedFilename(), src)
	return err
}

func (a *Attachment) deleteFile() error {
	return deleteFile(a.savedFilename())
}

func (a *Attachment) savedFilename() string {
	return path.Join(dataDir, a.Id.(primitive.ObjectID).Hex()+".attachment")
}

func (d *WithContext) NewAttachment(filename, contentType string, src io.Reader) (string, error) {
	// Sniff content type.
	var buf bytes.Buffer
	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(filename))
	}
	if contentType == "" {
		if _, err := io.CopyN(&buf, src, 512); err != nil {
			return "", err
		}
		contentType = http.DetectContentType(buf.Bytes())
	}

	oid := primitive.NewObjectID()
	attachment := Attachment{
		Id:          oid,
		Filename:    filename,
		ContentType: contentType,
	}
	if err := attachment.saveFile(io.MultiReader(&buf, src)); err != nil {
		return "", err
	}

	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.attachments.InsertOne(ctx, attachment); err != nil {
		// If there was an error, delete the file. Always return the other
		// error; we'll just print out the `deleteFile()` error for information.
		if deleteErr := attachment.deleteFile(); deleteErr != nil {
			log.Printf("%v\n", deleteErr)
		}
		return "", err
	}
	return oid.Hex(), nil
}

func (d *WithContext) Attachment(id string) (*Attachment, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res := d.attachments.FindOne(ctx, bson.M{"_id": oid})
	var attachment Attachment
	if err := res.Decode(&attachment); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find and decode attachment: %v", err)
	}
	return &attachment, nil
}

func (d *WithContext) AllAttachments() ([]Attachment, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.attachments.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find all attachments: %v", err)
	}
	attachments := make([]Attachment, 0)
	if err := cursor.All(ctx, &attachments); err != nil {
		return nil, fmt.Errorf("decode all attachments: %v", err)
	}
	return attachments, nil
}

func (d *WithContext) DeleteAttachment(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.attachments.DeleteOne(ctx, bson.M{"_id": oid}); err != nil {
		return fmt.Errorf("delete attachment: %v", err)
	}
	return (&Attachment{Id: oid}).deleteFile()
}

func (d *WithContext) DeleteAllAttachments() error {
	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.attachments.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("delete all attachments: %v", err)
	}
	filenames, err := filepath.Glob(filepath.Join(dataDir, "*.attachment"))
	if err != nil {
		log.Panicln(err)
	}
	return deleteAllFiles(filenames)
}
