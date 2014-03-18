package models

import (
	"github.com/coopernurse/gorp"
	"strconv"
)

type UserApp struct {
	UserAppId int
	UserId    int
	AppURL    string
	Name      string
	Path      string
}

func GetUserApps(txn *gorp.Transaction, session map[string]string) ([]*UserApp, error) {
	id, err := strconv.Atoi(session["user"])
	if err != nil {
		return nil, err
	}

	var apps []*UserApp
	_, err = txn.Select(&apps, "select * from dispatch_app where userid = $1", id)
	if err != nil {
		return nil, err
	}
	return apps, nil
}
