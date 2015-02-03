package main

import (
    "fmt"
    "net/http"
    "strconv"
    "net"
    "encoding/base64"

    "code.google.com/p/go-uuid/uuid"
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
    remote_port, err := strconv.Atoi(r.URL.Query().Get("remote_port"))
    if err != nil || remote_port > 0xFFFF {
        http.Error(w, "ERROR: That's not a valid port number.", http.StatusInternalServerError)
        return
    }

    workerUuid := uuid.New()
    commandChannel := make(chan workerCommand, 10)
    responseChannel := make(chan workerResponse, 10)
    fmt.Printf("Connecting to %s:%d...\n", remote_host, remote_port)
    remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remote_host, remote_port))
    if err != nil {
        http.Error(w, "ERROR: Could not connect.", http.StatusInternalServerError)
        return
    }

    newWorker := worker{remote: remote, commandChannel: commandChannel, responseChannel: responseChannel, uuid: workerUuid}
    workerMap[workerUuid] = newWorker

    fmt.Fprintf(w, "OK: %s", workerUuid)

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
                    http.Error(w, "ERROR: What the hell did you send me?", http.StatusInternalServerError)
                } else {
                    worker.commandChannel <- workerCommand{command: command_data, extra: data[:n]}
                    fmt.Fprintf(w, "OK: Sent.")
                }
            } else {
                http.Error(w, "ERROR: You forgot to send me data.", http.StatusInternalServerError)
            }
        } else {
            response := <-worker.responseChannel;
            switch response.response {
                case response_data:
                    data := base64.StdEncoding.EncodeToString(response.extra_byte)
                    fmt.Fprintf(w, "DATA: %s", data)
                case response_quit:
                    fmt.Fprintf(w, "QUIT: %s", response.extra_string)
            }
        }
    } else {
        http.Error(w, "ERROR: No such uuid.", http.StatusInternalServerError)
    }
}

func main() {
    http.HandleFunc("/connect/", connectHandler)
    http.HandleFunc("/sync/", syncHandler)
    http.ListenAndServe(":8080", nil)
}
