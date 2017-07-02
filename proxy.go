package abciproxy

import (
	"fmt"
	"net/http"

	abcicli "github.com/tendermint/abci/client"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/lib/server"
	tmlog "github.com/tendermint/tmlibs/log"
)

type ValidatorSetChange struct {
	Diffs           []*types.Validator
	ScheduledHeight uint64
}

// ProxyApplication is a super-simple proxy example.
// It just passes (almost) everything to another abci application
// However, if the CheckTX/DeliverTX starts with a given prefix, it echos the result
type ProxyApplication struct {
	types.BaseApplication
	next   abcicli.Client
	logger tmlog.Logger

	// to change concurrently the validator set
	lastHeight   uint64
	diffsChannel chan ValidatorSetChange
	diffs        map[uint64]ValidatorSetChange
}

var _ types.Application = &ProxyApplication{}

func NewProxyApp(next abcicli.Client) *ProxyApplication {
	return NewProxyAppWithLogger(next, tmlog.NewNopLogger())
}

func NewProxyAppWithLogger(next abcicli.Client, logger tmlog.Logger) *ProxyApplication {
	return &ProxyApplication{
		next:   next,
		logger: logger,
		//TODO: maybe a buffer of one isn't enough.
		diffsChannel: make(chan ValidatorSetChange, 1),
		diffs:        make(map[uint64]ValidatorSetChange),
		lastHeight:   0,
	}
}

func (app *ProxyApplication) Info() (resInfo types.ResponseInfo) {
	LogCall(app.logger)
	// TODO: better error handling!
	info, _ := app.next.InfoSync()
	return info
}

func (app *ProxyApplication) SetOption(key string, value string) (log string) {
	LogCall(app.logger, "key", key, "value", value)
	// TODO: better error handling!
	res := app.next.SetOptionSync(key, value)
	return res.Log
}

func (app *ProxyApplication) DeliverTx(tx []byte) types.Result {
	LogCall(app.logger, "tx", tx)
	return app.next.DeliverTxSync(tx)
}

func (app *ProxyApplication) CheckTx(tx []byte) types.Result {
	LogCall(app.logger, "tx", tx)
	return app.next.CheckTxSync(tx)
}

func (app *ProxyApplication) Commit() types.Result {
	LogCall(app.logger)
	return app.next.CommitSync()
}

func (app *ProxyApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	LogCall(app.logger, "query", reqQuery)
	// TODO: better error handling!
	res, _ := app.next.QuerySync(reqQuery)
	return res
}

func (app *ProxyApplication) InitChain(validators []*types.Validator) {
	LogCall(app.logger, "validators", validators)
	// TODO: better error handling!
	_ = app.next.InitChainSync(validators)
}

func (app *ProxyApplication) BeginBlock(hash []byte, header *types.Header) {
	LogCall(app.logger, "hash", hash, "header", header)
	// TODO: better error handling!
	_ = app.next.BeginBlockSync(hash, header)
}

func (app *ProxyApplication) ChangeValidator(newValidators []*types.Validator, targetHeight uint64) error {
	app.logger.Debug("received new validator set",
		"validators", newValidators,
		"targetHeight", targetHeight)
	if targetHeight <= app.lastHeight {
		return fmt.Errorf("Could not schedule for a block height back in time (wanted:%d, current:%d)", targetHeight, app.lastHeight)
	}
	app.diffsChannel <- ValidatorSetChange{
		Diffs:           newValidators,
		ScheduledHeight: targetHeight,
	}
	return nil
}

func mergeValidatorDiffs(merged, newChanges []*types.Validator) []*types.Validator {
	//TODO: maybe we should require something more involved, like notsubmitting two same changes
	return append(merged, newChanges...)
}

func (app *ProxyApplication) EndBlock(height uint64) (resEndBlock types.ResponseEndBlock) {
	LogCall(app.logger, "height", height)
	app.lastHeight = height
	// TODO: better error handling!
	res, _ := app.next.EndBlockSync(height)

	//
	haveValidators := true
	for haveValidators == true {
		select {
		case change := <-app.diffsChannel:
			if change.ScheduledHeight < height {
				app.logger.Error("got a validator change too late",
					"currentHeight", height,
					"targetHeight", change.ScheduledHeight,
					"diffs", change.Diffs)
			}
			if c, ok := app.diffs[change.ScheduledHeight]; ok == true {
				c.Diffs = mergeValidatorDiffs(c.Diffs, change.Diffs)
				app.diffs[change.ScheduledHeight] = c
			} else {
				app.diffs[change.ScheduledHeight] = change
			}
		default:
			//if none pending simply stop the consuming
			haveValidators = false
		}
	}

	if c, ok := app.diffs[height]; ok == true {
		res.Diffs = c.Diffs
		delete(app.diffs, height)
	} else {
		// remove any target app wanted changes
		res.Diffs = nil
	}

	if len(res.Diffs) != 0 {
		app.logger.Debug("submitting new validators", "validators", res.Diffs)
	}

	return res
}

type CurrentHeightResult struct {
	Height uint64 `json:"height"`
}

type ChangeValidatorsResult struct {
}

func (app *ProxyApplication) StartRPCServerr(rcpAddress string) {

	var routes = map[string]*rpcserver.RPCFunc{
		"change_validators": rpcserver.NewRPCFunc(func(validators []*types.Validator, scheduledHeight uint64) (*ChangeValidatorsResult, error) {
			err := app.ChangeValidator(validators, scheduledHeight)
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
		logger.Info("start RPC server", "address", rpcAddress)
		_, err := rpcserver.StartHTTPServer(rpcAddress, mux, logger)
		if err != nil {
			panic(err)
		}
	}()

}
