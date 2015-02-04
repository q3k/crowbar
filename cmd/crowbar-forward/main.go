package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/q3k/crowbar"
	"github.com/howeyc/gopass"
)

func main() {
	local := flag.String("local", "127.0.0.1:1122", "Local address to bind to, or - for stdin/out")
	remote := flag.String("remote", "HOST:PORT", "Remote address to establish tunnel to.")
	server := flag.String("server", "http://127.0.0.1:8080/", "Crowbar server to use.")
	username := flag.String("username", "", "Username to use.")
	password := flag.String("password", "", "Password to use, or empty for interactive getpass.")
	flag.Parse()

	if *username == "" {
		fmt.Fprintf(os.Stderr, "Username must be given.\n")
		return
	}
	if *password == "" {
		fmt.Fprintf(os.Stderr, "Password: ")
		*password = string(gopass.GetPasswd())
	}

	if *local == "-" {
		remoteConn, err := crowbar.Connect(*server, *username, *password, *remote)
		if err != nil {
			panic(err)
		}
		go io.Copy(remoteConn, os.Stdin)
		io.Copy(os.Stdout, remoteConn)
	} else {
		localListen, err := net.Listen("tcp", *local)
		if err != nil {
			panic(err)
		}
		for {
			localConn, err := localListen.Accept()
			if err != nil {
				panic(err)
			}
			remoteConn, err := crowbar.Connect(*server, *username, *password, *remote)
			if err != nil {
				panic(err)
			}
			go io.Copy(localConn, remoteConn)
			io.Copy(remoteConn, localConn)
		}
	}
}
