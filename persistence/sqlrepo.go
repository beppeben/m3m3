package persistence

import (
	. "github.com/beppeben/m3m3/domain"
	"github.com/beppeben/m3m3/web"
	"database/sql"
	"time"
	"errors"
)

type DbHandler interface { 
	Conn() *sql.DB
	Transact(txFunc func(*sql.Tx) (interface{}, error)) (interface{}, error)
}

type SqlRepo struct {
	h 		DbHandler
}

func NewRepo(h DbHandler) *SqlRepo {
	return &SqlRepo{h}
}

func (r *SqlRepo) InsertComment (c *Comment) error {
	st := "INSERT INTO comments VALUES(?,?,?,?,?,?)"
	result, err := r.h.Conn().Exec(st, nil, c.Item_id, c.Time, c.Text, c.Author, c.Likes)
	if err == nil {
		id, err := result.LastInsertId()
		if err == nil {
			c.Id = id
		}
	}
	return err
}

func (r *SqlRepo) InsertLike (username string, comment_id int64) (comment *Comment, err error) {
	obj, err := r.h.Transact(func (tx *sql.Tx) (interface{}, error) {
		var d time.Time
		var t,a string
		var likes int
		var item_id int64
		st := "SELECT * FROM comments WHERE id = ?"		
		err = tx.QueryRow(st, comment_id).Scan(&comment_id, &item_id, &d, &t, &a, &likes)
		if err != nil {
			panic("The comment does not exist")
		} 
		comment := &Comment{Id: comment_id, Item_id: item_id, Text: t, Author: a, Likes: likes + 1, Time: d}
		st = "INSERT INTO likes VALUES(?,?)"		
		_, err = tx.Exec(st, username, comment_id)
		if err != nil {
			panic("Comment already liked")
		}
		st = "UPDATE comments SET likes = likes + 1 WHERE id = ?"		
		_, err = tx.Exec(st, comment_id)
		if err != nil {
			panic("Problem updating comment likes")
		}
		return comment, nil
	})
	if obj != nil {
		return obj.(*Comment), err
	} else {
		return  nil, err
	}
	
}

func (r *SqlRepo) InsertTempToken (user *web.User, tempToken string, t time.Time) error {
	st := "INSERT INTO temp_tokens VALUES(?,?,?,?,?)"	
	_, err := r.h.Conn().Exec(st, tempToken, user.Name, user.Email, user.Pass, t)
	return err
}

func (r *SqlRepo) InsertAccessToken (tempToken string, name string, t time.Time) error {
	st := "INSERT INTO access_tokens VALUES(?,?,?)"	
	_, err := r.h.Conn().Exec(st, tempToken, name, t)
	return err
}

func (r *SqlRepo) InsertItem (img_url, title, source string) (int64, error) {
	st := "INSERT INTO items VALUES(?,?,?,?)"	
	result, err := r.h.Conn().Exec(st, nil, img_url, title, source)
	if err != nil {
		return -1, err
	} else {
		return result.LastInsertId()
	}
}

func (r *SqlRepo) DeleteAccessToken (token string) error {
	st := "DELETE FROM access_tokens WHERE token = ?"	
	_, err := r.h.Conn().Exec(st, token)
	return err
}

func (r *SqlRepo) DeleteTempToken (token string) error {
	st := "DELETE FROM temp_tokens WHERE token = ?"	
	_, err := r.h.Conn().Exec(st, token)
	return err
}

func (r *SqlRepo) InsertUserFromTempToken (tempToken string) (*web.User, error) {
	var name, pass, email string
	var t time.Time
	var user *web.User
	st := "SELECT * FROM temp_tokens WHERE token = ?"	
	err := r.h.Conn().QueryRow(st, tempToken).Scan(&tempToken, &name, &email, &pass, &t)	
	if (err != nil) {
		return user, err
	}
	st = "DELETE FROM temp_tokens WHERE token = ?"	
	_, err = r.h.Conn().Exec(st, tempToken)
	if (err != nil) {
		return &web.User{}, err
	}
	if time.Now().After(t) {
		return &web.User{}, errors.New("expired token")
	}
	st = "INSERT INTO users VALUES(?,?,?)"	
	_, err = r.h.Conn().Exec(st, name, email, pass)
	if (err != nil) {
		return &web.User{}, err
	}
	user = &web.User{Name: name, Email: email, Pass: pass}
	return user, err
}

