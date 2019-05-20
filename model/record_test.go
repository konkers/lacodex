package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyphraseType(t *testing.T) {
	values := []struct {
		val KeyphraseType
		enc string
	}{
		{KeyphraseTypeNone, "none"},
		{KeyphraseTypeBlue, "blue"},
		{KeyphraseTypeGreen, "green"},
	}

	for _, v := range values {
		enc, err := v.val.MarshalText()
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, v.enc, string(enc))

		var val KeyphraseType
		err = (&val).UnmarshalText([]byte(v.enc))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, v.val, val)
	}

	v := KeyphraseType(-1)
	_, err := v.MarshalText()
	if err == nil {
		t.Error("Expected error.")
	}

	err = (&v).UnmarshalText([]byte(""))
	if err == nil {
		t.Error("Expected error.")
	}

}

func TestRecordType(t *testing.T) {
	values := []struct {
		val RecordType
		enc string
	}{
		{RecordTypeTent, "tent"},
		{RecordTypeMailer, "mailer"},
		{RecordTypeScanner, "scanner"},
	}

	for _, v := range values {
		enc, err := v.val.MarshalText()
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, v.enc, string(enc))

		var val RecordType
		err = (&val).UnmarshalText([]byte(v.enc))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, v.val, val)
	}

	v := RecordType(-1)
	_, err := v.MarshalText()
	if err == nil {
		t.Error("Expected error.")
	}

	err = (&v).UnmarshalText([]byte(""))
	if err == nil {
		t.Error("Expected error.")
	}

}
