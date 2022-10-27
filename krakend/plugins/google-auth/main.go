// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const GOOGLE_OAUTH2_ENDPOINT = "https://accounts.google.com/o/oauth2/auth"

var GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
var GOOGLE_ACCESS_SCOPES = os.Getenv("GOOGLE_ACCESS_SCOPES")
var GOOGLE_CODE_REDIRECT_URL = os.Getenv("GOOGLE_CODE_REDIRECT_URL")

var ClientRegisterer = registerer("google-auth")

type registerer string

var logger Logger = nil

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ClientRegisterer))
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(_ context.Context, extra map[string]interface{}) (http.Handler, error) {
	// check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	config, _ := extra["google-auth"].(map[string]interface{})

	path, _ := config["path"].(string)
	logger.Debug(fmt.Sprintf("The plugin is now hijacking the path %s", path))

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == path {
			req, _ := http.NewRequest("GET", GOOGLE_OAUTH2_ENDPOINT, nil)
			q := req.URL.Query()
			q.Add("client_id", GOOGLE_CLIENT_ID)
			q.Add("redirect_uri", GOOGLE_CODE_REDIRECT_URL)
			q.Add("response_type", "code")
			q.Add("scope", GOOGLE_ACCESS_SCOPES)
			q.Add("access_type", "offline")
			req.URL.RawQuery = q.Encode()

			http.Redirect(w, req, req.URL.String(), http.StatusSeeOther)
			logger.Debug("redirect:", req.URL.String())
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if resp.Body == nil {
			return
		}
		io.Copy(w, resp.Body)
		resp.Body.Close()

	}), nil
}

func main() {}

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}