func (r *SqlRepo) GetUserNameByToken(token string) (string, error) {
	var name string
	var t time.Time
	st := "SELECT * FROM access_tokens WHERE token = ?"	
	err := r.h.Conn().QueryRow(st, token).Scan(&token, &name, &t)
	if err != nil {
		return name, err		
	} else {
		if time.Now().Before(t) {
			return name, nil
		} else {
			st = "DELETE FROM access_tokens WHERE token = ?"
			_, err = r.h.Conn().Exec(st, token)
			if err != nil {
				return "", err		
			}
			return "", errors.New("Token expired")
		}
	}
}

func (r *SqlRepo) GetUserByName(name string) (*web.User, error) {
	var pass, email string
	st := "SELECT * FROM users WHERE name = ?"	
	err := r.h.Conn().QueryRow(st, name).Scan(&name, &email, &pass)
	if (err != nil) {
		return &web.User{}, err
	} else {
		return &web.User{Name: name, Pass: pass, Email: email}, nil
	}	
}

func (r *SqlRepo) GetCommentsByItem(itemId int64) ([]*Comment, error) {
	st := "SELECT * FROM comments WHERE item = ? ORDER BY likes DESC"	
	rows, err := r.h.Conn().Query(st, itemId)
	defer rows.Close()
	comments := make([]*Comment, 0)
	for rows.Next() {
    		var id int64
		var likes int
		var date time.Time
    		var text, author string
    		err = rows.Scan(&id, &itemId, &date, &text, &author, &likes)
		if err != nil {
			return nil, err
		}
		comments = append (comments, &Comment{Id: id, Item_id: itemId,
				Text: text, Author: author, Time: date, Likes: likes})
	}
	err = rows.Err() 
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *SqlRepo) GetBestComments() ([]*Item, error) {
	st := "SELECT comments.*, items.imgurl, items.title, items.source from comments " +
		"INNER JOIN items ON comments.item=items.id ORDER BY likes DESC LIMIT 100"	
	rows, err := r.h.Conn().Query(st)
	defer rows.Close()
	items := make([]*Item, 0)
	for rows.Next() {
    		var item_id, comment_id int64
		var likes int
		var date time.Time
    		var text, author, url, title, source string
    		err = rows.Scan(&comment_id, &item_id, &date, &text, &author, &likes, &url, &title, &source)
		if err != nil {
			return nil, err
		}
		item := &Item{Id: item_id, Url: url, Title: title, Source: source}
		comment := &Comment{Id: comment_id, Text: text, Author: author, Likes: likes}
		item.BestComment = comment
		items = append (items, item)
	}
	err = rows.Err() 
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SqlRepo) GetUserByEmail(email string) (*web.User, error) {
	var pass, name string
	st := "SELECT * FROM users WHERE email = ?"	
	err := r.h.Conn().QueryRow(st, email).Scan(&name, &email, &pass)
	if (err != nil) {
		return &web.User{}, err
	} else {
		return &web.User{Name: name, Pass: pass, Email: email}, nil
	}	
}

func (r *SqlRepo) GetItemByUrl(img_url string) (*Item, error) {
	var title, source string
	var id int64
	st := "SELECT * FROM items WHERE imgurl = ?"	
	err := r.h.Conn().QueryRow(st, img_url).Scan(&id, &img_url, &title, &source)
	if err != nil {
		return &Item{}, err
	} else {
		return &Item{Id: id, Title: title, Source: source, Url: img_url}, nil
	}	
}

func (r *SqlRepo) GetItemById(id int64) (*Item, error) {
	var title, img_url, source string
	st := "SELECT * FROM items WHERE id = ?"	
	err := r.h.Conn().QueryRow(st, id).Scan(&id, &img_url, &title, &source)
	if err != nil {
		return &Item{}, err
	} else {
		return &Item{Id: id, Title: title, Source: source, Url: img_url}, nil
	}	
}