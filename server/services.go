package server

import (
	"fmt"
	"net/http"
	"encoding/json"
	crawler "github.com/beppeben/m3m3/crawler"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome!")
}


func getImageUrls (w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(crawler.Img_urls)
}
