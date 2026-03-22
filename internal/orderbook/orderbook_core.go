package orderbook

import (
	"time"
	"trade-exchange-engine/internal/domain"
)

// #genai OrderBook holds the in-memory book state for one market.
type OrderBook struct {
	Market     string
	BuyOrders  map[int64]*PriceLevel
	SellOrders map[int64]*PriceLevel

	bidPrices  []int64
	askPrices  []int64
	orderIndex map[string]*OrderNode
}

func NewOrderBook(market string) *OrderBook {
	return &OrderBook{
		Market:     market,
		BuyOrders:  make(map[int64]*PriceLevel),
		SellOrders: make(map[int64]*PriceLevel),
		orderIndex: make(map[string]*OrderNode),
	}
}

func (o *OrderBook) AddOrder(order *domain.Order) {
	node := &OrderNode{Order: order}
	if order.Side == domain.OrderSideBuy {
		price := order.Price
		if buyOrders, exists := o.BuyOrders[price]; !exists || buyOrders.IsEmpty() {
			pl := NewPriceLevel(price)
			o.BuyOrders[price] = pl
			o.insertBidPrice(price)
		}
		o.BuyOrders[price].AddOrder(node)
		o.orderIndex[order.ID] = node
	} else {
		price := order.Price
		if sellOrders, exists := o.SellOrders[price]; !exists || sellOrders.IsEmpty() {
			pl := NewPriceLevel(price)
			o.SellOrders[price] = pl
			o.insertAskPrice(price)
		}
		o.SellOrders[price].AddOrder(node)
		o.orderIndex[order.ID] = node
	}
}

// #genai RemoveOrderByID removes an order from the index without cancelling it.
func (o *OrderBook) RemoveOrderByID(id string) {
	delete(o.orderIndex, id)
}

// #genai GetOrder returns the order if it exists in the book's index, nil otherwise.
func (o *OrderBook) GetOrder(id string) *domain.Order {
	node, ok := o.orderIndex[id]
	if !ok {
		return nil
	}
	return node.Order
}

func (o *OrderBook) DeleteOrder(orderId string) bool {
	node, ok := o.orderIndex[orderId]
	if !ok {
		return false
	}
	order := node.Order
	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = time.Now()
	priceLevel := node.Level
	priceLevel.RemoveOrder(node)
	if priceLevel.IsEmpty() {
		if order.Side == domain.OrderSideBuy {
			o.RemoveBidPrice(priceLevel.Price)
		} else {
			o.RemoveAskPrice(priceLevel.Price)
		}
	}
	delete(o.orderIndex, orderId)
	return true
}
