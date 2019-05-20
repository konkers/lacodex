package model

import "time"

// Image stores an image by the hash of it's contents.
type Image struct {
	Hash string `storm:"id"`
	Data []byte
}

// ImageMetadata stores metadata bout an image.
//
// More than one ImageMetadata may exist for a single Image record.
type ImageMetadata struct {
	Pk         int       `storm:"id,increment"`
	Hash       string    `storm:"index"`
	CapturedAt time.Time `storm:"index"`
	FileName   string    `storm:"index"`
}
