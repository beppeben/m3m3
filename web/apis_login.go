package web

import (
	"net/http"
	"github.com/gorilla/context"
	"fmt"
	log "github.com/Sirupsen/logrus"
)


func (handler WebserviceHandler) Login (w http.ResponseWriter, r *http.Request) {
	nameOrMail := r.PostFormValue("name_email")
	pass := r.PostFormValue("pass")
	name, err, msg := handler.ILogin(nameOrMail, pass)
	if err != nil {
		log.Infof("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	handler.setCookies(name, w)
	fmt.Fprintf(w, "OK")
}

func (handler WebserviceHandler) Logout (w http.ResponseWriter, r *http.Request) {
	token := context.Get(r, "token")
	err, msg := handler.ILogout(token.(string))
	if err != nil {
		log.Infof("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	fmt.Fprintf(w, "OK")
}

func (handler WebserviceHandler) ConfirmEmail (w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	name, err, msg := handler.IConfirmEmail(token)
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	log.Infof("User %s confirmed!", name)
	handler.setCookies(name, w)
	http.Redirect(w, r, handler.config.GetServerUrl(), 301)
}

func (handler WebserviceHandler) Register (w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	pass := r.PostFormValue("pass")
	email := r.PostFormValue("email")
	temptoken, err, msg := handler.IRegister(&User{Name: name, Pass: pass, Email: email})
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	fmt.Fprintf(w, "OK")
	go func() {
		url := handler.config.GetServerUrl() + 
			"/services/confirm?token=" + temptoken
		err = handler.eutils.SendEmail(email, "Welcome to m3m3", "To confirm your email address, " +
			"please click on the following link:\r\n" + url)
		if err != nil {
			handler.repo.DeleteTempToken(temptoken)
		}
	}()
}