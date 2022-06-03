package watcher

import (
	"errors"

	"github.com/iter8-tools/iter8/abn/util"
	"github.com/iter8-tools/iter8/base/log"
)

func Lookup(name string, user string) (*Version, error) {
	// if user is not provided, use a random string
	if user == "" {
		user = util.RandomString(24)
		log.Logger.Debug("no user, using ", user)
	}

	// get app from name, fail if not present
	app, ok := apps[name]
	if !ok {
		return nil, errors.New("no versions found for application")
	}

	// use consistent to get version for user, fail if not present
	v, err := app.c.Get(user)
	if err != nil {
		return nil, err
	}
	version, ok := app.versions[v]
	if !ok {
		return nil, errors.New("can't find version")
	}

	return &version, nil
}
