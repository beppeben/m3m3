package domain

import (
	"time"
)

type Item struct {
	Tid         string   `json:"item_tid,omitempty"`
	Id          int64    `json:"item_id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Link        string   `json:"link,omitempty"`
	Source      string   `json:"source,omitempty"`
	Src         Source   `json:"-"`
	Ncomments   int      `json:"n_comm,omitempty"`
	BestComment *Comment `json:"b_comm,omitempty"`
	Url         string   `json:"img_url,omitempty"`
	Time        int64    `json:"-"`
	Score       int64    `json:"-"`
}

type Comment struct {
	Id      int64     `json:"id,omitempty"`
	Item_id int64     `json:"-"`
	Text    string    `json:"text,omitempty"`
	Author  string    `json:"author,omitempty"`
	Likes   int       `json:"likes"`
	Time    time.Time `json:"-"`
}

type Source interface {
	Increase()
	Decrease()
}

type ItemInfo struct {
	Item     *Item      `json:"item,omitempty"`
	Comments []*Comment `json:"comments,omitempty"`
}
