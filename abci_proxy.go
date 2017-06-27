package main

import (
	"flag"
	"os"
	"fmt"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	abcicli "github.com/tendermint/abci/client"
	"github.com/MultiverseHQ/abci_proxy/proxy"
	"github.com/tendermint/abci/server"
)

func main() {
	fmt.Printf("\n")
	fmt.Printf("Welcome to Multiverse\n")
	fmt.Printf("\n")
	fmt.Printf("This is the ABCi Proxy to handle validators sets changes on the flyyyyy .\n")
	fmt.Printf("\n")
	fmt.Printf("<3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3\n")
	fmt.Printf("\n")
	addrPtr := flag.String("addr", "tcp://0.0.0.0:46658", "Listen address")
	abciPtr := flag.String("abci", "socket", "socket | grpc")
	proxyPtr := flag.String("proxy", "tcp://0.0.0.0:46658", "Address of next abci app")
	flag.Parse()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	next := abcicli.NewSocketClient(*proxyPtr, true)

	// Start the listener
	srv, err := server.NewServer(*addrPtr, *abciPtr, proxy.NewProxyApp(next, []byte("echo")))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if _, err := srv.Start(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

}