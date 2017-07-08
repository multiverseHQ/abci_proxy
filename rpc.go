package abciproxy

import (
	"net/http"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/rpc/lib/server"
)

type CurrentHeightResult struct {
	Height uint64 `json:"height"`
}

type ChangeValidatorsResult struct {
}

type ValidatorPowerChange struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Power  uint64        `json:"power"`
}

func (app *ProxyApplication) StartRPCServer(rpcAddress string) {

	var routes = map[string]*rpcserver.RPCFunc{
		"change_validators": rpcserver.NewRPCFunc(func(validators []*ValidatorPowerChange, scheduledHeight uint64) (*ChangeValidatorsResult, error) {
			toABCI := make([]*types.Validator, 0, len(validators))
			for _, vpc := range validators {
				toABCI = append(toABCI, &types.Validator{
					PubKey: vpc.PubKey.Bytes(),
					Power:  vpc.Power,
				})
			}
			err := app.ChangeValidators(toABCI, scheduledHeight)
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
		_, err := rpcserver.StartHTTPServer(rpcAddress, mux, app.logger)
		if err != nil {
			panic(err)
		}
	}()

}
