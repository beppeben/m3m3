package web

import (
	"github.com/beppeben/m3m3/domain"
	"github.com/beppeben/m3m3/utils"
	"html/template"
	"net/http"
	"encoding/json"
	"strconv"
	"fmt"
	"github.com/gorilla/context"
	log "github.com/Sirupsen/logrus"
)

type ItemInfo struct {
	*domain.ItemInfo
	fromuser 	string
}

func (info *ItemInfo) LocalImgUrl() string {
	if (info.Item.Id != 0) {
		return "images/" + info.Item.Tid + "-" + strconv.FormatInt(info.Item.Id, 10) + ".jpg"
	} else {
		return "images/temp/" + info.Item.Tid + ".jpg"
	}
}

func (info *ItemInfo) ImgUrl() string {
	return utils.GetServerUrl() + "/" + info.LocalImgUrl()
}

func (info *ItemInfo) BaseUrl() string {
	return utils.GetServerUrl()
}

func (info *ItemInfo) ItemUrl() string {
	root := utils.GetServerUrl() + "/item.html?"
	if (info.Item.Id != 0) {
		return root + "item_id=" + strconv.FormatInt(info.Item.Id, 10) + "%26item_tid=" + info.Item.Tid
	} else {
		return root + "item_tid=" + info.Item.Tid
	}
}

func (handler WebserviceHandler) ItemHTML (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	info, err := handler.processItemRequest(w, r)
	if err != nil {return}
	if (r.FormValue("item_id") == "" && info.Item.Id != 0) || r.FormValue("item_tid") == ""  {
		http.Redirect(w, r, utils.GetServerUrl() + "/item.html?item_id=" + 
			strconv.FormatInt(info.Item.Id, 10) + "&item_tid=" + info.Item.Tid, 301)
	}
	t, err := template.ParseFiles(utils.GetHTTPDir() + "item-template.html")
	if err != nil {
		panic("Bad Template: " + err.Error())
	}
	info.fromuser = username
	t.Execute(w, info)
}

func (handler WebserviceHandler) processItemRequest (w http.ResponseWriter, r *http.Request) (*ItemInfo, error) {
	item_tid := r.FormValue("item_tid")
	item_id := r.FormValue("item_id")
	comment_id := r.FormValue("comment_id")
	var id, cid int64
	var err error
	if item_id != "" {
		id, err = strconv.ParseInt(item_id, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "ERROR_BAD_ITEM_ID")
			log.Infof("Bad item id: %s", err)
			return nil, err
		}
	}
	if comment_id != "" {
		cid, err = strconv.ParseInt(comment_id, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "ERROR_BAD_COMMENT_ID")
			log.Infof("Bad comment id: %s", err)
			return nil, err
		}
	}
	info, err, msg := handler.itemInteractor.GetItemInfo(item_tid, id, cid)
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, msg)
		return nil, err
	}
	return &ItemInfo{ItemInfo: info}, err
}

func (handler WebserviceHandler) GetItemInfo (w http.ResponseWriter, r *http.Request) {
	info, err := handler.processItemRequest(w, r)
	if err != nil {return}
	enc := json.NewEncoder(w)
	enc.Encode(info)
}

func (handler WebserviceHandler) GetBestComments (w http.ResponseWriter, r *http.Request) {
	comments, err, msg := handler.itemInteractor.GetBestComments()
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(comments)
}

func (handler WebserviceHandler) GetItems (w http.ResponseWriter, r *http.Request) {
	items := handler.itemInteractor.GetZippedItems()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.Write(items)	
}

func (handler WebserviceHandler) PostLike (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	comment_id := r.FormValue("comment_id")
	id, err := strconv.ParseInt(comment_id, 10, 64)
	if err != nil {
		log.Infof("%s", err)
		fmt.Fprintf(w, "ERROR_BAD_ID")
		return
	}
	err, msg := handler.itemInteractor.AddLike(username, id)
	if err != nil {
		log.Infof("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	fmt.Fprintf(w, "OK")
}

func (handler WebserviceHandler) DeleteComment (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	comment_id := r.FormValue("comment_id")
	id, err := strconv.ParseInt(comment_id, 10, 64)
	if err != nil {
		log.Infof("%s", err)
		fmt.Fprintf(w, "ERROR_BAD_ID")
		return
	}
	item_id, err, msg := handler.itemInteractor.DeleteComment(username, id)
	if err != nil {
		log.Warnf("Error deleting comment: %s", err)
		fmt.Fprintf(w, msg)
		return
	}
	if item_id != 0 {
		http.Redirect(w, r, utils.GetServerUrl() + "/item.html?item_id=" + strconv.FormatInt(item_id, 10), 301)
	} else {
		http.Redirect(w, r, utils.GetServerUrl(), 301)
	}
}

func (handler WebserviceHandler) PostComment (w http.ResponseWriter, r *http.Request) {
	username := context.Get(r, "user").(string)
	text := r.PostFormValue("comment")
	item_tid := r.PostFormValue("item_tid")
	item_id := r.PostFormValue("item_id")
	var id int64
	var err error
	if item_id != "" {
		id, err = strconv.ParseInt(item_id, 10, 64)
		if err != nil {
			log.Infof("Bad id: %s", err)
			fmt.Fprintf(w, "ERROR_BAD_ID")
			return
		}
	}	
	comment_id, err, msg := handler.itemInteractor.AddComment(username, text, item_tid, id)
	if err != nil {
		log.Warnf("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
	http.Redirect(w, r, utils.GetServerUrl() + "/item.html?item_id=" + 
			item_id + "&comment_id=" + strconv.FormatInt(comment_id, 10), 301)
	//fmt.Fprintf(w, "OK")
}