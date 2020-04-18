package database

import (
	"bytes"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type Attachment struct {
	Id          interface{} `json:"id" bson:"_id,omitempty"`
	Filename    string      `json:"filename"`
	Size        int64       `json:"size"`
	ContentType string      `json:"content_type" bson:"content_type"`
}

func (a *Attachment) Open() (io.ReadCloser, error) {
	file, err := os.OpenFile(a.savedFilename(), os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("open attachment file: %v", err)
	}
	return file, nil
}

func (a *Attachment) saveFile(src io.Reader) (err error) {
	file, err := os.OpenFile(a.savedFilename(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open attachment file: %v", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			err = fmt.Errorf("close attachment file: %v", err)
		}
	}()

	a.Size, err = io.Copy(file, src)
	if err != nil {
		return fmt.Errorf("copy attachment to file: %v", err)
	}
	return nil
}

func (a *Attachment) deleteFile() error {
	if err := os.Remove(a.savedFilename()); err != nil {
		return fmt.Errorf("delete attachment file: %v", err)
	}
	return nil
}

func (a *Attachment) savedFilename() string {
	return path.Join(dataDir, a.Id.(primitive.ObjectID).Hex()+".attachment")
}

func (d *Database) NewAttachment(filename, contentType string, src io.Reader) (string, error) {
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
	if _, err := d.attachments.InsertOne(newCtx(), attachment); err != nil {
		return "", fmt.Errorf("insert attachment: %v", err)
	}
	return oid.Hex(), nil
}

func (d *Database) Attachment(id string) (*Attachment, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("parse object id: %w", ErrNotFound)
	}

	cursor := d.attachments.FindOne(newCtx(), bson.M{"_id": oid})

	var attachment Attachment
	if err := cursor.Decode(&attachment); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("decode attachment: %v", err)
	}
	return &attachment, nil
}

func (d *Database) Attachments() ([]Attachment, error) {
	ctx := newCtx()
	cursor, err := d.attachments.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find attachments: %v", err)
	}

	attachments := make([]Attachment, 0)
	if err := cursor.All(ctx, &attachments); err != nil {
		return nil, fmt.Errorf("decode attachments: %v", err)
	}
	return attachments, nil
}

func (d *Database) DeleteAttachment(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("parse object id: %w", ErrNotFound)
	}

	res := d.attachments.FindOneAndDelete(newCtx(), bson.M{
		"_id": oid,
	}, options.FindOneAndDelete().SetProjection(bson.D{
		{"_id", 1},
	}))

	var attachment Attachment
	if err := res.Decode(&attachment); err != nil {
		return fmt.Errorf("decode attachment: %v", err)
	}
	return attachment.deleteFile()
}

func (d *Database) DeleteAttachments() error {
	_, err := d.attachments.DeleteMany(newCtx(), bson.M{})
	if err != nil {
		return fmt.Errorf("delete attachments: %v", err)
	}

	attachments, err := filepath.Glob(filepath.Join(dataDir, "*.attachment"))
	if err != nil {
		panic(err)
	}
	for _, filename := range attachments {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("delete attachment file: %v", err)
		}
	}
	return nil
}
