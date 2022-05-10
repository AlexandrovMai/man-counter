package storage

import "time"

type Visitors struct {
	tableName struct{} `pg:"visitors"`

	Id    string `pg:",unique,notnull"`
	Count int    `pg:",unique,notnull,default:0"`
}

type UserDevices struct {
	Id       string    `pg:",unique,notnull"`
	LastSeen time.Time `pg:"last_seen,default:now()"`
}
