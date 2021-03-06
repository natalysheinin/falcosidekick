package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Issif/falcosidekick/outputs"
	"github.com/Issif/falcosidekick/types"
)

// checkpayloadHandler prints received falco's payload in stdout (for debug) of daemon
func checkpayloadHandler(w http.ResponseWriter, r *http.Request) {
	// Read body
	requestDump, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		log.Printf("[ERROR] : %v\n", err.Error())
	}
	w.Write([]byte(requestDump))
	log.Printf("[DEBUG] : Falco's Payload =  %v\n", string(requestDump))
}

// mainHandler is Falco Sidekick' main handler (default).
func mainHandler(w http.ResponseWriter, r *http.Request) {

	var falcopayload types.FalcoPayload

	if r.Body == nil {
		http.Error(w, "Please send a valid request body", 400)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&falcopayload)
	if err != nil && err.Error() != "EOF" || len(falcopayload.Output) == 0 {
		http.Error(w, "Please send a valid request body : "+err.Error(), 400)
		return
	}

	if os.Getenv("SLACK_TOKEN") != "" {
		go outputs.SlackPost(falcopayload)
	}
	if os.Getenv("DATADOG_TOKEN") != "" {
		go outputs.DatadogPost(falcopayload)
	}
	if os.Getenv("ALERTMANAGER_HOST_PORT") != "" {
		go outputs.AlertmanagerPost(falcopayload)
	}
}

// pingHandler is a simple handler to test if daemon is UP.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

// test sends a test event to all enabled outputs.
func test(w http.ResponseWriter, r *http.Request) {
	testEvent := `{"output":"This is a test from Falco Sidekick","priority":"Debug","rule":"Test rule", "output_fields": {"proc.name":"falcosidekick","user.name":"falcosidekick"}}`

	port = "2801"
	if lport, err := strconv.Atoi(os.Getenv("LISTEN_PORT")); err == nil {
		if lport > 0 && lport < 65536 {
			port = os.Getenv("LISTEN_PORT")
		}
	}

	resp, err := http.Post("http://localhost:"+port, "application/json", bytes.NewBuffer([]byte(testEvent)))
	if err != nil {
		log.Printf("[DEBUG] : Test Failed. Falcosidekick can't call itself\n")
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] : Test sent\n")
	if resp.StatusCode == http.StatusOK {
		log.Printf("[DEBUG] : Test OK (%v)\n", resp.Status)
	} else {
		log.Printf("[DEBUG] : Test KO (%v)\n", resp.Status)
	}
}
