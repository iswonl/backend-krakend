// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const GOOGLE_TOKEN_ENDPOINT = "https://oauth2.googleapis.com/token"

var GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
var GOOGLE_CLIENT_SECRET = os.Getenv("GOOGLE_CLIENT_SECRET")
var GOOGLE_CODE_REDIRECT_URL = os.Getenv("GOOGLE_CODE_REDIRECT_URL")

var ClientRegisterer = registerer("google-code")

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

	config, _ := extra["google-code"].(map[string]interface{})

	path, _ := config["path"].(string)
	logger.Debug(fmt.Sprintf("The plugin is now hijacking the path %s", path))

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == path {
			fmt.Println("code:", req.URL.Query().Get("code"))
			post_req, _ := http.NewRequest("POST", GOOGLE_TOKEN_ENDPOINT, nil)
			q := post_req.URL.Query()
			q.Add("client_id", GOOGLE_CLIENT_ID)
			q.Add("client_secret", GOOGLE_CLIENT_SECRET)
			q.Add("code", req.URL.Query().Get("code"))
			q.Add("grant_type", "authorization_code")
			q.Add("redirect_uri", GOOGLE_CODE_REDIRECT_URL)
			post_req.URL.RawQuery = q.Encode()
			fmt.Println("request:", post_req.URL.String())

			client := &http.Client{}
			resp, err := client.Do(post_req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("token:", string(body))
			w.Write(body)
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
