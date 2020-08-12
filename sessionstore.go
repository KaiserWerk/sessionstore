package sessionstore

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

type SessionManager struct {
	Sessions []*Session
	Messages []*Message
}

type Message struct {
	MessageType		string
	Content			string
}

type Session struct {
	Id         string
	CookieName string
	Lifetime   time.Time
	Vars       map[string]string
}

var (
	mut *sync.RWMutex
)

func NewManager() *SessionManager {
	return &SessionManager{}
}

func (m *SessionManager) Create(name string, lt time.Time) (*Session, error) {
	id := uuid.New().String()
	fmt.Println("generated uuid:", id)
	for _, v := range m.Sessions {
		if v.Id == id {
			return nil, errors.New("could not use generated uuid because it is already in use")
		}
	}

	s := &Session{
		Id:         id,
		CookieName: name,
		Lifetime:   lt,
		Vars:       nil,
	}

	m.Sessions = append(m.Sessions, s)

	return s, nil
}

func (m *SessionManager) Get(id string) (*Session, error) {
	for _, v := range m.Sessions {
		if v.Id == id {
			return v, nil
		}
	}

	return nil, errors.New("could not find Session for given ID")
}

func (m *SessionManager) Remove(id string) {
	for i, v := range m.Sessions {
		if v.Id == id {
			mut.Lock()
			m.Sessions = removeIndex(m.Sessions, i)
			mut.Unlock()
			return
		}
	}
}

func (m *SessionManager) AddMessage(t string, msg string) {

}

func (m *SessionManager) GetMessage(t string) {

}

func (s *Session) Get(key string) (string, bool) {
	val, ok := s.Vars[key]
	return val, ok
}

func (s *Session) Set(key string, value string) {
	s.Vars[key] = value
}

func (s *Session) SetCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.CookieName,
		Value:    s.Id,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
	})
}

func (s *Session) GetCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(s.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, err
}

// Helper
func removeIndex(s []*Session, index int) []*Session {
	return append(s[:index], s[index+1:]...)
}
