package abciproxy

import (
	"bytes"
	"fmt"

	abcicli "github.com/tendermint/abci/client"
	"github.com/tendermint/abci/types"
	tmlog "github.com/tendermint/tmlibs/log"
)

// ProxyApplication is a super-simple proxy example.
// It just passes (almost) everything to another abci application
// However, if the CheckTX/DeliverTX starts with a given prefix, it echos the result
type ProxyApplication struct {
	types.BaseApplication
	next   abcicli.Client
	logger tmlog.Logger
	prefix []byte
}

var _ types.Application = &ProxyApplication{}

func NewProxyApp(next abcicli.Client, prefix []byte) *ProxyApplication {
	return NewProxyAppWithLogger(next, prefix, tmlog.NewNopLogger())

}

func NewProxyAppWithLogger(next abcicli.Client, prefix []byte, logger tmlog.Logger) *ProxyApplication {
	return &ProxyApplication{
		next:   next,
		prefix: prefix,
		logger: logger,
	}
}

func (app *ProxyApplication) makeEcho(tx []byte) string {
	return fmt.Sprintf("Echo: %s", string(tx[len(app.prefix):]))
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
	if bytes.HasPrefix(tx, app.prefix) {
		return types.NewResultOK(nil, app.makeEcho(tx))
	}
	return app.next.DeliverTxSync(tx)
}

func (app *ProxyApplication) CheckTx(tx []byte) types.Result {
	LogCall(app.logger, "tx", tx)
	if bytes.HasPrefix(tx, app.prefix) {
		return types.NewResultOK(nil, app.makeEcho(tx))
	}
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

func (app *ProxyApplication) EndBlock(height uint64) (resEndBlock types.ResponseEndBlock) {
	LogCall(app.logger, "height", height)
	// TODO: better error handling!
	res, _ := app.next.EndBlockSync(height)
	return res
}
