package web

import (
	"github.com/beppeben/m3m3/utils"
	"net/http"
	"fmt"
	log "github.com/Sirupsen/logrus"
)


func (handler WebserviceHandler) DeployFront (w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("bundle")
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "ERROR_BAD_FILE")
		return
	}
	err = utils.ExtractZipToHttpDir(file, r.ContentLength)
	if err == nil {
		fmt.Fprintf(w, "OK")
	} else {
		log.Warnf("%s", err)
		fmt.Fprintf(w, "ERROR")
	}
}