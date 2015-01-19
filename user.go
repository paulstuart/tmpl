package main

import (
	"encoding/json"
	"fmt"
)

type FullUser struct {
	ID       int64  `sql:"id" key:"true" table:"users"`
	Username string `sql:"username"`
	First    string `sql:"firstname"`
	Last     string `sql:"lastname"`
	Email    string `sql:"email"`
	Password string `sql:"password"`
	Salt     string `sql:"salt"`
	Role     int    `sql:"role"`
}

type User struct {
	ID       int64  `sql:"id" key:"true" table:"users"`
	Username string `sql:"username"`
	First    string `sql:"firstname"`
	Last     string `sql:"lastname"`
	Email    string `sql:"email"`
	Role     int    `sql:"role"`
}

func userList() []User {
	return nil
}

func (u *FullUser) User() *User {
	if u == nil {
		return nil
	}
	return &User{ID: u.ID, Username: u.Username, Email: u.Email}
}

/*
//TODO: use iota
type UserLevel int

// go \:generate stringer -type=UserLevel -output=userlevels.go
const (
	Normal UserLevel = iota
	Editor
	Admin
)
*/

/*
type userLevel struct {
	ID   int
	Name string
}

var userLevels = []userLevel{{0, "User"}, {1, "Editor"}, {2, "Admin"}}
*/

func (u *FullUser) Load(where string, args ...interface{}) error {
	if u == nil {
		return fmt.Errorf("nil object\n")
	}
	return db.ObjectLoad(u, where, args...)
}

func (u *FullUser) ByEmail(email string) error {
	return u.Load("where email=?", email)
}

func (u *FullUser) Add() error {
	var err error
	u.ID, err = db.ObjectInsert(*u)
	return err
}

func getUser(where string, args ...interface{}) (*User, error) {
	u := &User{}
	err := db.ObjectLoad(u, where, args...)
	return u, err
}

func UserByID(id interface{}) (*User, error) {
	return getUser("where id=?", id)
}

func UserByLogin(login string) (*User, error) {
	return getUser("where login=?", login)
}

func UserByEmail(email string) (*User, error) {
	return getUser("where email=?", email)
}

func (u *User) Add() error {
	var err error
	u.ID, err = db.ObjectInsert(*u)
	return err
}

func (u *User) Update() error {
	return nil
}

func (u *User) Delete() error {
	return nil
}

func (u *FullUser) SetPassword(password string) {
	u.Salt = RandomString(16)
	u.Password = SHA1(password + u.Salt)
}

func (u FullUser) Authenticate(password string) bool {
	return u.Password == SHA1(password+u.Salt)
}

func Authenticate(email, password string) *User {
	user := &FullUser{}
	if err := user.ByEmail(email); err == nil {
		if user.Authenticate(password) {
			return user.User()
		}
	}
	return nil
}

// find user by email, return as encrypted json string
func (u User) ToCookie() string {
	text, e1 := json.Marshal(u)
	if e1 != nil {
		fmt.Println("Marshal user", u, "Error", e1)
		return ""
	}
	secret, e2 := StringEncrypt(string(text))
	if e2 != nil {
		fmt.Println("Encrypt text", text, "Error", e2)
		return ""
	}
	return secret
}

func userCookie(email string) string {
	user, err := UserByEmail(email)
	if err != nil {
		fmt.Println("Can't find user", user, "Error", err)
		return ""
	}
	return user.ToCookie()
}

// unencrypt json string
func UserFromCookie(cookie string) *User {
	if len(cookie) > 0 {
		if plain, err := StringDecrypt(cookie); err == nil {
			var user User
			if err = json.Unmarshal([]byte(plain), &user); err == nil {
				return &user
			}
		}
	}
	return nil
}

// unencrypt json string
/*
func (user *User) FromCookie(cookie string) {
	if len(cookie) > 0 {
		if plain, err := StringDecrypt(cookie); err != nil {
			fmt.Println("Decrypt text", cookie, "Error", err)
			return
		} else {
			if err = json.Unmarshal([]byte(plain), user); err != nil {
				fmt.Println("Unmarshal text", plain, "Error", err)
			}
		}
	}
}

// unencrypt json string
func userFromCookie(cookie string) *User {
	user := &User{}
	user.FromCookie(cookie)
	return user
}
*/
