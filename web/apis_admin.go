package web

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

func (handler WebserviceHandler) DeployFront(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("bundle")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "ERROR_BAD_FILE")
		return
	}
	err = handler.sutils.ExtractZipToHttpDir(file, r.ContentLength)
	if err == nil {
		fmt.Fprintf(w, "OK")
	} else {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "ERROR")
	}
}
