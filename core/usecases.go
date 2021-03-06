package core

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	. "github.com/beppeben/m3m3/domain"
	"github.com/beppeben/m3m3/utils"
	"time"
)

type ItemRepository interface {
	GetBestComments() ([]*Item, error)
	GetItemById(id int64) (*Item, error)
	GetCommentById(id int64) (*Comment, error)
	DeleteComment(id int64) error
	DeleteItem(id int64) error
	InsertItem(item *Item) (int64, error)
	GetCommentsByItem(item_id int64, comment_id int64) ([]*Comment, error)
	InsertLike(username string, comment_id int64) (*Comment, error)
	InsertComment(comment *Comment) error
}

type SysUtils interface {
	PersistTempImage(item_tid string, id int64) error
	DeleteImage(item_id int64, item_tid string) error
}

type ItemInteractor struct {
	itemRepo    ItemRepository
	itemManager *utils.IManager
	utils       SysUtils
}

func NewItemInteractor(repo ItemRepository, manager *utils.IManager, u SysUtils) *ItemInteractor {
	return &ItemInteractor{itemRepo: repo, itemManager: manager, utils: u}
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
		id, err = in.itemRepo.InsertItem(item)
		if err != nil {
			return nil, err, "ERROR_DB"
		}
		//notify the item manager about the new item id
		in.itemManager.NotifyItemId(item_tid, id)
		err = in.utils.PersistTempImage(item_tid, id)
		if err != nil {
			return nil, err, "ERROR_DB"
		}
	} else {
		return nil, errors.New("Error: empty ids for item"), "ERROR_FORMAT"
	}
	return item, nil, ""

}

func (in *ItemInteractor) GetZippedItems() []byte {
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

func (in *ItemInteractor) AddComment(username string, text string, item_tid string, item_id int64) (int64, error, string) {
	if text == "" || (item_tid == "" && item_id == 0) {
		return 0, errors.New("Error: Empty fields"), "ERROR_FORMAT"
	}
	//create item if it does not exist
	item, err, msg := in.RetrieveItem(item_tid, item_id, true)
	if err != nil {
		return 0, err, msg
	}
	comment := &Comment{Item_id: item.Id, Time: time.Now(), Text: text, Author: username, Likes: 0}
	err = in.itemRepo.InsertComment(comment)
	if err != nil {
		return 0, err, "ERROR_DB"
	}
	log.Info("comment inserted in db")
	in.itemManager.NotifyComment(comment)
	return comment.Id, nil, ""
}

func (in *ItemInteractor) DeleteComment(username string, comment_id int64) (int64, error, string) {
	comment, err := in.itemRepo.GetCommentById(comment_id)
	if err != nil {
		return 0, fmt.Errorf("Error: nonexisting comment id %d", comment_id), "ERROR_NOCOMMENT"
	}
	if comment.Author != username {
		return 0, fmt.Errorf("Error: user %s cannot delete comment %d", username, comment_id), "ERROR_UNAUTHORIZED"
	}
	//delete comment from db
	err = in.itemRepo.DeleteComment(comment_id)
	if err != nil {
		return 0, err, "ERROR_DB"
	}
	//delete item from manager if present and showing the same comment
	item, ok := in.itemManager.GetItemById(comment.Item_id)
	if ok && item.BestComment != nil && item.BestComment.Id == comment_id {
		in.itemManager.Remove(item)
	}
	//if no comments left for the item, remove it from the db so it can reappear in the feed later on
	comments, err := in.itemRepo.GetCommentsByItem(comment.Item_id, 0)
	if err != nil {
		return 0, err, "ERROR_DB"
	}
	if len(comments) == 0 {
		item, err = in.itemRepo.GetItemById(comment.Item_id)
		if err != nil {
			return 0, err, "ERROR_DB"
		}
		err = in.itemRepo.DeleteItem(comment.Item_id)
		if err != nil {
			return 0, err, "ERROR_DB"
		}
		err = in.utils.DeleteImage(item.Id, item.Tid)
		if err != nil {
			return 0, err, "ERROR_DB"
		}
		return 0, nil, ""
	}
	return comment.Item_id, nil, ""
}
