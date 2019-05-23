package lacodex

import (
	"fmt"
	"net/http"
)

func httpError(w http.ResponseWriter, statusCode int, format string, args ...interface{}) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(format, args...)))
}
