package domain

import (
	"errors"
	"time"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeIOC    OrderType = "IOC"
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

	Size          int64
	Price         int64
	FilledSize    int64
	RemainingSize int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// #genai NewOrder creates an Order with invariants enforced at construction time.
func NewOrder(id, userID, market string, orderType OrderType, side OrderSide, size, price int64) *Order {
	now := time.Now()
	return &Order{
		ID:            id,
		UserID:        userID,
		Market:        market,
		Type:          orderType,
		Side:          side,
		Status:        OrderStatusOpen,
		Size:          size,
		Price:         price,
		FilledSize:    0,
		RemainingSize: size,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// #genai Validate checks domain invariants independent of the matching engine.
func (o *Order) Validate() error {
	if o.ID == "" {
		return errors.New("order ID is required")
	}
	if o.Size <= 0 {
		return errors.New("order size must be positive")
	}
	if o.Type != OrderTypeLimit && o.Type != OrderTypeMarket && o.Type != OrderTypeIOC {
		return errors.New("invalid order type")
	}
	if o.Type != OrderTypeMarket && o.Price <= 0 {
		return errors.New("limit/IOC order price must be positive")
	}
	if o.Side != OrderSideBuy && o.Side != OrderSideSell {
		return errors.New("invalid order side")
	}
	return nil
}
