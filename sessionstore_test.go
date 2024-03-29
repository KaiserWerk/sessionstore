package sessionstore

import (
	"os"
	"testing"
	"time"
)

func Test_NewManager(t *testing.T) {
	mgr := NewManager("test")
	if mgr == nil {
		t.Errorf("new manager was not created, mgr is nil")
	}
}

func Test_CreateSession(t *testing.T) {
	mgr := NewManager("test")
	wantLen := 1

	_, err := mgr.CreateSession(time.Now())
	if err != nil {
		t.Fatalf("could not create session: '%s'", err.Error())
	}

	if len(mgr.Sessions) != wantLen {
		t.Fatalf("expected len %d, got %d", wantLen, len(mgr.Sessions))
	}
}

func Test_generateSessionId(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{"with len 30", 30},
		{"with len 50", 30},
		{"with len 150", 150},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := generateSessionId(make([]*Session, 0), tc.wantLen)
			if err != nil {
				t.Fatalf("could not generate session ID: %s", err.Error())
			}

			if len(id) != tc.wantLen*2 { // encoding to hex doubles the byte count
				t.Fatalf("expected len %d, got %d", tc.wantLen, len(id))
			}
		})
	}
}

func Test_removeSessionIndex(t *testing.T) {
	s := []*Session{&Session{}, &Session{}, &Session{}}
	res := removeSessionIndex(s, 1)
	if len(res) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(res))
	}
}

func Test_Persistence(t *testing.T) {
	sn := "test123"
	mgr := NewManager(sn)
	file := "test.sess"
	defer os.Remove(file)

	err := mgr.ToFile(file)
	if err != nil {
		t.Fatalf("could not persist sessions to file: %s", err.Error())
	}

	mgr2, err := NewManagerFromFile(file)
	if err != nil {
		t.Fatalf("could not get sessions from file: %s", err.Error())
	}

	if mgr2.SessionName != sn {
		t.Fatalf("expected session name '%s', got '%s", sn, mgr2.SessionName)
	}

	if len(mgr2.Sessions) != 0 {
		t.Fatalf("expected 0 sessions in mgr from file, got %d", len(mgr2.Sessions))
	}
}
