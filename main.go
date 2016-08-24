package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
)

var httpClient = http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 128,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	},
}

func main() {
	log.Printf("Batch Deleter server starting")
	app := cli.App("batch-deleter", "A RESTful API for doing batch deletes from concept writers")
	port := app.StringOpt("port", "8080", "Port to listen on")

	app.Action = func() {
		runServer(*port)
	}
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	log.SetLevel(log.InfoLevel)
	log.Infof("Application started with args %s", os.Args)
	app.Run(os.Args)
}

func runServer(port string) {
	servicesRouter := mux.NewRouter()
	servicesRouter.HandleFunc("/batchdelete", batchDelete).Methods("POST")

	http.Handle("/", servicesRouter)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}

func batchDelete(w http.ResponseWriter, req *http.Request) {
	var body io.Reader = req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		unzipped, err := gzip.NewReader(req.Body)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer unzipped.Close()
		body = unzipped
	}

	dec := json.NewDecoder(body)
	inst := instructions{}
	err := dec.Decode(&inst)

	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Got instructions=%v", inst)

	basicAuth := req.Header.Get("Authorization")

	for _, host := range inst.Hosts {
		deleteAllUuids(host, inst.Path, inst.Uuids, basicAuth)
	}
}

func deleteAllUuids(host string, path string, uuids []string, basicAuth string) {
	log.Printf("Deleting for host=%s path=%s uuids=%s", host, path, uuids)
	for _, uuid := range uuids {
		log.Printf("uuid=%s", uuid)
		reqURL := host + "/" + path + "/" + uuid
		request, err := http.NewRequest("DELETE", reqURL, nil)
		request.Header.Set("Authorization", basicAuth)
		if err != nil {
			log.Errorf("Could not create request for reqURL=%s, err=%s", reqURL, err)
			continue
		}
		log.Printf("About to Delete %s", request.URL)
		resp, err := httpClient.Do(request)
		defer resp.Body.Close()
		if err != nil {
			log.Errorf("Error for reqURL=%s, err=%s", reqURL, err)
			continue
		}
		log.Infof("Response=%v", resp.StatusCode)
		if http.StatusNoContent != resp.StatusCode && http.StatusNotFound != resp.StatusCode {
			log.Errorf("Unexpected status code for reqURL=%s, code=%v", reqURL, resp.StatusCode)
			continue
		}
	}
}

func writeJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", errorMsg))
}
