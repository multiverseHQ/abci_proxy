package main

import (
	"fmt"
	"os"

	"github.com/MultiverseHQ/abci_proxy"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"

	abcicli "github.com/tendermint/abci/client"
	"github.com/tendermint/abci/server"
)

var logger tmlog.Logger
var opts options

func init() {
	opts = ParseOptions()

	baselogger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stderr))

	if opts.Verbose == true {
		logger = tmlog.NewFilter(baselogger, tmlog.AllowAll())
		logger.Info("Debug output")
	} else {
		logger = tmlog.NewFilter(baselogger, tmlog.AllowInfo())
	}
}

func Execute() error {
	fmt.Printf("\n")
	fmt.Printf("Welcome to Multiverse\n")
	fmt.Printf("\n")
	fmt.Printf("This is the ABCi Proxy to handle validators sets changes on the flyyyyy .\n")
	fmt.Printf("\n")
	fmt.Printf("<3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3 <3\n")
	fmt.Printf("\n")

	next := abcicli.NewSocketClient(opts.ProxyAddress, true)

	// Start the listener
	srv, err := server.NewServer(opts.Address, opts.ABCIType, abciproxy.NewProxyApp(next, []byte("echo")))
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if _, err := srv.Start(); err != nil {
		return err
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

	return nil
}

func main() {
	if err := Execute(); err != nil {
		logger.Error("unhandled error", "error", err)
	}
}
