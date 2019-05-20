package testutil

// T is an interface to abstract testing.T.  This lets us test failure
// conditions of test infrastructure.
type T interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}
