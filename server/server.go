package server

import (
	"log"
	"net/http"
	"time"
	"fmt"
	"github.com/beppeben/m3m3/db"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"io"
	"strings"
	"compress/gzip"
)

func init() {
	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler)
	authHandlers := commonHandlers.Append(authHandler)
	
	router := NewRouter()
	router.Get("/services/items", commonHandlers.ThenFunc(getItems))
	router.Post("/services/register", commonHandlers.ThenFunc(register))
	router.Post("/services/login", commonHandlers.ThenFunc(login))
	router.Get("/services/confirm", commonHandlers.ThenFunc(confirmEmail))
	router.Get("/services/itemInfo", commonHandlers.Append(gzipJsonHandler).ThenFunc(getItemInfo))
	router.Get("/services/logout", authHandlers.ThenFunc(logout))
	router.Get("/services/like", authHandlers.ThenFunc(postLike))
	router.Post("/services/comment", authHandlers.ThenFunc(postComment))
	router.Post("/services/deployFront", commonHandlers.ThenFunc(deployFront))
	log.Println("[SERV] Server ready to accept requests")
	http.ListenAndServe(":8080", router)
}

func Start_Server() {}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[SERV] Server Error: %+v", err)
				http.Error(w, http.StatusText(500) + ": " + fmt.Sprintf("%v", err), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}


func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//useful to test the frontend locally, remove in prod
		w.Header().Add("Access-Control-Allow-Origin", "*")
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[SERV] %s request to %q: time %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}


func authHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Printf("[SERV] Token cookie not sent: %s", err)
			http.Error(w, http.StatusText(401), 401)
			return
		}
		token := cookie.Value
		name, err := db.FindUserNameByToken(token)
		if err != nil {
			log.Printf("[SERV] Invalid token: %s", err)
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

func gzipJsonHandler(next http.Handler) http.Handler {
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


type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

func NewRouter() *router {
	return &router{httprouter.New()}
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}
