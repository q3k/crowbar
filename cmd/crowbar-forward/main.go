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
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/q3k/crowbar"
)

func main() {
	local := flag.String("local", "127.0.0.1:1122", "Local address to bind to, or - for stdin/out")
	remote := flag.String("remote", "HOST:PORT", "Remote address to establish tunnel to.")
	server := flag.String("server", "http://127.0.0.1:8080/", "Crowbar server to use.")
	username := flag.String("username", "", "Username to use.")
	password := flag.String("password", "", "Password to use.")
	flag.Parse()

	if *username == "" {
		fmt.Fprintf(os.Stderr, "Username must be given.\n")
		return
	}
	if *password == "" {
		fmt.Fprintf(os.Stderr, "Password must be given.\n")
		return
	}

	if *local == "-" {
		remoteConn, err := crowbar.Connect(*server, *username, *password, *remote)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		go io.Copy(remoteConn, os.Stdin)
		io.Copy(os.Stdout, remoteConn)
	} else {
		localListen, err := net.Listen("tcp", *local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		for {
			localConn, err := localListen.Accept()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}
			remoteConn, err := crowbar.Connect(*server, *username, *password, *remote)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}
			go io.Copy(localConn, remoteConn)
			io.Copy(remoteConn, localConn)
		}
	}
}
