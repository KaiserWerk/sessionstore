package sessionstore

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type SessionManager struct {
	SessionName string
	Sessions    []*Session
	mut         sync.RWMutex
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
func NewManager(sn string) *SessionManager {
	return &SessionManager{
		SessionName: sn,
		Sessions:    make([]*Session, 0),
	}
}

func NewManagerFromFile(file string) (*SessionManager, error) {
	m := SessionManager{}
	fh, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	if err = gob.NewDecoder(fh).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *SessionManager) ToFile(file string) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fh.Close()

	if err = gob.NewEncoder(fh).Encode(m); err != nil {
		return err
	}

	return nil
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

// GetSessionFromUrl is a convenience method to find an existing session by taking the session ID
// from a *url.URL query parameter.
func (m *SessionManager) GetSessionFromUrl(u *url.URL) (*Session, error) {
	return m.GetSession(u.Query().Get(m.SessionName))
}

// GetSessionFromCookie is a convenience method to find an existing session by taking the session ID
// from the cookie with the name initially set when creating the *SessionManager.
func (m *SessionManager) GetSessionFromCookie(r *http.Request) (*Session, error) {
	c, err := r.Cookie(m.SessionName)
	if err != nil {
		return nil, fmt.Errorf("could not read session cookie: %s", err.Error())
	}
	return m.GetSession(c.Value)
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

// GetMessage returns a previously set message
func (s *Session) GetMessage() Message {
	return s.Message
}

// GetVar returns whether the variable with the given name and the actual value, if it exists
func (s *Session) GetVar(key string) (string, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.Vars[key]
	return val, ok
}

// SetVar sets a attaches a variable with the given name and value
func (s *Session) SetVar(key string, value string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.Vars[key] = value
}

// SetCookie is a convenience method to set a session cookie with the initially chosen name.
func (m *SessionManager) SetCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.SessionName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
	})
}

// RemoveCookie is a convenience method to remove the session cookie (s
func (m *SessionManager) RemoveCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.SessionName,
		Value:    "",
		Path:     "/",
		MaxAge:   -10,
		HttpOnly: true,
	})
}

// GetCookieValue fetches the session ID from the session cookie of a given request
func (m *SessionManager) GetCookieValue(r *http.Request) (string, error) {
	c, err := r.Cookie(m.SessionName)
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
