// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/scylladb/scylla-go-driver"
)

var HOST_NAME = os.Getenv("HOST_NAME")
var SCYLLA_DB_HOST = os.Getenv("SCYLLA_DB_HOST")

var ClientRegisterer = registerer("auth-token")

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
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}
	config, _ := extra["auth-token"].(map[string]interface{})
	path, _ := config["path"].(string)
	logger.Debug(fmt.Sprintf("The plugin is now hijacking the path %s", path))

	ctx := context.Background()
	cfg := scylla.DefaultSessionConfig("emsdb", SCYLLA_DB_HOST)
	session, err := scylla.NewSession(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == path {
			jti := uuid.New()
			iat := time.Now()

			q, err := session.Prepare(requestCtx, "SELECT login, password FROM emsdb.users WHERE id=?")
			if err != nil {
				return nil, err
			}
			res, err := q.BindInt64(0, 64).Exec(requestCtx)
			if err != nil {
				return nil, err
			}

			print(string(res))

			jwt := map[string]any{
				"access_token": map[string]any{
					"aud": HOST_NAME,
					"jti": jti,
					"iat": iat,
					"exp": 3600,
				},
				"refresh_token": map[string]any{
					"aud": HOST_NAME,
					"jti": jti,
					"iat": iat,
					"exp": 604800,
				},
			}
			body, err := json.Marshal(jwt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(body)
			logger.Debug("token:", string(body))
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
