package abciproxy

import (
	"net/http"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/lib/server"
)

type CurrentHeightResult struct {
	Height uint64 `json:"height"`
}

type ChangeValidatorsResult struct {
}

func (app *ProxyApplication) StartRPCServer(rpcAddress string) {

	var routes = map[string]*rpcserver.RPCFunc{
		"change_validators": rpcserver.NewRPCFunc(func(validators []*types.Validator, scheduledHeight uint64) (*ChangeValidatorsResult, error) {
			err := app.ChangeValidators(validators, scheduledHeight)
			return &ChangeValidatorsResult{}, err
		}, "validators,scheduled_height"),
		"current_height": rpcserver.NewRPCFunc(func() (*CurrentHeightResult, error) {
			return &CurrentHeightResult{Height: app.lastHeight}, nil
		}, ""),
	}

	mux := http.NewServeMux()
	rpcserver.RegisterRPCFuncs(mux, routes, app.logger)
	wm := rpcserver.NewWebsocketManager(routes, nil)
	wm.SetLogger(app.logger)
	mux.HandleFunc("/websocket/endpoint", wm.WebsocketHandler)

	go func() {
		app.
			logger.Info("start RPC server", "address", rpcAddress)
		_, err := rpcserver.StartHTTPServer(rpcAddress, mux, app.logger)
		if err != nil {
			panic(err)
		}
	}()

}
