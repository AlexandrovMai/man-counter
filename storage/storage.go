package storage

import (
	"errors"
	"github.com/go-pg/pg/v10"
)

const (
	visitorsCounterId = "visitors-count-id"
)

type Storage struct {
	db *pg.DB
}

func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return errors.New("db object not inited")
}

func (s *Storage) UpateActiveUser(id string) error {
	_, err := s.db.Model(&UserDevices{Id: id}).OnConflict("(id) DO UPDATE").Set("last_seen=now()").Insert()
	return err
}

func (s *Storage) GetActiveUsersCount() (int, error) {
	return s.db.Model((*UserDevices)(nil)).Where("last_seen >= now() - '10 seconds'::interval").Count()
}

func (s *Storage) GetVisitorsCount() (int, error) {
	visitors := &Visitors{Id: visitorsCounterId}
	err := s.db.Model(visitors).WherePK().Select()
	if err != nil {
		return 0, err
	}

	return visitors.Count, err
}

func (s *Storage) SetZeroVisitors() error {
	_, err := s.db.Model(&Visitors{Id: visitorsCounterId}).OnConflict("(id) DO UPDATE").Set("count=0").Insert()
	return err
}

func (s *Storage) AddVisitor() error {
	_, err := s.db.Model(&Visitors{Id: visitorsCounterId}).Set("count=count+1").WherePK().Update()
	return err
}

func (s *Storage) DelVisitor() error {
	_, err := s.db.Model(&Visitors{Id: visitorsCounterId}).Set("count=count-1").WherePK().Update()
	return err
}
