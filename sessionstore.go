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
	sessCleanupMutex sync.Mutex
	mgrs []*SessionManager
)


// NewManager creates and returns a new *SessionManager
func NewManager(cn string) *SessionManager {
	sm := &SessionManager{
		CookieName: cn,
		Sessions: make([]Session, 0),
		Messages: make([]Message, 0),
	}

	mgrs = append(mgrs, sm)

	return sm
}

// CreateSession creates a new Session under the *SessionManager
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
	go m.cleanup()
	return s, nil
}

// GetSession retrieves the Session with the supplied session ID
func (m *SessionManager) GetSession(id string) (Session, error) {
	for k, v := range m.Sessions {
		if v.Id == id {
			if !v.Lifetime.After(time.Now()) {
				m.Sessions = removeSessionIndex(m.Sessions, k)
			} else {
				return v, nil
			}
		}
	}
	return Session{}, errors.New("could not find Session for given ID")
}

// RemoveSession removes the Session with the supplied session ID
func (m *SessionManager) RemoveSession(id string) error {
	for i, v := range m.Sessions {
		if v.Id == id {
			sessMgrMutex.Lock()
			m.Sessions = removeSessionIndex(m.Sessions, i)
			sessMgrMutex.Unlock()
			return nil
		}
	}
	return errors.New("could not find Session for the given ID")
}

// RemoveAllSessions removes all Sessions from a *SessionManager
func (m *SessionManager) RemoveAllSessions() {
	 m.Sessions = []Session{}
}

// AddMessage adds a flash message to the *SessionManager's flash bag
func (m *SessionManager) AddMessage(t string, content string) {
	msg := Message{
		MessageType: t,
		Content:     content,
	}
	m.Messages = append(m.Messages, msg)
}

func (m *SessionManager) GetMessages() []Message {
	if len(m.Messages) > 0 {
		tmp := make([]Message, len(m.Messages))
		copy(tmp, m.Messages)
		m.Messages = make([]Message, 0)
		return tmp
	}
	return nil
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

// RemoveCookie removes the session cookie (s
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

// GetCookieValue fetches the session ID from the session cookie of a given request
func (m *SessionManager) GetCookieValue(r *http.Request) (string, error) {
	c, err := r.Cookie(m.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// removeSessionIndex removes a session from a session slice with the given index
func removeSessionIndex(s []Session, index int) []Session {
	return append(s[:index], s[index+1:]...)
}

// generateSessionId generates a new session ID
func generateSessionId(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// cleanup removed all invalid sessions
func (m *SessionManager) cleanup() {
	for k := range m.Sessions {
		if m.Sessions[k].Lifetime.After(time.Now()) { // after?
			sessCleanupMutex.Lock()
			m.Sessions = removeSessionIndex(m.Sessions, k)
			sessCleanupMutex.Unlock()
		}
	}
}