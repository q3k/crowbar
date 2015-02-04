// Copyright (c) 2015, Segiusz 'q3k' Bazanski <sergiusz@bazanski.pl>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

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
	extra   []byte
}

type workerResponse struct {
	response     string
	extra_byte   []byte
	extra_string string
}

func workerQuit(responseChannel chan workerResponse, extra string) {
	response := workerResponse{response: response_quit, extra_string: extra}
	responseChannel <- response
}

type worker struct {
	remote          net.Conn
	uuid            string
	commandChannel  chan workerCommand
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
				continue_loop = false
				wWorker.commandChannel <- workerCommand{command: "bogus"}
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
