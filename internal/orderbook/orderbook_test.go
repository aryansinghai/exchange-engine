package orderbook

import (
	"testing"
	"time"
	"trade-exchange-engine/internal/domain"
)

// #genai

func newOrder(id string, side domain.OrderSide, price, size int64) *domain.Order {
	return &domain.Order{
		ID:            id,
		Market:        "BTC/USD",
		Type:          domain.OrderTypeLimit,
		Side:          side,
		Size:          size,
		Price:         price,
		RemainingSize: size,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestAddOrder_SingleBuy(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 10000, 5))

	if len(ob.BuyOrders) != 1 {
		t.Fatalf("expected 1 buy price level, got %d", len(ob.BuyOrders))
	}
	if ob.BuyOrders[10000].GetHead().Order.ID != "1" {
		t.Fatal("expected order 1 at head")
	}
}

func TestAddOrder_SingleSell(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideSell, 10000, 5))

	if len(ob.SellOrders) != 1 {
		t.Fatalf("expected 1 sell price level, got %d", len(ob.SellOrders))
	}
	if ob.SellOrders[10000].GetHead().Order.ID != "1" {
		t.Fatal("expected order 1 at head")
	}
}

func TestAddOrder_MultipleSamePriceFIFO(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("2", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("3", domain.OrderSideBuy, 100, 1))

	pl := ob.BuyOrders[100]
	if pl.GetHead().Order.ID != "1" {
		t.Fatal("FIFO violated: head should be order 1")
	}
	if pl.GetTail().Order.ID != "3" {
		t.Fatal("FIFO violated: tail should be order 3")
	}
}

func TestAddOrder_MultiplePriceLevels(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("2", domain.OrderSideBuy, 200, 1))
	ob.AddOrder(newOrder("3", domain.OrderSideBuy, 150, 1))

	if len(ob.BuyOrders) != 3 {
		t.Fatalf("expected 3 buy price levels, got %d", len(ob.BuyOrders))
	}
}

func TestBestBidPrice_Ordering(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("2", domain.OrderSideBuy, 300, 1))
	ob.AddOrder(newOrder("3", domain.OrderSideBuy, 200, 1))

	best, ok := ob.BestBidPrice()
	if !ok || best != 300 {
		t.Fatalf("expected best bid 300, got %d", best)
	}
}

func TestBestAskPrice_Ordering(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideSell, 300, 1))
	ob.AddOrder(newOrder("2", domain.OrderSideSell, 100, 1))
	ob.AddOrder(newOrder("3", domain.OrderSideSell, 200, 1))

	best, ok := ob.BestAskPrice()
	if !ok || best != 100 {
		t.Fatalf("expected best ask 100, got %d", best)
	}
}

func TestBestBidPrice_EmptyBook(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	_, ok := ob.BestBidPrice()
	if ok {
		t.Fatal("expected no best bid on empty book")
	}
}

func TestBestAskPrice_EmptyBook(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	_, ok := ob.BestAskPrice()
	if ok {
		t.Fatal("expected no best ask on empty book")
	}
}

func TestDeleteOrder_Exists(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 5))

	ok := ob.DeleteOrder("1")
	if !ok {
		t.Fatal("expected delete to succeed")
	}

	_, hasBid := ob.BestBidPrice()
	if hasBid {
		t.Fatal("expected no bids after deleting only order")
	}
}

func TestDeleteOrder_NotFound(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ok := ob.DeleteOrder("nonexistent")
	if ok {
		t.Fatal("expected delete of nonexistent order to return false")
	}
}

func TestDeleteOrder_SetsStatusCancelled(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	order := newOrder("1", domain.OrderSideSell, 100, 5)
	ob.AddOrder(order)
	ob.DeleteOrder("1")

	if order.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected status CANCELLED, got %s", order.Status)
	}
}

func TestDeleteOrder_MiddleOfQueue(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("2", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("3", domain.OrderSideBuy, 100, 1))

	ob.DeleteOrder("2")

	pl := ob.BuyOrders[100]
	if pl.GetHead().Order.ID != "1" {
		t.Fatal("head should still be order 1")
	}
	if pl.GetTail().Order.ID != "3" {
		t.Fatal("tail should still be order 3")
	}
	if pl.GetHead().Next.Order.ID != "3" {
		t.Fatal("order 1 next should be order 3 after removing 2")
	}
}

func TestDeleteOrder_RemovesPriceLevelWhenEmpty(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 1))
	ob.AddOrder(newOrder("2", domain.OrderSideBuy, 200, 1))

	ob.DeleteOrder("1")

	if _, exists := ob.BuyOrders[100]; exists {
		t.Fatal("price level 100 should be removed when empty")
	}
	best, ok := ob.BestBidPrice()
	if !ok || best != 200 {
		t.Fatalf("expected best bid 200 after removing 100-level, got %d", best)
	}
}

func TestRemoveOrderByID(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	ob.AddOrder(newOrder("1", domain.OrderSideBuy, 100, 1))

	ob.RemoveOrderByID("1")

	ok := ob.DeleteOrder("1")
	if ok {
		t.Fatal("after RemoveOrderByID, DeleteOrder should return false (not in index)")
	}
}

func TestGetOrder_Exists(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	order := newOrder("1", domain.OrderSideBuy, 100, 5)
	ob.AddOrder(order)

	found := ob.GetOrder("1")
	if found == nil {
		t.Fatal("expected to find order")
	}
	if found.ID != "1" {
		t.Fatalf("expected order ID 1, got %s", found.ID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	ob := NewOrderBook("BTC/USD")
	found := ob.GetOrder("nonexistent")
	if found != nil {
		t.Fatal("expected nil for nonexistent order")
	}
}
