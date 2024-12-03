package dataextractor

type Job struct {
	Id          int64  `json:"id"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	BytesSynced int64  `json:"bytesSynced"`
	RowsSynced  int64  `json:"rowsSynced"`
}
