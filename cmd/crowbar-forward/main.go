package main

import (
	"flag"
	"fmt"

	"github.com/q3k/crowbar"
)

func main() {
	local := flag.String("local", "127.0.0.1:22", "Local address to bind to.")
	remote := flag.String("remote", "HOST:PORT", "Remote address to establish tunnel to.")
	server := flag.String("server", "http://example.com:80/", "Crowbar server to use.")
	flag.Parse()

	fmt.Println(*local, *remote, *server)

	_, err := crowbar.Connect(*server, "q3k", "dupa.8", *remote)
	fmt.Println(err)
}
