package server

import (
	"fmt"
	"net/http"
	"github.com/beppeben/m3m3/crawler"
	"github.com/beppeben/m3m3/db"
	. "github.com/beppeben/m3m3/utils"
	"log"
	"time"
	"github.com/gorilla/context"
	//"compress/gzip"
)

func getItems (w http.ResponseWriter, r *http.Request) {
	//json.NewEncoder(w).Encode(crawler.Img_urls)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	//fmt.Fprintf(w, crawler.GetItems())
	w.Write(crawler.GetZippedItems())
	
}

func getImageUrls (w http.ResponseWriter, r *http.Request) {
	
}

func login (w http.ResponseWriter, r *http.Request) {
	nameOrMail := r.PostFormValue("name_email")
	pass := r.PostFormValue("pass")
	if (nameOrMail == "" || pass == ""){
		fmt.Fprintf(w, "ERROR_FORMAT")
		return
	}
	var name, storedPass string
	if (ValidateEmail(nameOrMail)){
		//we are given an email address
		user, err := db.FindUserByEmail(nameOrMail)
		if (err != nil) {
			fmt.Fprintf(w, "ERROR_EMAIL_NOT_EXISTING")
			return
		} 		
		name = user.Name
		storedPass = user.Pass
	} else {
		//we are given a name
		user, err := db.FindUserByName(nameOrMail)
		if (err != nil) {
			fmt.Fprintf(w, "ERROR_NAME_NOT_EXISTING")
			return
		} 		
		name = nameOrMail
		storedPass = user.Pass
	}
	if (ComputeMd5(pass) != storedPass) {
		fmt.Fprintf(w, "ERROR_WRONG_PASS")
		return
	}
	
	setAuthCookies(w, name)
	fmt.Fprintf(w, "OK")
}

func setAuthCookies (w http.ResponseWriter, name string) {
	token := RandString(32)
	expire := time.Now().AddDate(0,2,0)
	//cookie := http.Cookie{"token", token, "/", "45.55.210.25", expire, expire.Format(time.UnixDate), 
	//	86400, false, false, "token=" + token, []string{"token=" + token}}
	cookie := http.Cookie{
		Name: "token", 
		Value: token,
		Path: "/",
		Expires: expire,		
	}	
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{
		Name: "name", 
		Value: name,
		Path: "/",
		Expires: expire,		
	}	
	http.SetCookie(w, &cookie)
	db.InsertAccessToken(token, name, expire)
}

func logout (w http.ResponseWriter, r *http.Request) {
	//token := r.FormValue("token")
	token := context.Get(r, "token")
	err := db.DeleteAccessToken(token.(string))
	if err != nil {
		log.Printf("[SERV] Database error: %s", err)
		fmt.Fprintf(w, "DB_ERROR")
		return
	}
	fmt.Fprintf(w, "OK")
}


func confirmEmail (w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	user, err := db.InsertUserFromTempToken(token)
	if err != nil {
		log.Printf("[SERV] Database error: %s", err)
		fmt.Fprintf(w, "DB_ERROR")
		return
	}
	setAuthCookies(w, user.Name)
	fmt.Fprintf(w, "OK")
}

func register (w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	pass := r.PostFormValue("pass")
	email := r.PostFormValue("email")
	
	if (name == "" || pass == "" || email == ""){
		fmt.Fprintf(w, "ERROR_FORMAT")
		return
	}	
	if (!ValidateEmail(email)){
		fmt.Fprintf(w, "ERROR_EMAIL_INVALID")
		return
	}	
	_, err := db.FindUserByName(name)
	if (err == nil) {
		fmt.Fprintf(w, "ERROR_NAME_USED")
		return
	} 
	_, err = db.FindUserByEmail(email)
	if (err == nil) {
		fmt.Fprintf(w, "ERROR_EMAIL_USED")
		return
	}
	
	pass = ComputeMd5(pass)	
	token := RandString(32)
	expire := time.Now().AddDate(0, 0, 1)
	//err = db.InsertUserToken(&User{Name: name, Pass: pass, Email: email}, token, expire)
	err = db.InsertTempToken(&User{Name: name, Pass: pass, Email: email}, token, expire)
	
	if err != nil {
		fmt.Fprintf(w, "ERROR_DB")
		log.Printf("[SERV] Database error: %s", err)	
		return
	}

	SendEmail(email, "Welcome to m3m3", "To confirm your email address, please click on the " + "following link:\r\n/token=" + token)
	
	fmt.Fprintf(w, "OK")
}
