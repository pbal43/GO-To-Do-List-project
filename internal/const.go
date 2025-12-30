package internal

import "time"

const (
	MinFifteen            = 15 * time.Minute
	OneWeek               = 24 * 7 * time.Hour
	SecTen                = 10 * time.Second
	SecFive               = 5 * time.Second
	MinOne                = 60 * time.Second
	MinFive               = 5 * time.Minute
	SecTwo                = 2 * time.Second
	MaxAgeForAccessToken  = 3600 * 24
	MaxAgeForRefreshToken = 3600 * 24 * 7
)
