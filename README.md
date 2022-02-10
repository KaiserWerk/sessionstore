# sessionstore

[![Go Reference](https://pkg.go.dev/badge/github.com/KaiserWerk/sessionstore.svg)](https://pkg.go.dev/github.com/KaiserWerk/sessionstore/v2)

This is a pure Go session store implementation. A session manager is use to create 
and remove sessions, add, edit and remove values bound to a session. This can be used to contextualize sessions,
e.g. for logically separating user login and admin login or shoping carts or ....

## Prerequisites

* You should have a basic knowledge of how sessions and cookies work.

## Usage

#### SessionManager

As initialization, __create__ a new SessionManager with a supplied session cookie name:

```
var sessMgr *sessionstore.SessionManager = sessionstore.NewManager("MY_SESSION_NAME")
```

#### Session

Then, you can __create__ a Session with a supplied validity, e.g. 30 days: 

```
sess, err := sessMgr.CreateSession(time.Now().Add(30 * 24 * time.Hour))
```

__Retrieve__ a Session by ID (typically obtained from a session cookie):

```
sess, err := sessMgr.GetSession("123abc")
```

__Remove__ a Session by ID (when a user logs out): 

```
err := sessMgr.RemoveSession("123abc")
```

#### Session Variables/Values

__Add__ or __set__ a value to a session (not SessionManager!) and __get__ it:

```
sess.SetVar("key", "value")

val, exists := sess.GetVar("key")
```

#### Cookies
coming soon
```

```

