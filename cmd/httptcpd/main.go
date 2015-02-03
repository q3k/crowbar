package main

import (
    "fmt"
    "net/http"
    "strconv"
    "net"
    "encoding/base64"

    "code.google.com/p/go-uuid/uuid"

    "github.com/q3k/crowbar"
)

const command_data string = "DATA"
const command_stop string = "STOP"

const response_data string = "DATA"
const response_quit string = "QUIT"

type workerCommand struct {
    command string
    extra []byte
}

type workerResponse struct {
    response string
    extra_byte []byte
    extra_string string
}

func workerQuit(responseChannel chan workerResponse, extra string) {
    response := workerResponse{response: response_quit, extra_string: extra}
    responseChannel <- response
}

type worker struct {
    remote net.Conn
    uuid string
    commandChannel chan workerCommand
    responseChannel chan workerResponse
}

func socketWorker(wWorker worker) {
    fmt.Println("Worker starting...")

    continue_loop := true
    go func() {
        for continue_loop {
            data := make([]byte, 512)
            n, err := wWorker.remote.Read(data)
            if err != nil {
                workerQuit(wWorker.responseChannel, "Read error.")
                wWorker.commandChannel <- workerCommand{command: "bogus"}
                continue_loop = false
            } else {
                wWorker.responseChannel <- workerResponse{response: response_data, extra_byte: data[:n]}
            }
        }
    }()

    for continue_loop {
        command := <-wWorker.commandChannel
        switch command.command {
        case command_stop:
            workerQuit(wWorker.responseChannel, "Worker stopped.")
            continue_loop = false
        case command_data:
            wWorker.remote.Write(command.extra)
        }
    }
    fmt.Println("Worker exiting...")
}

var workerMap = map[string]worker{}

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
            response := <-worker.responseChannel;
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
    http.HandleFunc(crowbar.EndpointConnect, connectHandler)
    http.HandleFunc(crowbar.EndpointSync, syncHandler)
    http.ListenAndServe(":8080", nil)
}
