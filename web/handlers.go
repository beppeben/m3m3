package web

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"io"
	"net/http"
	"strings"
	"time"
)

func (handler WebserviceHandler) BasicAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authError := func() {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"m3m3\"")
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		}

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			authError()
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			authError()
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || !handler.ValidateAdmin(pair[0], pair[1]) {
			authError()
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (handler WebserviceHandler) ValidateAdmin(username, password string) bool {
	if username == "admin" && password == handler.config.GetAdminPass() {
		return true
	}
	return false
}

func (handler WebserviceHandler) RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Warnf("%v", err)
				http.Error(w, http.StatusText(500)+": "+fmt.Sprintf("%v", err), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (handler WebserviceHandler) NoCacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

func (handler WebserviceHandler) LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//useful to test the frontend locally, remove in prod
		w.Header().Add("Access-Control-Allow-Origin", "*")
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Infof("%s request to %q: time %v", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func (handler WebserviceHandler) AuthHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			//serve even if cookie not set
			next.ServeHTTP(w, r)
			return
		}
		token := cookie.Value
		name, err := handler.repo.GetUserNameByToken(token)
		if err != nil {
			log.Infof("%v", err)
			http.Error(w, http.StatusText(401), 401)
			return
		}
		context.Set(r, "user", name)
		context.Set(r, "token", token)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (handler WebserviceHandler) GzipJsonHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzr, r)
	}
	return http.HandlerFunc(fn)
}

/*
func (handler WebserviceHandler) SetCookiesHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		user := context.Get(r, "user")

		if user != nil {
			log.Debugf("Setting cookies for %s", user)
			username := user.(string)
			token := utils.RandString(32)
			expire := time.Now().AddDate(0,2,0)
			cookie := http.Cookie{Name: "token", Value: token, Path: "/", Expires: expire}
			http.SetCookie(w, &cookie)
			cookie = http.Cookie{Name: "name", Value: username, Path: "/", Expires: expire}
			http.SetCookie(w, &cookie)
			err := handler.repo.InsertAccessToken(token, username, expire)
			if err != nil {
				log.Warnf("%v", err)
			}
		} else {
			log.Debugln("Cannot set cookies to null user name")
		}
	}

	return http.HandlerFunc(fn)
}
*/
