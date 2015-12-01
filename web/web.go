package web

import (
	"github.com/beppeben/m3m3/utils"
	"github.com/beppeben/m3m3/core"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"fmt"
	"time"
	"strings"
	"github.com/gorilla/context"
	"github.com/justinas/alice"
	cache "github.com/patrickmn/go-cache"
	log "github.com/Sirupsen/logrus"
)

type Repository interface { 
	InsertAccessToken(token string, name string, expire time.Time) error
	GetUserNameByToken(token string) (string, error)	
	GetUserByEmail(email string) (*User, error)
	GetUserByName(name string) (*User, error)
	DeleteAccessToken(token string) error
	InsertUserFromTempToken(token string) (*User, error)
	InsertTempToken(user *User, token string, expire time.Time) error
	DeleteTempToken(token string) error
}


type WebserviceHandler struct { 
	itemInteractor	*core.ItemInteractor
	repo 			Repository
	frouter			*http.ServeMux
	mrouter			*router
	usercache		*cache.Cache
}

type router struct {
	*httprouter.Router
}

func NewWebHandler(it *core.ItemInteractor, repo Repository) *WebserviceHandler {
	return &WebserviceHandler{itemInteractor: it, repo: repo, usercache: cache.New(24*time.Hour, time.Hour)}
}

func (h WebserviceHandler) StartServer() {	
	commonHandlers := alice.New(context.ClearHandler, h.LoggingHandler, h.RecoverHandler, h.NoCacheHandler)
	authHandlers := commonHandlers.Append(h.AuthHandler)
	h.mrouter = NewRouter()
	h.frouter = http.NewServeMux()
	h.frouter.Handle("/", http.FileServer(http.Dir(utils.GetHTTPDir())))
	
	h.mrouter.Get("/item.html", authHandlers.ThenFunc(h.ItemHTML))
	h.mrouter.Get("/services/items", commonHandlers.ThenFunc(h.GetItems))
	h.mrouter.Get("/services/bestComments", commonHandlers.Append(h.GzipJsonHandler).ThenFunc(h.GetBestComments))
	h.mrouter.Post("/services/register", commonHandlers.ThenFunc(h.Register))
	h.mrouter.Post("/services/login", commonHandlers.ThenFunc(h.Login))
	h.mrouter.Get("/services/confirm", commonHandlers.ThenFunc(h.ConfirmEmail))
	h.mrouter.Get("/services/itemInfo", commonHandlers.Append(h.GzipJsonHandler).ThenFunc(h.GetItemInfo))
	h.mrouter.Get("/services/logout", authHandlers.ThenFunc(h.Logout))
	h.mrouter.Get("/services/like", authHandlers.ThenFunc(h.PostLike))
	h.mrouter.Post("/services/comment", authHandlers.ThenFunc(h.PostComment))
	h.mrouter.Get("/services/deletecomment", authHandlers.ThenFunc(h.DeleteComment))
	h.mrouter.Post("/services/deployFront", commonHandlers.Append(h.BasicAuth).ThenFunc(h.DeployFront))
		
	var err error
	
	r := http.NewServeMux()
	r.HandleFunc("/", h.FrontHandler)
	go func() {
		log.Infof("Server launched on port %s", utils.GetServerPort())	
    		err = http.ListenAndServe(":" + utils.GetServerPort(), r)		
		if err != nil {
			panic(err.Error())
		}		
	}()
	
	
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

func (handler WebserviceHandler) FrontHandler(w http.ResponseWriter, r *http.Request) {
	
	//if (strings.HasPrefix(r.URL.Path, "/item.html")) {
	//	handler.ItemHTML(w, r)
	//}
	if strings.HasPrefix(r.URL.Path, "/services/") || strings.HasPrefix(r.URL.Path, "/item.html") {
		ip := strings.Split(r.RemoteAddr,":")[0] 
		_, found := handler.usercache.Get(ip)
		if (!found) {
			handler.usercache.Set(ip, "", cache.DefaultExpiration)
			log.Infof("New user: %s. Total daily users: %d", ip, handler.usercache.ItemCount())			
		}
		handler.mrouter.ServeHTTP(w,r)
	} else {
		if strings.HasPrefix(r.URL.Path, "/images/") {
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate", 31556926))
		}		
		handler.frouter.ServeHTTP(w, r)
	}
	
}



func (handler WebserviceHandler) setCookies (username string, w http.ResponseWriter) {
	if username != "" {
		log.Debugf("Setting cookies for %s", username)
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

