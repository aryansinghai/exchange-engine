package orderbook

import (
	"testing"
	"time"
	"trade-exchange-engine/internal/domain"
)

func TestOrderbook(t *testing.T) {
	orderbook := NewOrderBook("BTC/USD")
	orderbook.AddOrder(&domain.Order{
		ID:        "1",
		Market:    "BTC/USD",
		Side:      domain.OrderSideBuy,
		Size:      1.0,
		Price:     10000.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	orderbook.AddOrder(&domain.Order{
		ID:        "2",
		Market:    "BTC/USD",
		Side:      domain.OrderSideSell,
		Size:      1.0,
		Price:     10000.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if len(orderbook.BuyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(orderbook.BuyOrders))
	}
	if len(orderbook.SellOrders) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(orderbook.SellOrders))
	}
	if orderbook.BuyOrders[10000.0].GetHead().ID != "1" {
		t.Errorf("Expected buy order 1, got %s", orderbook.BuyOrders[10000.0].GetHead().ID)
	}
	if orderbook.SellOrders[10000.0].GetHead().ID != "2" {
		t.Errorf("Expected sell order 2, got %s", orderbook.SellOrders[10000.0].GetHead().ID)
	}
}
