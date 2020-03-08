package main

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	usersKey   = "users:"
	minNameLen = 2
	maxNameLen = 32
	nameRegexp = `^[a-zA-Z0-9_-]+$`
	minPassLen = 8
	bcryptCost = bcrypt.DefaultCost
)