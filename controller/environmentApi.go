/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/fission/fission"
	"strings"
)

func (api *API) EnvironmentApiList(w http.ResponseWriter, r *http.Request) {
	envs, err := api.EnvironmentStore.List()
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	resp, err := json.Marshal(envs)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	api.respondWithSuccess(w, resp)
}

func (api *API) EnvironmentApiCreate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	var env fission.Environment
	err = json.Unmarshal(body, &env)
	if err != nil {
		log.Printf("Failed to unmarshal request body: [%v]", body)
		api.respondWithError(w, err)
		return
	}

	uid, err := api.EnvironmentStore.Create(&env)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	m := &fission.Metadata{Name: env.Metadata.Name, Uid: uid}
	resp, err := json.Marshal(m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	api.respondWithSuccess(w, resp)
}

func (api *API) EnvironmentApiGet(w http.ResponseWriter, r *http.Request) {
	var m fission.Metadata
	var env *fission.Environment

	vars := mux.Vars(r)
	m.Name = vars["environment"]
	m.Uid = r.FormValue("uid") // empty if uid is absent

	// TODO should just be inserted into the store, if workflow-engine is available
	// Lookup environment, unless it is the special 'workflow' case
	if strings.EqualFold(m.Name, "workflow") {
		env = &fission.Environment{
			Metadata: fission.Metadata{
				Name: "workflow",
			},
		}
	} else {
		fetched, err := api.EnvironmentStore.Get(&m)
		if err != nil {
			api.respondWithError(w, err)
			return
		}
		env = fetched
	}

	resp, err := json.Marshal(env)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	api.respondWithSuccess(w, resp)
}

func (api *API) EnvironmentApiUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["environment"]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	var env fission.Environment
	err = json.Unmarshal(body, &env)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	if name != env.Metadata.Name {
		err = fission.MakeError(fission.ErrorInvalidArgument, "Environment name doesn't match URL")
		api.respondWithError(w, err)
		return
	}

	uid, err := api.EnvironmentStore.Update(&env)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	m := &fission.Metadata{Name: env.Metadata.Name, Uid: uid}
	resp, err := json.Marshal(m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}
	api.respondWithSuccess(w, resp)
}

func (api *API) EnvironmentApiDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var m fission.Metadata
	m.Name = vars["environment"]

	m.Uid = r.FormValue("uid") // empty if uid is absent
	if len(m.Uid) == 0 {
		log.WithFields(log.Fields{"httpTrigger": m.Name}).Info("Deleting all versions")
	}

	err := api.EnvironmentStore.Delete(m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	api.respondWithSuccess(w, []byte(""))
}
