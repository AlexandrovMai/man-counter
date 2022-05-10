package devices

import (
	"man-counter/storage"
)

type Devices struct {
	s *storage.Storage
}

func New(s *storage.Storage) *Devices {
	return &Devices{
		s,
	}
}

func (v *Devices) OnMessage(key, value []byte) error {

	return v.s.UpateActiveUser(string(key))
}
