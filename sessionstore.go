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
	CookieName		string
	Sessions 		[]*Session
	Messages 		[]*Message
}

type Message struct {
	MessageType		string
	Content			string
}

type Session struct {
	Id         string
	Lifetime   time.Time
	Vars       map[string]interface{}
}

var (
	mut *sync.RWMutex
)

func NewManager(cn string) *SessionManager {
	return &SessionManager{
		CookieName: cn,
		Sessions: make([]*Session, 10),
		Messages: make([]*Message, 10),
	}
}

func (m *SessionManager) CreateSession(name string, lt time.Time) (*Session, error) {
	id := uuid.New().String()
	fmt.Println("generated uuid:", id)
	for _, v := range m.Sessions {
		if v.Id == id {
			return nil, errors.New("could not use generated uuid because it is already in use")
		}
	}

	s := &Session{
		Id:         id,
		Lifetime:   lt,
		Vars:       nil,
	}

	m.Sessions = append(m.Sessions, s)

	return s, nil
}

func (m *SessionManager) GetSession(id string) (*Session, error) {
	for _, v := range m.Sessions {
		if v.Id == id {
			return v, nil
		}
	}

	return nil, errors.New("could not find Session for given ID")
}

func (m *SessionManager) RemoveSession(id string) {
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

func (s *Session) GetVar(key string) (interface{}, bool) {
	val, ok := s.Vars[key]
	return val, ok
}

func (s *Session) SetVar(key string, value string) {
	s.Vars[key] = value
}

func (m *SessionManager) SetCookie(w http.ResponseWriter, value string) {
	fmt.Printf("Setting cookie with name %s, value %s", m.CookieName, value)
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
	})
}

func (m *SessionManager) GetCookieValue(r *http.Request) (string, error) {
	c, err := r.Cookie(m.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// Helper
func removeIndex(s []*Session, index int) []*Session {
	return append(s[:index], s[index+1:]...)
}
