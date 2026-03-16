package domain

import "time"

type Trade struct {
	ID          string
	Market      string
	BuyOrderID  string
	SellOrderID string
	Price       int64
	Size        int64
	Timestamp   time.Time
}
