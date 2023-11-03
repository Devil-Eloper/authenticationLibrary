package datamodels

import "time"

type TokenObject struct {
	AccessToken   string
	ExpiresIn     int
	RetrievedDate time.Time
}
