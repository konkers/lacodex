package model

import "time"

// ImageMetadata stores metadata bout an image.
//
// More than one ImageMetadata may exist for a single Image record.
type ImageMetadata struct {
	Id         int       `storm:"id,increment"`
	Hash       string    `storm:"index"`
	CapturedAt time.Time `storm:"index"`
	FileName   string    `storm:"index,unique"`
	Record     int       `storm:"index"`
}
