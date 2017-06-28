package abciproxy

import (
	"runtime"

	tmlog "github.com/tendermint/tmlibs/log"
)

func LogCall(logger tmlog.Logger, keyvals ...interface{}) {
	fpc, _, _, ok := runtime.Caller(1)
	if ok == false {
		panic("Could not found caller")
	}
	fun := runtime.FuncForPC(fpc)
	logger.Debug(fun.Name(), keyvals...)
}
