package main

import "flag"

type options struct {
	Address    string
	ABCIType   string
	AppAddress string
	Verbose    bool
}

func ParseOptions() options {
	var opts options

	flag.StringVar(&opts.Address, "addr", "tcp://0.0.0.0:46658", "Listen address for tendermind node")
	flag.StringVar(&opts.ABCIType, "abci", "socket", "socket | grpc")
	flag.StringVar(&opts.AppAddress, "proxy", "tcp://0.0.0.0:46659", "Address of next ABCI app")
	flag.BoolVar(&opts.Verbose, "verbose", false, "verbose output")
	flag.BoolVar(&opts.Verbose, "v", false, "verbose output")
	flag.Parse()

	return opts
}
