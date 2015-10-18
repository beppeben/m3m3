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
	//"strings"
	//"compress/gzip"
	"encoding/json"
	"archive/zip"
	"os"
	"io"
	//"net/url"
)

type ItemInfo struct {
	It 			*Item		`json:"item,omitempty"`
	Comments		[]*Comment	`json:"comments,omitempty"`
}

func getItemInfo (w http.ResponseWriter, r *http.Request) {
	item_tid := r.FormValue("item_tid")
	item_id := r.FormValue("item_id")
	if item_tid == "" && item_id == "" {
		fmt.Fprintf(w, "ERROR_FORMAT")
		return
	} 
	item, err := retrieveItem (w, item_tid, item_id, false)
	if err != nil {
		return
	}
	var comments []*Comment
	if item.Id != 0 {
		comments, err = db.FindCommentsByItem(item.Id)
		if err != nil {
			fmt.Fprintf(w, "ERROR_DB")
			log.Printf("[SERV] Database error: %s", err)
			return
		}
	}	
	result := &ItemInfo{It: item, Comments: comments}
	enc := json.NewEncoder(w)
	enc.Encode(result)
}

func getBestComments (w http.ResponseWriter, r *http.Request) {
	comments, err := db.FindBestComments()
	if err != nil {
		fmt.Fprintf(w, "ERROR_DB")
		log.Printf("[SERV] Database error: %s", err)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(comments)
}

func getItems (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.Write(crawler.GetZippedItems())	
}

func retrieveItem (w http.ResponseWriter, item_tid string, item_id string, create bool) (*Item, error) {
	var item *Item
	var err error
	var id int64
	if item_id != "" {
		id, err = strconv.ParseInt(item_id, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "ERROR_BAD_ID")
			return nil, err
		}
		item, err = db.FindItemById(id)
		if err != nil {
			fmt.Fprintf(w, "ERROR_NOITEM")
			return nil, err
		}
	} else if item_tid != "" {

		var ok bool
		item, ok = crawler.GetItemByTid(item_tid)
		//avoid comments to unmanaged items, unless they're done by id
		if !ok {				
			fmt.Fprintf(w, "ERROR_UNMANAGED")
			return nil, err
		}
		if !create || item.Id != 0 {
			return item, nil
		}
		//create the item in the db
		id, err = db.InsertItem(item.Url, item.Title, item.Source)	
		if err != nil {
			fmt.Fprintf(w, "ERROR_DB")
			log.Printf("[SERV] Database error: %s", err)
			return nil, err
		}
		//notify the item manager about the new item id
		crawler.NotifyItemId (item_tid, id)
		
		//go SaveImageIfNeeded(item)
		err = PersistTempImage (item_tid, id)
		if err != nil {
			fmt.Fprintf(w, "ERROR_DB")
			log.Printf("[SERV] Database error: %s", err)
			return nil, err
		}
	} 
	return item, nil
}

func deployFront (w http.ResponseWriter, r *http.Request) {	
	/*
	username := context.Get(r, "user").(string)
	if username != GetAdmin() {
		http.Error(w, http.StatusText(401), 401)
		return
	}
	*/
	file, _, err := r.FormFile("bundle")
	if err != nil {
		fmt.Fprintf(w, "ERROR_BAD_FILE")
		return
	}
	reader, err := zip.NewReader(file, r.ContentLength)
	if err != nil {
    		fmt.Fprintf(w, "ERROR")
		return
	}   
	for _, zf := range reader.File {
    		dst, err := os.Create(GetHTTPDir() + zf.Name)
    		if err != nil {
			log.Printf("[SERV] OS error creating file: %s", err)
        		fmt.Fprintf(w, "ERROR")
			return
    		}
    		defer dst.Close()
    		src, err := zf.Open()
    		if err != nil {
        		log.Printf("[SERV] OS error opening file: %s", err)
        		fmt.Fprintf(w, "ERROR")
			return
    		}
    		defer src.Close()
    		io.Copy(dst, src)
	}   
	fmt.Fprintf(w, "OK")
}

func postLike (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	comment_id := r.FormValue("comment_id")
	id, err := strconv.ParseInt(comment_id, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "ERROR_BAD_ID")
		return
	}
	comment, err := db.InsertLike(username, id)
	if err != nil {
		fmt.Fprintf(w, "ERROR_BAD_REQUEST")
		log.Printf("[SERV] Database error: %s", err)
		return
	}
	crawler.NotifyComment(comment)
	fmt.Fprintf(w, "OK")
}

func postComment (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	text := r.PostFormValue("comment")
	item_tid := r.PostFormValue("item_tid")
	item_id := r.PostFormValue("item_id")
	if text == "" || (item_tid == "" && item_id == "") {
		fmt.Fprintf(w, "ERROR_FORMAT")
		return
	} 	
	item, err := retrieveItem (w, item_tid, item_id, true)
	if err != nil {
		return
	}
	comment := &Comment{Item_id: item.Id, Time: time.Now(), Text: text, Author: username, Likes: 0}
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
	http.Redirect(w, r, "http://45.55.210.25", 301)
	//fmt.Fprintf(w, "OK")
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

	SendEmail(email, "Welcome to m3m3", "To confirm your email address, please click on the " +
		"following link:\r\nhttp://45.55.210.25/services/confirm?token=" + token)

	fmt.Fprintf(w, "OK")
}
