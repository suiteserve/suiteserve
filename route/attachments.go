package route

import (
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"log"
	"net/http"
)

func (res *res) attachmentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id, ok := mux.Vars(r)["attachmentId"]
		if !ok {
			http.Error(w, "No ID provided", http.StatusBadRequest)
			return
		}

		if err := res.db.GetAttachment(id, w); err == gridfs.ErrFileNotFound {
			http.Error(w, "Attachment not found", http.StatusNotFound)
		} else if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
	}
}

func (res *res) attachmentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		src, header, err := r.FormFile("file")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := src.Close(); err != nil {
				log.Println(err)
			}
		}()

		id, err := res.db.SaveAttachment(header.Filename, src)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		loc, err := res.router.Get("attachment").URL("attachmentId", id)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", loc.String())
		w.WriteHeader(http.StatusCreated)
	}
}
