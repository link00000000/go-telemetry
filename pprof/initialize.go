package pprof

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func init() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	runtime.SetMutexProfileFraction(16)
	runtime.SetBlockProfileRate(16)
}
