// Copyright 2014 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"net/http"
	"path"

	"bitbucket.org/tshannon/freehold/app"
	"bitbucket.org/tshannon/freehold/fail"
	"bitbucket.org/tshannon/freehold/log"
	"bitbucket.org/tshannon/freehold/permission"
)

type ApplicationInput struct {
	Id   *string `json:"id,omitempty"`
	File *string `json:"file,omitempty"`
}

func appGet(w http.ResponseWriter, r *http.Request) {
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}

	if !auth.canRead(permission.Application()) {
		four04(w, r)
		return
	}

	input := &ApplicationInput{}
	err = parseJson(r, input)
	if errHandled(err, w) {
		return
	}

	if input.Id != nil {
		a, err := app.Get(*input.Id)
		if errHandled(err, w) {
			return
		}

		if a == nil {
			four04(w, r)
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
			Data:   a,
		})
		return
	}

	apps, err := app.All()
	if errHandled(err, w) {
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   apps,
	})
	return

}

func appPost(w http.ResponseWriter, r *http.Request) {
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}

	if !auth.canWrite(permission.Application()) {
		errHandled(fail.New("You do not have permissions to install an application.", nil), w)
		return
	}

	input := &ApplicationInput{}
	err = parseJson(r, input)
	if errHandled(err, w) {
		return
	}
	if input.File == nil {
		errHandled(fail.New("JSON request must contain file property", nil), w)
		return
	}

	a, err := app.Install(*input.File, auth.User.Username())
	if errHandled(err, w) {
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   a,
	})

}

func appPut(w http.ResponseWriter, r *http.Request) {
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}

	if !auth.canWrite(permission.Application()) {
		errHandled(fail.New("You do not have permissions to update an application.", nil), w)
		return
	}

	input := &ApplicationInput{}
	err = parseJson(r, input)
	if errHandled(err, w) {
		return
	}
	if input.File == nil {
		errHandled(fail.New("JSON request must contain file property", nil), w)
		return
	}
	a, err := app.Upgrade(*input.File, auth.User.Username())
	if errHandled(err, w) {
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   a,
	})
}

func appDelete(w http.ResponseWriter, r *http.Request) {
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}

	if !auth.canWrite(permission.Application()) {
		errHandled(fail.New("You do not have permissions to uninstall an application.", nil), w)
		return
	}

	input := &ApplicationInput{}
	err = parseJson(r, input)
	if errHandled(err, w) {
		return
	}

	if input.Id == nil {
		errHandled(fail.New("JSON request must contain the id property", nil), w)
		return
	}

	err = app.Uninstall(*input.Id)
	if errHandled(err, w) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func appRootGet(w http.ResponseWriter, r *http.Request) {
	_, pth := splitRootAndPath(r.URL.Path)
	if pth != "/" {
		four04Page(w, r)
		return
	}
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}
	serveApp(w, r, appHandler.Root(r), auth)

}

func serveApp(w http.ResponseWriter, r *http.Request, appid string, auth *Auth) {
	a, err := app.Get(appid)

	if errHandled(err, w) {
		return
	}
	if a == nil {
		four04Page(w, r)
		return
	}

	root := a.Root
	if root == "" {
		log.Error(errors.New("App " + a.Id + " has no root specified."))
	}

	root = path.Join("/", a.Id, root)
	serveResource(w, r, root, auth)
}

func appAvailableGet(w http.ResponseWriter, r *http.Request) {
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}

	if !auth.canRead(permission.AppAvailable()) {
		errHandled(fail.New("You do not have permissions to view available applications.", nil), w)
		return
	}

	apps, fails, err := app.Available()
	if errHandled(err, w) {
		return
	}

	if len(fails) > 0 {
		respondJsend(w, &JSend{
			Status:   statusFail,
			Data:     apps,
			Failures: fails,
		})
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   apps,
	})
}

func appAvailablePost(w http.ResponseWriter, r *http.Request) {
	auth, err := authenticate(w, r)
	if errHandled(err, w) {
		return
	}

	if !auth.canWrite(permission.AppAvailable()) {
		errHandled(fail.New("You do not have permissions to post a new available application.", nil), w)
		return
	}

	input := &ApplicationInput{}
	err = parseJson(r, input)
	if errHandled(err, w) {
		return
	}
	if input.File == nil {
		errHandled(fail.New("JSON request must contain file property", nil), w)
		return
	}

	file, err := app.PostAvailable(*input.File)
	if errHandled(err, w) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   file,
	})
	return

}
