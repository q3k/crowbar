package main

import (
	"os"
	"fmt"
	"bufio"
	"strings"
	"crypto/sha256"
	"crypto/hmac"
)

type localUser struct {
	username	string
	password	string
}

type User interface {
	Authenticate(nonce []byte, hmac []byte) bool
}

func (u localUser) Authenticate(nonce []byte, givenMac []byte) bool {
	mac := hmac.New(sha256.New, []byte(u.password))
	mac.Write(nonce)
	expectedMac := mac.Sum(nil)
	return hmac.Equal(expectedMac, givenMac)
}

var userMap = map[string]User{}

func UserGet(username string) (User, bool) {
	val, ok := userMap[username]
	return val, ok
}

func loadUsersFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open user file: %s\n", err)
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Count(line, ":") != 1 {
			fmt.Fprintf(os.Stderr, "Invalid userfile line: %s\n", line)
			continue
		}
		parts := strings.Split(line, ":")

		user := localUser{username: parts[0], password: parts[1]}
		fmt.Fprintf(os.Stderr, "Loaded user %s\n", user.username)
		userMap[user.username] = user
	}
}
