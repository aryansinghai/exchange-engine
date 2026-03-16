package orderbook

import (
	"slices"
	"time"
	"trade-exchange-engine/internal/domain"

	"github.com/google/uuid"
)

type OrderBook struct {
	Market     string
	BuyOrders  map[int64]*domain.PriceLevel
	SellOrders map[int64]*domain.PriceLevel

	bidPrices  []int64
	askPrices  []int64
	orderIndex map[string]*domain.Order
}

func NewOrderBook(market string) *OrderBook {
	return &OrderBook{
		Market:     market,
		BuyOrders:  make(map[int64]*domain.PriceLevel),
		SellOrders: make(map[int64]*domain.PriceLevel),
		orderIndex: make(map[string]*domain.Order),
	}
}

// ProcessOrder
func (o *OrderBook) ProcessOrder(order *domain.Order) []*domain.Trade {
	var trades []*domain.Trade

	if order.Side == domain.OrderSideBuy {
		trades = o.processBuyOrder(order)
	} else {
		trades = o.processSellOrder(order)
	}

	return trades
}

func (o *OrderBook) processBuyOrder(order *domain.Order) []*domain.Trade {
	var trades []*domain.Trade

	for {
		buyQtyLeft := order.Size - order.FilledSize
		if buyQtyLeft <= 0 {
			break
		}
		bestAskPrice, ok := o.BestAskPrice()
		if !ok {
			break
		}
		if order.Price < bestAskPrice {
			break
		}
		queue := o.SellOrders[bestAskPrice]
		if queue.IsEmpty() {
			break
		}
		bestAskOrder := queue.GetHead()
		sellQtyLeft := bestAskOrder.Size - bestAskOrder.FilledSize
		tradeSize := min(buyQtyLeft, sellQtyLeft)
		if tradeSize == 0 {
			break
		}

		trades = append(trades, &domain.Trade{
			ID:          uuid.New().String(),
			Market:      o.Market,
			BuyOrderID:  order.ID,
			SellOrderID: bestAskOrder.ID,
			Price:       bestAskPrice,
			Size:        tradeSize,
			Timestamp:   time.Now(),
		})
		order.FilledSize += tradeSize
		bestAskOrder.FilledSize += tradeSize
		if bestAskOrder.Size-bestAskOrder.FilledSize == 0 {
			_ = queue.PopHead()
			if queue.IsEmpty() {
				o.RemoveAskPrice(bestAskPrice)
			}
		}
	}

	if order.FilledSize == order.Size {
		order.Status = domain.OrderStatusFilled
	} else {
		if order.FilledSize > 0 {
			order.Status = domain.OrderStatusPartiallyFilled
		} else {
			order.Status = domain.OrderStatusOpen
		}
		o.AddOrder(order)
	}

	return trades
}

func (o *OrderBook) processSellOrder(order *domain.Order) []*domain.Trade {
	var trades []*domain.Trade

	for {
		sellQtyLeft := order.Size - order.FilledSize
		if sellQtyLeft <= 0 {
			break
		}
		bestBidPrice, ok := o.BestBidPrice()
		if !ok {
			break
		}
		if order.Price > bestBidPrice {
			break
		}
		queue := o.BuyOrders[bestBidPrice]
		if queue.IsEmpty() {
			break
		}
		bestBidOrder := queue.GetHead()
		buyQtyLeft := bestBidOrder.Size - bestBidOrder.FilledSize
		tradeSize := min(sellQtyLeft, buyQtyLeft)
		if tradeSize == 0 {
			break
		}

		trades = append(trades, &domain.Trade{
			ID:          uuid.New().String(),
			Market:      o.Market,
			BuyOrderID:  bestBidOrder.ID,
			SellOrderID: order.ID,
			Price:       bestBidPrice,
			Size:        tradeSize,
			Timestamp:   time.Now(),
		})
		order.FilledSize += tradeSize
		bestBidOrder.FilledSize += tradeSize
		if bestBidOrder.Size-bestBidOrder.FilledSize == 0 {
			_ = queue.PopHead()
			if queue.IsEmpty() {
				o.RemoveBidPrice(bestBidPrice)
			}
		}
	}

	if order.FilledSize == order.Size {
		order.Status = domain.OrderStatusFilled
	} else {
		if order.FilledSize > 0 {
			order.Status = domain.OrderStatusPartiallyFilled
		} else {
			order.Status = domain.OrderStatusOpen
		}
		o.AddOrder(order)
	}

	return trades
}

func (o *OrderBook) AddOrder(order *domain.Order) {
	if order.Side == domain.OrderSideBuy {
		price := order.Price
		if buyOrders, exists := o.BuyOrders[price]; !exists || buyOrders.IsEmpty() {
			o.insertBidPrice(price)
		}
		o.BuyOrders[price].AddOrder(order)
		o.orderIndex[order.ID] = order
	} else {
		price := order.Price
		if sellOrders, exists := o.SellOrders[price]; !exists || sellOrders.IsEmpty() {
			o.insertAskPrice(price)
		}
		o.SellOrders[price].AddOrder(order)
		o.orderIndex[order.ID] = order
	}
}

func (o *OrderBook) insertBidPrice(price int64) {

	idx := 0
	for idx < len(o.bidPrices) && o.bidPrices[idx] > price {
		idx++
	}
	o.bidPrices = append(o.bidPrices, 0)
	copy(o.bidPrices[idx+1:], o.bidPrices[idx:])
	o.bidPrices[idx] = price
}

func (o *OrderBook) insertAskPrice(price int64) {
	idx := 0
	for idx < len(o.askPrices) && o.askPrices[idx] < price {
		idx++
	}
	o.askPrices = append(o.askPrices, 0)
	copy(o.askPrices[idx+1:], o.askPrices[idx:])
	o.askPrices[idx] = price
}

// best bid price
func (o *OrderBook) BestBidPrice() (int64, bool) {
	if len(o.bidPrices) == 0 {
		return 0, false
	}
	return o.bidPrices[0], true
}

// best ask price
func (o *OrderBook) BestAskPrice() (int64, bool) {
	if len(o.askPrices) == 0 {
		return 0, false
	}
	return o.askPrices[0], true
}

// remove bid price
func (o *OrderBook) RemoveBidPrice(price int64) {
	pl, ok := o.BuyOrders[price]
	if !ok {
		return
	}
	if !pl.IsEmpty() {
		return
	}
	delete(o.BuyOrders, price)
	o.bidPrices = slices.DeleteFunc(o.bidPrices, func(p int64) bool {
		return p == price
	})
}

// remove ask price
func (o *OrderBook) RemoveAskPrice(price int64) {

	pl, ok := o.SellOrders[price]
	if !ok {
		return
	}

	if !pl.IsEmpty() {
		return
	}

	delete(o.SellOrders, price)

	o.askPrices = slices.DeleteFunc(o.askPrices, func(p int64) bool {
		return p == price
	})
}

func (o *OrderBook) DeleteOrder(orderId string) bool {
	order, ok := o.orderIndex[orderId]
	if !ok {
		return false
	}
	priceLevel := order.PriceLevel
	priceLevel.RemoveOrder(order)
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
