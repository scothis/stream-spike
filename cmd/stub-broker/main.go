/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/go-martini/martini"
)

var (
	subscriptions = sync.Map{}
)

type exists interface{}

func splitStreamName(host string) string {
	chunks := strings.Split(host, ".")
	stream := chunks[0]
	return stream
}

func streamSubscribers(stream string) *sync.Map {
	subscribersInterface, ok := subscriptions.Load(stream)
	if !ok {
		return nil
	}
	subscribers := subscribersInterface.(sync.Map)
	return &subscribers
}

func main() {
	m := martini.Classic()

	m.Post("/", func(req *http.Request, res http.ResponseWriter) {
		host := req.Host
		fmt.Printf("Recieved request for %s\n", host)
		streamName := splitStreamName(host)
		subscribers := streamSubscribers(streamName)
		if subscribers == nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusAccepted)
		go func() {
			// make upstream requests
			client := &http.Client{}

			subscribers.Range(func(key, value interface{}) bool {
				subscribed, _ := key.(string)

				go func() {
					url := fmt.Sprintf("http://%s/", subscribed)
					request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
					if err != nil {
						fmt.Printf("Unable to create subscriber request %v", err)
					}
					// TODO pass other headers
					request.Header.Set("content-type", req.Header.Get("content-type"))
					_, err = client.Do(request)
					if err != nil {
						fmt.Printf("Unable to complete subscriber request %v", err)
					}
				}()

				return true
			})
		}()
	})

	m.Group("/streams/:stream", func(r martini.Router) {
		r.Put("", func(params martini.Params, res http.ResponseWriter) {
			stream := params["stream"]
			fmt.Printf("Create stream %s\n", stream)
			_, _ = subscriptions.LoadOrStore(stream, sync.Map{})
			res.WriteHeader(http.StatusAccepted)
		})
		r.Delete("", func(params martini.Params, res http.ResponseWriter) {
			stream := params["stream"]
			fmt.Printf("Delete stream %s\n", stream)
			subscriptions.Delete(stream)
			res.WriteHeader(http.StatusAccepted)
		})

		r.Group("/subscriptions/:subscription", func(r martini.Router) {
			r.Put("", func(params martini.Params, res http.ResponseWriter) {
				stream := params["stream"]
				subscription := params["subscription"]
				subscribers := streamSubscribers(stream)
				if subscribers == nil {
					res.WriteHeader(http.StatusNotFound)
					return
				}
				fmt.Printf("Create subscription %s for stream %s\n", subscription, stream)
				// TODO store subscription params
				subscribers.Store(subscription, "")
				res.WriteHeader(http.StatusAccepted)
			})
			r.Delete("", func(params martini.Params, res http.ResponseWriter) {
				stream := params["stream"]
				subscription := params["subscription"]
				subscribers := streamSubscribers(stream)
				if subscribers == nil {
					res.WriteHeader(http.StatusNotFound)
					return
				}
				fmt.Printf("Delete subscription %s for stream %s\n", subscription, stream)
				subscriptions.Delete(subscription)
				res.WriteHeader(http.StatusAccepted)
			})
		})
	})

	m.Run()
}
