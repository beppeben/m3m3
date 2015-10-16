package db

import (
	"database/sql"
	. "github.com/beppeben/m3m3/utils"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"fmt"
	
)

var (
	db 						*sql.DB
	stFindUserByName			*sql.Stmt
	stFindUserByEmail		*sql.Stmt
	stFindCommentsByItem		*sql.Stmt
	stInsertUser				*sql.Stmt
	stInsertTempToken		*sql.Stmt
	stInsertLike				*sql.Stmt
	stUpdateLike				*sql.Stmt
	stFindTempToken			*sql.Stmt
	stDeleteTempToken		*sql.Stmt
	stInsertAccessToken		*sql.Stmt
	stDeleteAccessToken		*sql.Stmt
	stFindAccessToken		*sql.Stmt
	stFindItemByUrl			*sql.Stmt
	stFindItemById			*sql.Stmt
	stFindCommentById		*sql.Stmt
	stInsertItem				*sql.Stmt
	stInsertComment			*sql.Stmt
)

func init() {
	Connect_Database()

}

func Connect_Database() {
	var err error
	db, err = sql.Open("mysql", GetUserDB()+":"+GetPassDB()+"@/m3m3?parseTime=true")
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
		err = DeleteAllImages()
		if err != nil {
			log.Printf("[DB] Problem deleting images: %v", err)			
		}
		log.Println("[DB] Tables created successfully")
	}
	
	InitializeStatements()
}

func InitializeStatements() {
	var err error
	
	stFindUserByName, err = db.Prepare("SELECT * FROM users WHERE name = ?")
    PanicOnErr(err)	
	
	stFindUserByEmail, err = db.Prepare("SELECT * FROM users WHERE email = ?")
    PanicOnErr(err)
	
	stFindCommentsByItem, err = db.Prepare("SELECT * FROM comments WHERE item = ? ORDER BY likes DESC")
    PanicOnErr(err)
	
	stInsertUser, err = db.Prepare("INSERT INTO users VALUES(?,?,?)")
	PanicOnErr(err)
	
	stInsertLike, err = db.Prepare("INSERT INTO likes VALUES(?,?)")
	PanicOnErr(err)
	
	stUpdateLike, err = db.Prepare("UPDATE comments SET likes = likes + 1 WHERE id = ?")
	PanicOnErr(err)
	
	stInsertTempToken, err = db.Prepare("INSERT INTO temp_tokens VALUES(?,?,?,?,?)")
	PanicOnErr(err)
	
	stFindTempToken, err = db.Prepare("SELECT * FROM temp_tokens WHERE token = ?")
    PanicOnErr(err)
	
	stDeleteTempToken, err = db.Prepare("DELETE FROM temp_tokens WHERE token = ?")
    PanicOnErr(err)
		
	stInsertAccessToken, err = db.Prepare("INSERT INTO access_tokens VALUES(?,?,?)")
    PanicOnErr(err)
	
	stDeleteAccessToken, err = db.Prepare("DELETE FROM access_tokens WHERE token = ?")
    PanicOnErr(err)
	
	stFindAccessToken, err = db.Prepare("SELECT * FROM access_tokens WHERE token = ?")
    PanicOnErr(err)
	
	stFindItemByUrl, err = db.Prepare("SELECT * FROM items WHERE imgurl = ?")
    PanicOnErr(err)
	
	stFindItemById, err = db.Prepare("SELECT * FROM items WHERE id = ?")
    PanicOnErr(err)
	
	stFindCommentById, err = db.Prepare("SELECT * FROM comments WHERE id = ?")
    PanicOnErr(err)
		
	stInsertItem, err = db.Prepare("INSERT INTO items VALUES(?,?,?,?)")
	PanicOnErr(err)
	
	stInsertComment, err = db.Prepare("INSERT INTO comments VALUES(?,?,?,?,?,?)")
	PanicOnErr(err)
}

func Transact(db *sql.DB, txFunc func(*sql.Tx) (interface{}, error)) (obj interface{}, err error) {
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