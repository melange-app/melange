package models

import (
	"bytes"
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/robfig/revel"
	"io"
)

var (
	UserAuthenticationError = errors.New("Unable to Authenticate User from Username/Password")
)

type User struct {
	UserId         int
	Username       string
	Password       string `db:"-"`
	Name           string
	Avatar         string
	Description    string
	HashedPassword []byte
	Transient      bool `db:"-"`
}

func (u *User) GetAvatar() string {
	if u.Avatar == "" {
		h := md5.New()
		io.WriteString(h, fmt.Sprintf("%v@%v", u.Username, "airdispat.ch"))
		return fmt.Sprintf("http://www.gravatar.com/avatar/%x?d=identicon&size=500", h.Sum(nil))
	}
	return u.Avatar
}

func (u *User) String() string {
	return fmt.Sprintf("User(%s)", u.Username)
}

func (user *User) Validate(v *revel.Validation) {
	v.Check(user.Username,
		revel.Required{},
		revel.MaxSize{20},
	)

	v.Check(user.Name,
		revel.Required{},
		revel.MaxSize{100},
	)
}

func AuthenticateUser(username string, password string, dbm *gorp.Transaction) (*User, error) {
	var authenticatedUser []*User

	_, err := dbm.Select(&authenticatedUser,
		"select * from dispatch_user where username = $1", username)

	if err != nil {
		panic(err)
	}
	if len(authenticatedUser) != 1 {
		return nil, UserAuthenticationError
	}

	err = bcrypt.CompareHashAndPassword(authenticatedUser[0].HashedPassword, []byte(password))
	if err != nil {
		return nil, UserAuthenticationError
	}
	return authenticatedUser[0], nil
}

func CreateUser(username string, password string, name string) *User {
	newUser := &User{
		Name:      name,
		Username:  username,
		Password:  password,
		Transient: true,
	}
	newUser.UpdatePassword(password)
	return newUser
}

func (u *User) VerifyPassword(password string) bool {
	bcryptPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return bytes.Equal(bcryptPassword, u.HashedPassword)
}

func (u *User) UpdatePassword(password string) {
	bcryptPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.HashedPassword = bcryptPassword
}

func (u *User) Save(txn *gorp.Transaction) error {
	var err error
	if u.Transient {
		err = txn.Insert(u)
		if err == nil {
			u.Transient = false
		}
	} else {
		_, err = txn.Update(u)
	}
	return err
}
