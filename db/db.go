package db

import (
	"database/sql"
	. "github.com/beppeben/m3m3/utils"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"fmt"
	"time"
	"errors"
)

var (
	db 						*sql.DB
	stFindUserByName			*sql.Stmt
	stFindUserByEmail		*sql.Stmt
	stInsertUser				*sql.Stmt
	stInsertTempToken		*sql.Stmt
	stFindTempToken			*sql.Stmt
	stDeleteTempToken		*sql.Stmt
	stInsertAccessToken		*sql.Stmt
	stDeleteAccessToken		*sql.Stmt
	stFindAccessToken		*sql.Stmt
)


func init() {
	Connect_Database()

}

func Connect_Database() {
	var err error
	db, err = sql.Open("mysql", GetUserDB()+":"+GetPassDB()+"@/m3m3")
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	} else {
		log.Println("[DB] Established connection with database")
	}
	
	if (ResetDB()){
		err = dropTables()
		if err != nil {
			return			
		}	
		err = createTables()	
		if err != nil {
			return			
		}
		log.Println("[DB] Tables created successfully")
	}
	
	InitializeStatements()
}

func InitializeStatements() {
	var err error
	stFindUserByName, err = db.Prepare("SELECT * FROM users WHERE name = ?")
    if err != nil {
        panic(err.Error()) 
    }	
	stFindUserByEmail, err = db.Prepare("SELECT * FROM users WHERE email = ?")
    if err != nil {
        panic(err.Error()) 
    }
	stInsertUser, err = db.Prepare("INSERT INTO users VALUES(?,?,?)")
	if err != nil {
        panic(err.Error()) 
    }
	stInsertTempToken, err = db.Prepare("INSERT INTO temp_tokens VALUES(?,?,?,?,?)")
	if err != nil {
        panic(err.Error()) 
    }
	stFindTempToken, err = db.Prepare("SELECT * FROM temp_tokens WHERE token = ?")
    if err != nil {
        panic(err.Error()) 
    }
	stDeleteTempToken, err = db.Prepare("DELETE FROM temp_tokens WHERE token = ?")
    if err != nil {
        panic(err.Error()) 
    }	
	stInsertAccessToken, err = db.Prepare("INSERT INTO access_tokens VALUES(?,?,?)")
    if err != nil {
        panic(err.Error()) 
    }
	stDeleteAccessToken, err = db.Prepare("DELETE FROM access_tokens WHERE token = ?")
    if err != nil {
        panic(err.Error()) 
    }
	stFindAccessToken, err = db.Prepare("SELECT * FROM access_tokens WHERE token = ?")
    if err != nil {
        panic(err.Error()) 
    }
}
/*
func InsertUserToken(user *User, tempToken string, t time.Time) error{
	return Transact(db, func (tx *sql.Tx) error {
        if _, err := tx.Stmt(stInsertUser).Exec(user.Name, user.Email, user.Pass); err != nil {
            return err
        }
        if _, err := tx.Stmt(stInsertTempToken).Exec(tempToken, user.Name, t); err != nil {
            return err
        }
		return nil
    })
}
*/

func InsertTempToken(user *User, tempToken string, t time.Time) error {
	_, err := stInsertTempToken.Exec(tempToken, user.Name, user.Email, user.Pass, t)
	return err
}

func InsertAccessToken (tempToken string, name string, t time.Time) error {
	_, err := stInsertAccessToken.Exec(tempToken, name, t)
	return err
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

func Transact(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
    tx, err := db.Begin()
    if err != nil {
        return
    }
    defer func() {
        if p := recover(); p != nil {
            switch p := p.(type) {
            case error:
                err = p
            default:
                err = fmt.Errorf("%s", p)
            }
        }
        if err != nil {
            tx.Rollback()
            return
        }
        err = tx.Commit()
    }()
    return txFunc(tx)
}

func FindUserNameByToken(token string) (string, error) {
	var name string
	err := stFindAccessToken.QueryRow(token).Scan(&token, &name)
	return name, err
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

func FindUserByEmail(email string) (User, error) {
	var pass, name string
	err := stFindUserByEmail.QueryRow(email).Scan(&name, &email, &pass)
	if (err != nil) {
		return User{}, err
	} else {
		return User{Name: name, Pass: pass, Email: email}, nil
	}	
}

func dropTables() error{
	stmts, err := ReadLines("./config/drop.sql")
	if err != nil {
		log.Printf("[DB] Drop script not found: %v", err)	
		return err	
	} else {
		for i, _ := range stmts {
			execute(stmts[i], false)
		}
	}	
	return nil
}

func createTables() error{
	stmts, err := ReadLines("./config/create.sql")
	if err != nil {
		log.Printf("[DB] Create script not found: %v", err)
		return err		
	} else {
		for i, _ := range stmts {
			execute(stmts[i], true)
		}
	}	
	return nil
}

func execute(statement string, throw bool) error{
	//log.Printf("[DB] Executing: %s", statement)
	st, err := db.Prepare(statement)
    	if err != nil {
        panic(err.Error()) 
    	}
    	defer st.Close()	
	_, err = st.Exec()
    if err != nil {
		log.Printf("[DB] Error: %v, executing statement: %s", err, statement)
		if (throw) {
			panic(err.Error()) 
		} 	
   	}	
	return err
}

