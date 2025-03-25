package main

import (
	"flag"
	"github.com/armon/go-socks5"
	"log"
)

var (
	flagListen = flag.String("listen", ":8686", "Address to listen on.")
)

func main() {
	flag.Parse()
	srv, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %v", *flagListen)
	log.Fatal(srv.ListenAndServe("tcp", *flagListen))
}
