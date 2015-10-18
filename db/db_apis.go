package db

import (
	"database/sql"
	. "github.com/beppeben/m3m3/utils"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"errors"
)


func InsertLike (username string, comment_id int64) (comment *Comment, err error) {
	obj, err := Transact(db, func (tx *sql.Tx) (interface{}, error) {
		var d time.Time
		var t,a string
		var likes int
		var item_id int64
		err = tx.Stmt(stFindCommentById).QueryRow(comment_id).Scan(&comment_id, &item_id, &d, &t, &a, &likes)
		if err != nil {
			panic("The comment does not exist")
		} 
		comment := &Comment{Id: comment_id, Item_id: item_id, Text: t, Author: a, Likes: likes + 1, Time: d}
		_, err = tx.Stmt(stInsertLike).Exec(username, comment_id)
		if err != nil {
			panic("Comment already liked")
		}
		_, err = tx.Stmt(stUpdateLike).Exec(comment_id)
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


func InsertComment (comment *Comment) error {
	result, err := stInsertComment.Exec(nil, comment.Item_id, comment.Time, comment.Text, comment.Author, comment.Likes)
	if err == nil {
		id, err := result.LastInsertId()
		if err == nil {
			comment.Id = id
		}
	}
	return err
}

func InsertTempToken (user *User, tempToken string, t time.Time) error {
	_, err := stInsertTempToken.Exec(tempToken, user.Name, user.Email, user.Pass, t)
	return err
}

func InsertAccessToken (tempToken string, name string, t time.Time) error {
	_, err := stInsertAccessToken.Exec(tempToken, name, t)
	return err
}

func InsertItem (img_url, title, source string) (int64, error) {
	result, err := stInsertItem.Exec(nil, img_url, title, source)
	if err != nil {
		return -1, err
	} else {
		return result.LastInsertId()
	}
}

func DeleteAccessToken (token string) error {
	_, err := stDeleteAccessToken.Exec(token)
	return err
}

func InsertUserFromTempToken (tempToken string) (*User, error) {
	var name, pass, email string
	var t time.Time
	var user *User
	err := stFindTempToken.QueryRow(tempToken).Scan(&tempToken, &name, &email, &pass, &t)	
	if (err != nil) {
		return user, err
	}
	_, err = stDeleteTempToken.Exec(tempToken)
	if (err != nil) {
		return &User{}, err
	}
	if time.Now().After(t) {
		return &User{}, errors.New("expired token")
	}
	_, err = stInsertUser.Exec(name, email, pass)
	if (err != nil) {
		return &User{}, err
	}
	user = &User{Name: name, Email: email, Pass: pass}
	return user, err
}



func FindUserNameByToken(token string) (string, error) {
	var name string
	var t time.Time
	err := stFindAccessToken.QueryRow(token).Scan(&token, &name, &t)
	if err != nil {
		return name, err		
	} else {
		if time.Now().Before(t) {
			return name, nil
		} else {
			stDeleteAccessToken.Exec(token)
			return "", errors.New("Token expired")
		}
	}
}

func FindUserByName(name string) (*User, error) {
	var pass, email string
	err := stFindUserByName.QueryRow(name).Scan(&name, &email, &pass)
	if (err != nil) {
		return &User{}, err
	} else {
		return &User{Name: name, Pass: pass, Email: email}, nil
	}	
}

func FindCommentsByItem(itemId int64) ([]*Comment, error) {
	rows, err := stFindCommentsByItem.Query(itemId)
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


func FindBestComments() ([]*Item, error) {
	rows, err := stFindBestComments.Query()
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


func FindUserByEmail(email string) (*User, error) {
	var pass, name string
	err := stFindUserByEmail.QueryRow(email).Scan(&name, &email, &pass)
	if (err != nil) {
		return &User{}, err
	} else {
		return &User{Name: name, Pass: pass, Email: email}, nil
	}	
}

func FindItemByUrl(img_url string) (*Item, error) {
	var title, source string
	var id int64
	err := stFindItemByUrl.QueryRow(img_url).Scan(&id, &img_url, &title, &source)
	if err != nil {
		return &Item{}, err
	} else {
		return &Item{Id: id, Title: title, Source: source, Url: img_url}, nil
	}	
}

func FindItemById(id int64) (*Item, error) {
	var title, img_url, source string
	err := stFindItemById.QueryRow(id).Scan(&id, &img_url, &title, &source)
	if err != nil {
		return &Item{}, err
	} else {
		return &Item{Id: id, Title: title, Source: source, Url: img_url}, nil
	}	
}



