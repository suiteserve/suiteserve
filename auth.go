package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
)

const (
	usersKey   = "users:"
	minNameLen = 2
	maxNameLen = 32
	nameRegexp = `^[a-zA-Z0-9_-]+$`
	minPassLen = 8
	bcryptCost = bcrypt.DefaultCost
)

type role string

const (
	viewerRole role = "viewer"
	adminRole       = "admin"
)

type user struct {
	Name string `json:"name,omitempty"`
	Pass []byte `json:"pass,omitempty"`
	Role role   `json:"role"`
}

func (s srv) findUser(name string) (user, error) {
	var userJson string
	err := s.db.View(func(tx *buntdb.Tx) error {
		var err error
		userJson, err = tx.Get(usersKey + name)
		return err
	})
	if err == buntdb.ErrNotFound {
		return user{}, httpError{
			error:  fmt.Errorf("user '%s' not found", name),
			Status: http.StatusNotFound,
		}
	}

	var u user
	if err := json.Unmarshal([]byte(userJson), &u); err != nil {
		return user{}, err
	}
	u.Name = name
	u.Pass = nil
	return u, nil
}

func (s srv) createUser(name, pass string, role role) error {
	// Validate name and password.
	if len(name) < minNameLen || len(name) > maxNameLen {
		return httpError{
			error: fmt.Errorf("user name '%s' must be between %d and %d characters",
				name, minNameLen, maxNameLen),
			Status: http.StatusBadRequest,
		}
	}
	if valid, err := regexp.MatchString(nameRegexp, name); err != nil {
		return err
	} else if !valid {
		return httpError{
			error: fmt.Errorf("user name '%s' must match regex %s",
				name, nameRegexp),
			Status: http.StatusBadRequest,
		}
	}
	if len(pass) < minPassLen {
		return httpError{
			error: fmt.Errorf("user pass must be at least %d characters",
				minPassLen),
			Status: http.StatusBadRequest,
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcryptCost)
	if err != nil {
		return err
	}

	user := user{
		Pass: hash,
		Role: role,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *buntdb.Tx) error {
		key := usersKey + name
		if _, err := tx.Get(key); err == nil {
			return httpError{
				error:  fmt.Errorf("user '%s' already exists", name),
				Status: http.StatusConflict,
			}
		} else if err != buntdb.ErrNotFound {
			return err
		}

		_, _, err = tx.Set(key, string(userJson), nil)
		return err
	})
}

func (s srv) deleteUser(name string) error {
	return s.db.Update(func(tx *buntdb.Tx) error {
		if _, err := tx.Delete(usersKey + name); err == buntdb.ErrNotFound {
			return httpError{
				error: fmt.Errorf("user '%s' not found", name),
				Status: http.StatusNotFound,
			}
		} else {
			return err
		}
	})
}