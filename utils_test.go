package lacodex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/glog"
)

func TestWarnIfError(t *testing.T) {
	warnings := glog.Stats.Warning.Lines()
	warnIfError(fmt.Errorf("err"), "Warning")
	assert.NotEqual(t, warnings, glog.Stats.Warning.Lines())
}
