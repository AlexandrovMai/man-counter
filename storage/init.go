package storage

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"log"
)

func InitStorage(dbName, user, pass, url string) (*Storage, error) {
	st := &Storage{
		db: pg.Connect(&pg.Options{
			User:      user,
			Database:  dbName,
			Password:  pass,
			Addr:      url,
			OnConnect: onConnected,
		}),
	}

	_, err := st.db.Exec("SELECT 1;")
	if err != nil {
		log.Println("Error: cannot connect to DB: ", err.Error())
		return nil, err
	}

	// create 0 users field
	visitors := &Visitors{
		Id:    visitorsCounterId,
		Count: 0,
	}
	_, err = st.db.Model(visitors).OnConflict("DO NOTHING").Insert()
	if err != nil {
		log.Println("Error: cannot create visitors row: ", err.Error())
		return nil, err
	}

	return st, nil
}

func onConnected(ctx context.Context, conn *pg.Conn) (err error) {
	err = conn.Model((*UserDevices)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err != nil {
		log.Println("Error: creating user devices model: ", err.Error())
		return
	}
	err = conn.Model((*Visitors)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err != nil {
		log.Println("Error: creating visitors model model: ", err.Error())
	}
	return
}
