package sessionstore

import "time"

type Session struct {
	Id			string
	Lifetime	time.Time
	Vars		map[string]string
}