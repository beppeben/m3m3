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
	"strconv"
	//"compress/gzip"
)

func getItems (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.Write(crawler.GetZippedItems())
	
}

func postComment (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	text := r.PostFormValue("comment")
	img_url := r.PostFormValue("img_url")
	item_id := r.PostFormValue("item_id")
	log.Printf("[SERV] Posting comment user: %s, url: %s, id: %s", username, img_url, item_id)
	if text == "" || (img_url == "" && item_id == "") {
		fmt.Fprintf(w, "ERROR_FORMAT")
		return
	} 
	var item *Item
	var err error
	var id int64
	if item_id != "" {
		log.Println("id not null")
		id, err = strconv.ParseInt(item_id, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "ERROR_BAD_ID")
			return
		}
		item, err = db.FindItemById(id)
		if err != nil {
			fmt.Fprintf(w, "ERROR_NOITEM")
			return
		}
	} else if img_url != "" {
		log.Println("url not null")
		item, err = db.FindItemByUrl(img_url)
		if err != nil {
			log.Println("item not in db")
			//item does not exist in the db
			var ok bool
			item, ok = crawler.GetItemByUrl (img_url)
			//avoid comments to unmanaged items, unless they're done by id
			if !ok {				
				fmt.Fprintf(w, "ERROR_UNMANAGED")
				return
			}
			//create the item in the db
			id, err = db.InsertItem(img_url, item.Title)	
			if err != nil {
				fmt.Fprintf(w, "ERROR_DB")
				log.Printf("[SERV] Database error: %s", err)
				return
			}
			//notify the item manager about the new item id
			crawler.NotifyItemId (img_url, id)
		} else {
			log.Println("item in db")
		}
		go SaveImageIfNeeded(item)
	} 
	
	comment := &Comment{Item_id: item.Id, Time: time.Now(), Text: text, Author: username}
	err = db.InsertComment (comment)
	if err != nil {
		fmt.Fprintf(w, "ERROR_DB")
		log.Printf("[SERV] Database error: %s", err)
		return
	}
	crawler.NotifyComment(comment)
	fmt.Fprintf(w, "OK")

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
