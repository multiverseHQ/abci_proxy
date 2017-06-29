package abciproxy

import (
	"encoding/binary"

	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
)

// TestApplication is to be used in our test to actually test that our
// application is going from observer state to validator state. Its
// simply a counter, but we register all calls to its method.
type TestApplication struct {
	types.BaseApplication

	hashCount int
	txCount   int
	serial    bool

	InfoCalls      InterceptedMethod
	CommitCalls    InterceptedMethod
	SetOptionCalls InterceptedMethod
	DeliverTxCalls InterceptedMethod
	CheckTxCalls   InterceptedMethod
	QueryCalls     InterceptedMethod
	EndBlockCalls  InterceptedMethod
}

func NewTestApplication(serial bool) *TestApplication {
	return &TestApplication{
		serial:         serial,
		InfoCalls:      NewInterceptedMethod("Info"),
		CommitCalls:    NewInterceptedMethod("Commit"),
		SetOptionCalls: NewInterceptedMethod("SetOption"),
		DeliverTxCalls: NewInterceptedMethod("DeliverTx"),
		CheckTxCalls:   NewInterceptedMethod("CheckTx"),
		QueryCalls:     NewInterceptedMethod("Query"),
		EndBlockCalls:  NewInterceptedMethod("EndBlock"),
	}
}

func (app *TestApplication) Info() types.ResponseInfo {
	app.InfoCalls.Notify()
	return types.ResponseInfo{Data: cmn.Fmt("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}
}

func (app *TestApplication) SetOption(key string, value string) (log string) {
	app.SetOptionCalls.Notify(key, value)
	if key == "serial" && value == "on" {
		app.serial = true
	}
	return ""
}

func (app *TestApplication) DeliverTx(tx []byte) types.Result {
	app.DeliverTxCalls.Notify(tx)
	if app.serial {
		if len(tx) > 8 {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Max tx size is 8 bytes, got %d", len(tx)))
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(tx):], tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue != uint64(app.txCount) {
			return types.ErrBadNonce.SetLog(cmn.Fmt("Invalid nonce. Expected %v, got %v", app.txCount, txValue))
		}
	}
	app.txCount++
	return types.OK
}

func (app *TestApplication) CheckTx(tx []byte) types.Result {
	if app.serial {
		if len(tx) > 8 {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Max tx size is 8 bytes, got %d", len(tx)))
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(tx):], tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return types.ErrBadNonce.SetLog(cmn.Fmt("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue))
		}
	}
	return types.OK
}

func (app *TestApplication) Commit() types.Result {
	app.CommitCalls.Notify()
	app.hashCount++
	if app.txCount == 0 {
		return types.OK
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	return types.NewResultOK(hash, "")
}

func (app *TestApplication) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	app.QueryCalls.Notify()
	switch reqQuery.Path {
	case "hash":
		return types.ResponseQuery{Value: []byte(cmn.Fmt("%v", app.hashCount))}
	case "tx":
		return types.ResponseQuery{Value: []byte(cmn.Fmt("%v", app.txCount))}
	default:
		return types.ResponseQuery{Log: cmn.Fmt("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}
	}
}

func (app *TestApplication) EndBlock(height uint64) (resEndBlock types.ResponseEndBlock) {
	app.EndBlockCalls.Notify(height)
	return types.ResponseEndBlock{}
}
