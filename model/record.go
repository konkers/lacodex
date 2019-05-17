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

// Record represents a piece of interesting data in La Mulana
type Record struct {
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
