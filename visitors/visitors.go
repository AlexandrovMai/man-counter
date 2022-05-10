package visitors

import (
	"encoding/json"
	"log"
	"man-counter/storage"
)

type Visitors struct {
	s *storage.Storage
}

func New(s *storage.Storage) *Visitors {
	return &Visitors{
		s,
	}
}

type Payload struct {
	Value int    `json:"v"`
	Name  string `json:"n"`
}

func (v *Visitors) OnMessage(key, value []byte) error {
	var vp []Payload
	err := json.Unmarshal(value, &vp)
	if err != nil {
		log.Println("Visitors. OnMessage Error", err.Error(), "p", string(value))
		return nil
	}

	for i := 0; i < len(vp); i++ {
		if vp[i].Name != "vis" {
			log.Println("Visitors. OnMessage Error - bad payload... continue")
			continue
		}
		if vp[i].Value < 0 {
			return v.s.DelVisitor()
		} else {
			return v.s.AddVisitor()
		}
	}

	return nil
}
