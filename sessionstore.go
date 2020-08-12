package sessionstore

import(
	"errors"
	"fmt"
	"github.com/google/uuid"
)

var (
	sessions []*Session
)

func Create() (*Session, error) {
	id := uuid.New().String()
	fmt.Println("generated uuid:", id)
	for _, v := range sessions {
		if v.Id == id {
			return nil, errors.New("could not use generated uuid because it is already in use")
		}
	}

	s := &Session{
		Id:   id,
		Vars: nil,
	}

	sessions = append(sessions, s)

	return s, nil
}

func (s *Session) Get(key string) (string, bool) {
	val, ok := s.Vars[key]
	return val, ok
}

func (s *Session) Set(key string, value string) {
	s.Vars[key] = value
}