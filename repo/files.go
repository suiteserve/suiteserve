package repo

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// -rw-------
const filePerm = 0600

type fileRepo struct {
	pattern string
}

func newFileRepo(pattern string) *fileRepo {
	return &fileRepo{pattern}
}

func (r *fileRepo) deleteAll() error {
	filenames, err := filepath.Glob(r.pattern)
	if err != nil {
		panic(err)
	}
	for _, filename := range filenames {
		if err := os.Remove(filename); err != nil {
			return err
		}
	}
	return nil
}

type fileAccessor struct {
	*fileRepo
	id string
}

func (r *fileRepo) newFileAccessor(id string) *fileAccessor {
	return &fileAccessor{r, id}
}

func (a *fileAccessor) Open() (io.ReadCloser, error) {
	// Open the file as read-only.
	return os.OpenFile(a.filename(), os.O_RDONLY, filePerm)
}

func (a *fileAccessor) filename() string {
	return strings.Replace(a.pattern, "*", a.id, 1)
}

func (a *fileAccessor) save(src io.Reader) (int64, error) {
	// Create a new non-existent file as write-only.
	file, err := os.OpenFile(a.filename(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, filePerm)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(file, src)
	// Always close the file. If there was an error during `Copy()`, still
	// return that, otherwise return the `Close()` error.
	if closeErr := file.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	// If there was an error anywhere, delete the file.
	if err != nil {
		a.delete()
		return n, err
	}
	return n, nil
}

func (a *fileAccessor) delete() {
	err := os.Remove(a.filename())
	if err != nil {
		log.Println(err)
	}
}

type attachmentFile struct {
	*fileAccessor
	info *AttachmentInfo
}

func (r *fileRepo) newAttachmentFile(info *AttachmentInfo) *attachmentFile {
	return &attachmentFile{r.newFileAccessor(info.Id), info}
}

func (f *attachmentFile) Info() *AttachmentInfo {
	return f.info
}

func (f *attachmentFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.info)
}
