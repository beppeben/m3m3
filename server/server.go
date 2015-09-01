package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func init() {
	// db := sql.Open("postgres", "...")
	//appC := appContext{nil}
	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler)
	router := NewRouter()
	//router.Get("/admin", commonHandlers.Append(appC.authHandler).ThenFunc(appC.adminHandler))
	//router.Get("/about", commonHandlers.ThenFunc(aboutHandler))
	router.Get("/", commonHandlers.ThenFunc(indexHandler))
	router.Get("/images", commonHandlers.ThenFunc(getImageUrls))
	//router.Get("/teas/:id", commonHandlers.ThenFunc(appC.teaHandler))

	log.Println("[SERV] Server ready to accept requests")
	http.ListenAndServe(":8080", router)
}

func Start_Server() {}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[SERV] Server Error: %+v", err)
				http.Error(w, http.StatusText(500), 500)
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

type appContext struct {
	db *sql.DB
}

/*
func (c *appContext) authHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		user, err := map[string]interface{}{}, errors.New("test")
		// user, err := getUser(c.db, authToken)
		log.Println(authToken)

		if err != nil {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		context.Set(r, "user", user)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (c *appContext) adminHandler(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user")
	// Maybe other operations on the database
	json.NewEncoder(w).Encode(user)
}
*/

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
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
