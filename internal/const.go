package internal

import "time"

//nolint:staticcheck // Требует убрать суффиксы, что снизит прозрачность
const (
	FifteenMin           = 15 * time.Minute //nolint:revive // Требует убрать суффиксы, что снизит прозрачность
	OneWeek              = 24 * 7 * time.Hour
	TenSec               = 10 * time.Second //nolint:revive // Требует убрать суффиксы, что снизит прозрачность
	FiveSec              = 5 * time.Second  //nolint:revive // Требует убрать суффиксы, что снизит прозрачность
	OneMin               = 60 * time.Second //nolint:revive // Требует убрать суффиксы, что снизит прозрачность
	FiveMin              = 5 * time.Minute  //nolint:revive // Требует убрать суффиксы, что снизит прозрачность
	TwoSec               = 2 * time.Second  //nolint:revive // Требует убрать суффиксы, что снизит прозрачность
	MaxAgeForAccessToken = 3600 * 24

	MaxAgeForRefreshToken = 3600 * 24 * 7
)
