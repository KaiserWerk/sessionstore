package sessionstore

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"
)

type SessionManager struct {
	CookieName string
	Sessions   []*Session
	mut        sync.RWMutex
}

type Message struct {
	MessageType string
	Content     string
}

type Session struct {
	Id       string
	Lifetime time.Time
	Vars     map[string]string
	Message  Message
	mut      sync.RWMutex
}

// NewManager creates and returns a new *SessionManager
func NewManager(cn string) *SessionManager {
	return &SessionManager{
		CookieName: cn,
		Sessions:   make([]*Session, 0),
	}
}

// CreateSession creates a new Session under the *SessionManager
func (m *SessionManager) CreateSession(lt time.Time) (*Session, error) {
	id, err := generateSessionId(m.Sessions, 30)
	if err != nil {
		return nil, err
	}

	s := Session{
		Id:       id,
		Lifetime: lt,
		Vars:     make(map[string]string),
	}
	m.Sessions = append(m.Sessions, &s)

	return &s, nil
}

// GetSession retrieves the Session with the supplied session ID
func (m *SessionManager) GetSession(id string) (*Session, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	for k, v := range m.Sessions {
		if v.Id == id {
			if !v.Lifetime.After(time.Now()) {
				m.Sessions = removeSessionIndex(m.Sessions, k)
			} else {
				return v, nil
			}
		}
	}
	return nil, errors.New("could not find Session for given ID")
}

// RemoveSession removes the Session with the supplied session ID
func (m *SessionManager) RemoveSession(id string) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	for i, v := range m.Sessions {
		if v.Id == id {
			m.Sessions = removeSessionIndex(m.Sessions, i)
			return nil
		}
	}
	return errors.New("could not find Session for the given ID")
}

// RemoveAllSessions removes all Sessions from a *SessionManager
func (m *SessionManager) RemoveAllSessions() {
	m.Sessions = []*Session{}
}

// SetMessage sets a flash message to the *Session
func (s *Session) SetMessage(t string, content string) {
	s.Message = Message{
		MessageType: t,
		Content:     content,
	}
}

func (s *Session) GetMessage() Message {
	return s.Message
}

func (s *Session) GetVar(key string) (string, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.Vars[key]
	return val, ok
}

func (s *Session) SetVar(key string, value string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.Vars[key] = value
}

// SetCookie is a convenience method to set a session cookie with the initially chosen name.
func (m *SessionManager) SetCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
	})
}

// RemoveCookie is a convenience method to remove the session cookie (s
func (m *SessionManager) RemoveCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -10,
		HttpOnly: true,
	})
}

// GetCookieValue fetches the session ID from the session cookie of a given request
func (m *SessionManager) GetCookieValue(r *http.Request) (string, error) {
	c, err := r.Cookie(m.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// removeSessionIndex removes a session from a session slice with the given index
func removeSessionIndex(s []*Session, index int) []*Session {
	return append(s[:index], s[index+1:]...)
}

// generateSessionId generates a new session ID
func generateSessionId(ss []*Session, length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	for _, v := range ss {
		if v.Id == id {
			return generateSessionId(ss, length)
		}
	}

	return id, nil
}
