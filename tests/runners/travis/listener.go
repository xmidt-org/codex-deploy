/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

/*
headers:
content-length →21
content-type →application/json
date →Tue, 15 Jan 2019 00:25:58 GMT
x-tr1d1um-build →0.1.1-311
x-tr1d1um-flavor →cherry
x-tr1d1um-region →ch2g
x-tr1d1um-server →tr1d1um-cd-wgs.webpa.comcast.net
x-tr1d1um-start-time →13 Dec 18 03:44 UTC

message body:
{
    "message": "Success"
}
*/

func StartListener() {
	r := mux.NewRouter()
	// caducues-ct.xmidt.comcast.net:8090/api/v2/hook
	r.HandleFunc("/api/v2/hook", HandlePostRequest).
		Methods("POST")
}

//This will handle just webhook registration.  It will validate the registration.
func HandlePostRequest(w http.ResponseWriter, r *http.Request) {

}
