package core

import (
	. "github.com/beppeben/m3m3/domain"
	"github.com/beppeben/m3m3/utils"
	"errors"
	"time"
	log "github.com/Sirupsen/logrus"
)

type ItemRepository interface { 
	GetBestComments() ([]*Item, error)
	GetItemById(id int64) (*Item, error)
	InsertItem(url, title, source string) (int64, error)
	GetCommentsByItem(item_id int64, comment_id int64) ([]*Comment, error)
	InsertLike(username string, comment_id int64) (*Comment, error)
	InsertComment(comment *Comment) error	
}


type ItemInteractor struct { 
	itemRepo 		ItemRepository 
	itemManager 		*utils.IManager
}

func NewItemInteractor(repo ItemRepository, manager *utils.IManager) *ItemInteractor {
	return &ItemInteractor{itemRepo: repo, itemManager: manager}
}


func (in *ItemInteractor) GetBestComments() ([]*Item, error, string) {
	items, err := in.itemRepo.GetBestComments()
	if err != nil {
		return nil, err, "ERROR_DB"
	} else {
		return items, nil, ""
	}
}

func (in *ItemInteractor) RetrieveItem(item_tid string, id int64, create bool) (*Item, error, string) {
	var item *Item
	var err error
	if id != 0 {		
		item, err = in.itemRepo.GetItemById(id)
		if err != nil {
			return nil, err, "ERROR_NOITEM"
		}
	} else if item_tid != "" {
		var ok bool
		item, ok = in.itemManager.GetItemByTid(item_tid)
		//avoid comments to unmanaged items, unless they're done by id
		if !ok {				
			return nil, err, "ERROR_UNMANAGED"
		}
		if !create || item.Id != 0 {
			return item, nil, ""
		}
		//create the item in the db
		id, err = in.itemRepo.InsertItem(item.Url, item.Title, item.Source)
		if err != nil {
			return nil, err, "ERROR_DB"
		}
		//notify the item manager about the new item id
		in.itemManager.NotifyItemId(item_tid, id)
		err = utils.PersistTempImage(item_tid, id)
		if err != nil {
			return nil, err, "ERROR_DB"
		}
	} else {
		return nil, errors.New("Error: empty ids for item"), "ERROR_FORMAT"
	}
	return item, nil, ""

}

func (in *ItemInteractor) GetZippedItems() ([]byte) {
	return in.itemManager.GetZippedJson()
}

//get item with comments list, by id or tid (optionally a comment_id to appear on top)
func (in *ItemInteractor) GetItemInfo(item_tid string, item_id int64, comment_id int64) (*ItemInfo, error, string) {
	item, err, msg := in.RetrieveItem(item_tid, item_id, false)
	if err != nil {
		return nil, err, msg
	}
	var comments []*Comment
	if item.Id != 0 {		
		comments, err = in.itemRepo.GetCommentsByItem(item.Id, comment_id)
		if err != nil {
			return nil, err, "ERROR_DB"
		}
	}	
	return &ItemInfo{Item: item, Comments: comments}, err, ""
}

func (in *ItemInteractor) AddLike(username string, comment_id int64) (error, string) {
	comment, err := in.itemRepo.InsertLike(username, comment_id)
	if err != nil {
		return err, "ERROR_BAD_REQUEST"
	}
	in.itemManager.NotifyComment(comment)
	return nil, ""
}

func (in *ItemInteractor) AddComment(username string, text string, item_tid string, item_id int64) (error, string) {
	if text == "" || (item_tid == "" && item_id == 0) {
		return errors.New("Error: Empty fields"), "ERROR_FORMAT"
	} 	
	log.Info("retrieving item")
	//create item if it does not exist
	item, err, msg := in.RetrieveItem(item_tid, item_id, true)
	if err != nil {
		return err, msg
	}
	log.Info("item retrieved")
	if item == nil {
		log.Info("item is null")
	}
	comment := &Comment{Item_id: item.Id, Time: time.Now(), Text: text, Author: username, Likes: 0}
	err = in.itemRepo.InsertComment(comment)
	if err != nil {
		return err, "ERROR_DB"
	}
	log.Info("comment inserted in db")
	in.itemManager.NotifyComment(comment)
	return nil, ""
}