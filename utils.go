package abciproxy

import (
	"fmt"
	"runtime"

	tmlog "github.com/tendermint/tmlibs/log"
)

func CallerName() string {
	fpc, _, _, ok := runtime.Caller(2)
	if ok == false {
		panic("Could not found caller")
	}
	fun := runtime.FuncForPC(fpc)
	return fun.Name()
}

func LogCall(logger tmlog.Logger, keyvals ...interface{}) {
	logger.Debug(CallerName(), keyvals...)
}

func NotYetImplemented() error {
	return fmt.Errorf("%s is not et implemented", CallerName())
}
