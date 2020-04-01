package route

import (
	"log"
	"net/http"
)

func (res *res) attachmentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
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
