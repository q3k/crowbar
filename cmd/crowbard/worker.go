package main

import (
    "fmt"
    "net"
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

