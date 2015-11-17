package rpc

import (
	"strings"

	"github.com/melange-app/nmcd/btcjson"
)

const (
	lowestValue  = "0"
	lookupLength = 50
	numToSee     = 20
)

func (r *Server) LookupPrefix(prefix string) ([]string, error) {
	var foundNames []string

	iteration := 0
	search := prefix
	for iteration < lookupLength {
		search = search + lowestValue

		currentSearch := search
		alreadyRun := false
	findLoop:
		for {
			cmd, err := btcjson.NewNameScanCmd(nil, currentSearch, numToSee)
			if err != nil {
				return nil, err
			}

			reply, err := r.Send(cmd)
			if err != nil {
				return nil, err
			}

			if reply.Result == nil {
				return nil, errNilReply
			}

			nameScan := reply.Result.([]*btcjson.NameScanResult)

			for index, result := range nameScan {
				// Don't want to repeat names.
				if alreadyRun && index == 0 {
					continue
				}

				if strings.HasPrefix(result.Name, prefix) {
					foundNames = append(foundNames, result.Name)
					currentSearch = result.Name
				} else {
					break findLoop
				}
			}
			alreadyRun = true
		}

		iteration++
	}

	return foundNames, nil
}

func (r *Server) LookupName(name string) (string, bool, error) {
	cmd, err := btcjson.NewNameShowCmd(nil, name)
	if err != nil {
		return "", false, err
	}

	reply, err := r.Send(cmd)
	if err != nil {
		return "", false, err
	}

	if reply.Result == nil {
		return "", false, errNilReply
	} else if reply.Error != nil {
		// A code of -4 indicates that the name was not found
		// in the database.
		if reply.Error.Code == -4 {
			return "", false, nil
		}

		return "", false, reply.Error
	}

	nameInfo := reply.Result.(*btcjson.NameInfoResult)

	if nameInfo.Expired {
		return "", false, nil
	}

	// Return the name
	return nameInfo.Value, true, nil
}
