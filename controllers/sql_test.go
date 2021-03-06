package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nuveo/prest/api"
	"github.com/nuveo/prest/config"
)

func TestMain(m *testing.M) {
	config.InitConf()
	createMockScripts(config.PrestConf.QueriesPath)
	writeMockScripts(config.PrestConf.QueriesPath)

	code := m.Run()

	removeMockScripts(config.PrestConf.QueriesPath)
	os.Exit(code)
}

func TestExecuteScriptQuery(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/testing/script-get/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := ExecuteScriptQuery(r, "fulltable", "get_all")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Write(resp)
	}))

	r.HandleFunc("/testing/script-post/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := ExecuteScriptQuery(r, "fulltable", "write_all")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Write(resp)
	}))

	ts := httptest.NewServer(r)
	defer ts.Close()

	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"Execute script GET method", "/testing/script-get/?field1=gopher", "GET", 200},
		{"Execute script POST method", "/testing/script-post/?field1=gopherzin&field2=pereira", "POST", 200},
		// errors
		{"Execute script GET method invalid", "/testing/script-get/?nonexistent=gopher", "GET", 400},
		{"Execute script POST method invalid", "/testing/script-post/?nonexistent=gopher", "POST", 400},
	}

	apiReq := api.Request{}
	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, ts.URL+tc.url, apiReq, tc.method, tc.status, "ExecuteScriptQuery")
	}
}

func TestExecuteFromScripts(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/_QUERIES/{queriesLocation}/{script}", ExecuteFromScripts)
	server := httptest.NewServer(router)
	defer server.Close()

	r := api.Request{}

	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"Get results using scripts by GET method", "/_QUERIES/fulltable/get_all?field1=gopher", "GET", 200},
		{"Get results using scripts by POST method", "/_QUERIES/fulltable/write_all?field1=gopherzin&field2=pereira", "POST", 200},
		{"Get results using scripts by PUT method", "/_QUERIES/fulltable/put_all?field1=trump&field2=pereira", "PUT", 200},
		{"Get results using scripts by PATCH method", "/_QUERIES/fulltable/patch_all?field1=temer&field2=trump", "PATCH", 200},
		{"Get results using scripts by DELETE method", "/_QUERIES/fulltable/delete_all?field1=trump", "DELETE", 200},
		// errors
		{"Get errors using nonexistent folder", "/_QUERIES/fullnon/delete_all?field1=trump", "DELETE", 400},
		{"Get errors using nonexistent script", "/_QUERIES/fulltable/some_com_all?field1=trump", "DELETE", 400},
		{"Get errors with invalid params in script", "/_QUERIES/fulltable/get_all?column1=gopher", "GET", 400},
		{"Get errors with invalid execution of sql", "/_QUERIES/fulltable/create_table?field1=test7", "POST", 400},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, server.URL+tc.url, r, tc.method, tc.status, "ExecuteFromScripts")
	}
}
