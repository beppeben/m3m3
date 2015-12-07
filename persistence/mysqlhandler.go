package persistence

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

type MySqlConfig interface {
	GetUserDB() 	string
	GetPassDB() 	string
	GetDBName() 	string
	ResetDB()	bool
}

type SysUtils interface {
	DeleteImages() error
	GetDropStatements() ([]string, error)
	GetCreateStatements() ([]string, error)
}

type MySqlHandler struct { 
	Connection 	*sql.DB 
	utils		SysUtils
}


func NewMySqlHandler(c MySqlConfig, u SysUtils) *MySqlHandler { 
	conn, err := sql.Open("mysql", c.GetUserDB() + ":" + 
		c.GetPassDB() + "@/" + c.GetDBName() + "?parseTime=true") 
	if err != nil {
		panic(err.Error())
	}
	err = conn.Ping()
	if err != nil {
		panic(err.Error())
	} else {
		log.Infoln("Established connection with database")
	}
	if (c.ResetDB()){
		stmts, err := u.GetDropStatements()
		if err != nil {
			panic(err.Error())
		}
		dropTables(conn, stmts)
		stmts, err = u.GetCreateStatements()
		if err != nil {
			panic(err.Error())
		}
		err = createTables(conn, stmts)	
		if err != nil {
			panic(err.Error())
		}
		err = u.DeleteImages()
		if err != nil {
			panic(err.Error())
		}
		log.Info("Tables created successfully")
	}
	
	return &MySqlHandler{Connection: conn, utils: u}
}

func dropTables(conn *sql.DB, stmts []string) {
	for i, _ := range stmts {
		_, err := conn.Exec(stmts[i])
		if err != nil {
			log.Info(err.Error())
		}				
	}
}

func createTables(conn *sql.DB, stmts []string) error{
	for i, _ := range stmts {
		_, err := conn.Exec(stmts[i])
		if err != nil {
			return err
		}				
	}
	return nil
}


func (handler *MySqlHandler) Conn() *sql.DB { 
	return handler.Connection
}

func (handler *MySqlHandler) Transact(txFunc func(*sql.Tx) (interface{}, error)) (obj interface{}, err error) {    
	tx, err := handler.Connection.Begin()
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