package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"net/http"
	"strconv"

	"code.google.com/p/go-uuid/uuid"

	"github.com/q3k/crowbar"
)

func connectHandler(w http.ResponseWriter, r *http.Request) {
	remote_host := r.URL.Query().Get("remote_host")
	if remote_host == "" {
		crowbar.WriteHTTPError(w, "Invalid host")
		return
	}
	remote_port, err := strconv.Atoi(r.URL.Query().Get("remote_port"))
	if err != nil || remote_port > 0xFFFF {
		crowbar.WriteHTTPError(w, "Invalid port number.")
		return
	}

	workerUuid := uuid.New()
	commandChannel := make(chan workerCommand, 10)
	responseChannel := make(chan workerResponse, 10)
	fmt.Printf("Connecting to %s:%d...\n", remote_host, remote_port)
	remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remote_host, remote_port))
	if err != nil {
		crowbar.WriteHTTPError(w, fmt.Sprintf("Could not connect to %s:%d", remote_host, remote_port))
		return
	}

	newWorker := worker{remote: remote, commandChannel: commandChannel, responseChannel: responseChannel, uuid: workerUuid}
	workerMap[workerUuid] = newWorker

	crowbar.WriteHTTPOK(w, workerUuid)

	go socketWorker(newWorker)
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	workerUuid := r.URL.Query().Get("uuid")
	if worker, ok := workerMap[workerUuid]; ok {
		if r.Method == "POST" {
			r.ParseForm()
			if b64_parts, ok := r.Form["data"]; ok {
				b64 := b64_parts[0]
				decodeLen := base64.StdEncoding.DecodedLen(len(b64))
				data := make([]byte, decodeLen)
				n, err := base64.StdEncoding.Decode(data, []byte(b64))
				if err != nil {
					crowbar.WriteHTTPError(w, "Could not decode B64.")
				} else {
					worker.commandChannel <- workerCommand{command: command_data, extra: data[:n]}
					crowbar.WriteHTTPOK(w, "Sent.")
				}
			} else {
				crowbar.WriteHTTPError(w, "Data is required.")
			}
		} else {
			response := <-worker.responseChannel
			switch response.response {
			case response_data:
				crowbar.WriteHTTPData(w, response.extra_byte)
			case response_quit:
				crowbar.WriteHTTPQuit(w, response.extra_string)
			}
		}
	} else {
		crowbar.WriteHTTPError(w, "No such UUID")
	}
}

func main() {
	var listen = flag.String("listen", "0.0.0.0:8080", "Address to bind HTTP server to")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "Server starting on %s...\n", *listen)
	http.HandleFunc(crowbar.EndpointConnect, connectHandler)
	http.HandleFunc(crowbar.EndpointSync, syncHandler)
	http.ListenAndServe(*listen, nil)
}
