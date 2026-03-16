package domain

import "time"

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderStatus string

const (
	OrderStatusOpen            OrderStatus = "OPEN"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
)

type Order struct {
	ID     string
	UserID string
	Market string
	Type   OrderType
	Side   OrderSide
	Status OrderStatus

	Size       int64
	Price      int64
	FilledSize int64

	CreatedAt time.Time
	UpdatedAt time.Time

	Prev       *Order
	Next       *Order
	PriceLevel *PriceLevel
}
