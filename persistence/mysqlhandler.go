package persistence

import (
	. "github.com/beppeben/m3m3/utils"
	"database/sql"
	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

type MySqlHandler struct { 
	Connection *sql.DB 
}


func NewMySqlHandler() *MySqlHandler { 
	conn, err := sql.Open("mysql", GetUserDB() + ":" + 
		GetPassDB() + "@/" + GetDBName() + "?parseTime=true") 
	if err != nil {
		panic(err.Error())
	}
	err = conn.Ping()
	if err != nil {
		panic(err.Error())
	} else {
		log.Infoln("Established connection with database")
	}
	if (ResetDB()){
		err = dropTables(conn)
		if err != nil {
			panic(err.Error())
		}	
		err = createTables(conn)	
		if err != nil {
			panic(err.Error())
		}
		err = DeleteAllImages()
		if err != nil {
			panic(err.Error())
		}
		log.Infoln("Tables created successfully")
	}
	
	return &MySqlHandler{Connection: conn}
}

func dropTables(conn *sql.DB) error{
	stmts, err := ReadLines("./config/drop.sql")
	if err != nil {
		return err
	} else {
		for i, _ := range stmts {
			_, err = conn.Exec(stmts[i])
			if err != nil {
				log.Info(err.Error())
			}				
		}
	}	
	return nil
}

func createTables(conn *sql.DB) error{
	stmts, err := ReadLines("./config/create.sql")
	if err != nil {
		return err
	} else {
		for i, _ := range stmts {
			_, err = conn.Exec(stmts[i])
			if err != nil {
				return err
			}				
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