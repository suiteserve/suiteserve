package auth

import (
	"encoding/json"
	"errors"
	"github.com/tidwall/buntdb"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

const (
	usersKey = "users:"
)

type Role string

const (
	ViewerRole Role = "viewer"
	AdminRole       = "admin"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	Name string `json:"name,omitempty"`
	Pass []byte `json:"pass,omitempty"`
	Role Role   `json:"role"`
}

func CreateUser(db *buntdb.DB, name, pass string, role Role) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		Pass: hash,
		Role: role,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return User{}, err
	}

	if err := db.Update(func(tx *buntdb.Tx) error {
		key := usersKey + name
		if _, err := tx.Get(key); err == nil {
			return ErrUserExists
		} else if err != buntdb.ErrNotFound {
			return err
		}

		_, _, err := tx.Set(key, string(userJson), nil)
		return err
	}); err != nil {
		return User{}, err
	}
	return user, err
}

func FindAllUsers(db *buntdb.DB) ([]User, error) {
	userJsons := map[string]string{}
	if err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(usersKey+"*", func(key, userJson string) bool {
			name := strings.SplitN(key, ":", 2)[1]
			userJsons[name] = userJson
			return true
		})
	}); err != nil {
		return nil, err
	}

	return userJsonsToUsers(userJsons)
}

func FindUserByName(db *buntdb.DB, name string) (User, error) {
	user, err := getUserByName(db, name)
	if err != nil {
		return User{}, err
	}

	user.Pass = nil
	return user, nil
}

func FindUsersByRole(db *buntdb.DB, role Role) ([]User, error) {
	err := db.CreateIndex("users.role", usersKey+"*", buntdb.IndexJSON("role"))
	if err != nil && err != buntdb.ErrIndexExists {
		return nil, err
	}

	userJsons := map[string]string{}
	if err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual("users.role", `{"role":"`+string(role)+`"}`,
			func(key, userJson string) bool {
				name := strings.SplitN(key, ":", 2)[1]
				userJsons[name] = userJson
				return true
			})
	}); err != nil {
		return nil, err
	}

	return userJsonsToUsers(userJsons)
}

func CheckUserPass(db *buntdb.DB, name, pass string) (bool, error) {
	user, err := getUserByName(db, name)
	if err != nil {
		return false, err
	}

	if err := bcrypt.CompareHashAndPassword(user.Pass, []byte(pass)); err == nil {
		return true, nil
	} else if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	} else {
		return false, err
	}
}

func DeleteUser(db *buntdb.DB, name string) error {
	return db.Update(func(tx *buntdb.Tx) error {
		if _, err := tx.Delete(usersKey + name); err == buntdb.ErrNotFound {
			return ErrUserNotFound
		} else {
			return err
		}
	})
}

func getUserByName(db *buntdb.DB, name string) (User, error) {
	var userJson string
	if err := db.View(func(tx *buntdb.Tx) error {
		var err error
		if userJson, err = tx.Get(usersKey + name); err == buntdb.ErrNotFound {
			return ErrUserNotFound
		} else {
			return err
		}
	}); err != nil {
		return User{}, err
	}

	var user User
	if err := json.Unmarshal([]byte(userJson), &user); err != nil {
		return User{}, err
	}
	user.Name = name
	return user, nil
}

func userJsonsToUsers(userJsons map[string]string) ([]User, error) {
	var users []User
	for name, userJson := range userJsons {
		var user User
		if err := json.Unmarshal([]byte(userJson), &user); err != nil {
			return nil, err
		}
		user.Name = name
		user.Pass = nil
		users = append(users, user)
	}
	return users, nil
}
