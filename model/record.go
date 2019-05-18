package model

import "fmt"

// KeyphraseType enumerates the types of keyphrase.
type KeyphraseType int

const (
	// KeyphraseTypeNone represents text that is not a keyphrase.
	KeyphraseTypeNone KeyphraseType = iota

	// KeyphraseTypeBlue represents a blue keyphrase.
	KeyphraseTypeBlue

	// KeyphraseTypeGreen represents a greenkeyphrase.
	KeyphraseTypeGreen
)

// RecordType enumerates the types of records.
type RecordType int

const (
	// RecordTypeTent is a record taken inside a tent.
	RecordTypeTent RecordType = iota

	// RecordTypeMailer is a record taken from the xelpud mailer.
	RecordTypeMailer

	// RecordTypeScanner is a record taken from a tablet.
	RecordTypeScanner
)

// Record represents a piece of interesting data in La Mulana
type Record struct {
	Type       RecordType                 `json:"type"`
	Text       string                     `json:"text"`
	Keyphrases map[KeyphraseType][]string `json:"keyphrases"`
}

func (t KeyphraseType) MarshalText() ([]byte, error) {
	switch t {
	case KeyphraseTypeNone:
		return []byte("none"), nil
	case KeyphraseTypeBlue:
		return []byte("blue"), nil
	case KeyphraseTypeGreen:
		return []byte("green"), nil
	}

	return nil, fmt.Errorf("Unknown KeyphraseType %v", t)
}

func (t *KeyphraseType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "none":
		*t = KeyphraseTypeNone
		return nil
	case "blue":
		*t = KeyphraseTypeBlue
		return nil
	case "green":
		*t = KeyphraseTypeGreen
		return nil
	}

	return fmt.Errorf("Unknown KeyphraseType %s", string(text))
}

func (t RecordType) MarshalText() ([]byte, error) {
	switch t {
	case RecordTypeTent:
		return []byte("tent"), nil
	case RecordTypeMailer:
		return []byte("mailer"), nil
	case RecordTypeScanner:
		return []byte("scanner"), nil
	}

	return nil, fmt.Errorf("Unknown RecordType %v", t)
}

func (t *RecordType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "tent":
		*t = RecordTypeTent
		return nil
	case "mailer":
		*t = RecordTypeMailer
		return nil
	case "scanner":
		*t = RecordTypeScanner
		return nil
	}

	return fmt.Errorf("Unknown RecordType %s", string(text))
}
