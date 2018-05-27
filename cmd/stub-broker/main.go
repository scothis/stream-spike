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
	"fmt"
	"net/http"

	"github.com/go-martini/martini"
)

var (
	subscriptions = make(map[string]map[string]struct{})
	exists        = struct{}{}
)

func main() {
	m := martini.Classic()

	m.Post("/", func(req *http.Request, res http.ResponseWriter) {
		fmt.Printf("Recieved request for %s\n", req.Host)
		res.WriteHeader(http.StatusAccepted)
	})

	m.Group("/streams", func(r martini.Router) {
		r.Put("/:stream", func(params martini.Params, res http.ResponseWriter) {
			stream := params["stream"]
			fmt.Printf("Create stream %s\n", stream)
			subscriptions[stream] = make(map[string]struct{})
			res.WriteHeader(http.StatusAccepted)
		})
		r.Delete("/:stream", func(params martini.Params, res http.ResponseWriter) {
			stream := params["stream"]
			fmt.Printf("Delete stream %s\n", stream)
			delete(subscriptions, stream)
			res.WriteHeader(http.StatusAccepted)
		})
	})

	m.Run()
}
