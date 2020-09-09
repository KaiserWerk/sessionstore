package sessionstore

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SessionManager struct {
	CookieName		string
	Sessions 		[]Session
	Messages 		[]Message
}

type Message struct {
	MessageType		string
	Content			string
}

type Session struct {
	Id          string
	Lifetime    time.Time

	Vars        map[string]string
}

var (
	sessMgrMutex sync.Mutex
	sessMutex sync.Mutex
)

func NewManager(cn string) *SessionManager {
	return &SessionManager{
		CookieName: cn,
		Sessions: make([]Session, 0),
		Messages: make([]Message, 0),
	}
}

func (m *SessionManager) CreateSession(lt time.Time) (Session, error) {
	id := generateSessionId(30)

	for _, v := range m.Sessions {
		if v.Id == id {
			return Session{}, errors.New("could not use generated session id because it is already in use")
		}
	}

	s := Session{
		Id:         id,
		Lifetime:   lt,
		Vars:       make(map[string]string),
	}

	sessMgrMutex.Lock()
	defer sessMgrMutex.Unlock()
	m.Sessions = append(m.Sessions, s)

	return s, nil
}

func (m *SessionManager) GetSession(id string) (Session, error) {
	for _, v := range m.Sessions {
		if v.Id == id {
			return v, nil
		}
	}

	return Session{}, errors.New("could not find Session for given ID")
}

func (m *SessionManager) RemoveSession(id string) error {
	for i, v := range m.Sessions {
		if v.Id == id {
			sessMgrMutex.Lock()
			m.Sessions = removeIndex(m.Sessions, i)
			sessMgrMutex.Unlock()

			return nil
		}
	}

	return nil
}

func (m *SessionManager) RemoveAllSessions() {
	 m.Sessions = []Session{}
}

func (m *SessionManager) AddMessage(t string, content string) {
	msg := Message{
		MessageType: t,
		Content:     content,
	}
	m.Messages = append(m.Messages, msg)
}

func (m *SessionManager) GetMessages() []Message {
	l := m.Messages
	// reset!
	m.Messages = make([]Message, 0)
	fmt.Println("returning", len(l), "messages")
	return l
}

func (s Session) GetVar(key string) (string, bool) {
	val, ok := s.Vars[key]
	return val, ok
}

func (s Session) SetVar(key string, value string) {
	sessMutex.Lock()
	defer sessMutex.Unlock()
	s.Vars[key] = value
}

func (m *SessionManager) SetCookie(w http.ResponseWriter, value string) error {
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
	})

	return nil
}

func (m *SessionManager) RemoveCookie(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -10,
		HttpOnly: true,
	})

	return nil
}

func (m *SessionManager) GetCookieValue(r *http.Request) (string, error) {
	c, err := r.Cookie(m.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// Helper

func removeIndex(s []Session, index int) []Session {
	return append(s[:index], s[index+1:]...)
}

func generateSessionId(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}