package main

import (
	"time"
	"net/http"

	"github.com/golang/glog"

	"github.com/piiproxy/api"
	"github.com/gorilla/mux"
)

const port = ":8888"
const readTimeout = 10 * time.Second
const writeTimeout = 10 * time.Second
const maxHeaderBytes = 1 << 20

func main() {
        // config.LoadGlogConfig();
	api.LoadDataFromJson(); // TODO: better way later

	startServer();
}

func startServer() {
	httpServer := httpServerWithHandlers();
	glog.Fatal(httpServer.ListenAndServe());
}

func httpServerWithHandlers() *http.Server {

	router := mux.NewRouter()

	// Handlers - IAM permissions for reading the PII stash tables
	router.HandleFunc("/stash-read/assign", api.AssignReadAccess);
	router.HandleFunc("/stash-read/remove", api.RemoveReadAccess);
	router.HandleFunc("/iamrole", api.GetRoleInfo);

	// Handlers - PII Submission and Lookup API.
	router.HandleFunc("/pii/1/_table/{table}", api.CreateTable).Methods("PUT");
	router.HandleFunc("/pii/1/{table}/{id}", api.PutPIIData).Methods("PUT");
	router.HandleFunc("/pii/1/{table}/{id}", api.GetPIIData).Methods("GET");

	http.Handle("/", router);

	// log.Fatal(http.ListenAndServe(port, nil))
	httpServer := &http.Server{
		Addr:           port,
		Handler:        nil,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	};

	return httpServer;
}
