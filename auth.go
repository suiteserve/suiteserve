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
	Hash []byte `json:"hash"`
	Role role   `json:"role"`
}

func (s srv) getUser(name string) (user, error) {
	var user user
	userJson, err := s.getUserJson(name)
	if err != nil {
		return user, err
	}
	err = json.Unmarshal([]byte(userJson), &user)
	return user, err
}

func (s srv) getUserJson(name string) (string, error) {
	var userJson string
	err := s.db.View(func(tx *buntdb.Tx) error {
		var err error
		userJson, err = tx.Get(usersKey + name)
		return err
	})
	if err == buntdb.ErrNotFound {
		return "", StatusError{
			fmt.Errorf("user '%s' not found", name),
			http.StatusNotFound,
		}
	}
	return userJson, err
}

func (s srv) createUser(name, pass string, role role) error {
	// Validate name and password.
	if len(name) < minNameLen || len(name) > maxNameLen {
		return StatusError{
			fmt.Errorf("user name '%s' must be between %d and %d characters",
				name, minNameLen, maxNameLen),
			http.StatusBadRequest,
		}
	}
	if valid, err := regexp.MatchString(nameRegexp, name); err != nil {
		return err
	} else if !valid {
		return StatusError{
			fmt.Errorf("user name '%s' must match regex %s",
				name, nameRegexp),
			http.StatusBadRequest,
		}
	}
	if len(pass) < minPassLen {
		return StatusError{
			fmt.Errorf("user pass must be at least %d characters",
				minPassLen),
			http.StatusBadRequest,
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcryptCost)
	if err != nil {
		return err
	}

	user := user{
		Hash: hash,
		Role: role,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *buntdb.Tx) error {
		key := usersKey + name
		if _, err := tx.Get(key); err == nil {
			return StatusError{
				fmt.Errorf("user '%s' already exists", name),
				http.StatusBadRequest,
			}
		} else if err != buntdb.ErrNotFound {
			return err
		}
		_, _, err = tx.Set(key, string(userJson), nil)
		return err
	})
}
