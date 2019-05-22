package lacodex

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

func httpError(w http.ResponseWriter, statusCode int, format string, args ...interface{}) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(format, args...)))
}

func warnIfError(err error, format string, args ...interface{}) {
	if err != nil {
		s := fmt.Sprintf(format, args...)
		s += ": " + err.Error()
		glog.Warning(s)
	}
}
