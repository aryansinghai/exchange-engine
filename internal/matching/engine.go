package matching

import (
	"errors"
	"time"
	"trade-exchange-engine/internal/domain"
	"trade-exchange-engine/internal/orderbook"

	"github.com/google/uuid"
)

// #genai PriceQuantity represents aggregated quantity at a single price level.
type PriceQuantity struct {
	Price    int64
	Quantity int64
}

// #genai BookSnapshot is a read-only view of the current order book state.
type BookSnapshot struct {
	Market string
	Bids   []PriceQuantity
	Asks   []PriceQuantity
}

// #genai MatchingService defines the public contract for the matching engine.
type MatchingService interface {
	SubmitOrder(order *domain.Order) ([]*domain.Trade, error)
	CancelOrder(orderId string) error
	GetBookSnapshot() BookSnapshot
}

// #genai Engine drives order matching against an OrderBook.
type Engine struct {
	book *orderbook.OrderBook
}

func NewEngine(book *orderbook.OrderBook) *Engine {
	return &Engine{book: book}
}

func (e *Engine) SubmitOrder(order *domain.Order) ([]*domain.Trade, error) {
	if err := order.Validate(); err != nil {
		return nil, err
	}
	initializeOrder(order)

	var trades []*domain.Trade
	var err error

	switch order.Type {
	case domain.OrderTypeLimit, domain.OrderTypeIOC:
		if order.Side == domain.OrderSideBuy {
			trades = e.processBuyOrder(order)
		} else {
			trades = e.processSellOrder(order)
		}
	case domain.OrderTypeMarket:
		trades, err = e.processMarketOrder(order)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid order type")
	}

	return trades, nil
}

func (e *Engine) processMarketOrder(order *domain.Order) ([]*domain.Trade, error) {
	var trades []*domain.Trade
	var err error

	if order.Side == domain.OrderSideBuy {
		trades, err = e.processBuyMarketOrder(order)
	} else {
		trades, err = e.processSellMarketOrder(order)
	}
	return trades, err
}

func (e *Engine) processBuyMarketOrder(order *domain.Order) ([]*domain.Trade, error) {
	var trades []*domain.Trade

	for {
		buyQtyLeft := order.RemainingSize
		if buyQtyLeft <= 0 {
			break
		}
		bestAskPrice, ok := e.book.BestAskPrice()
		if !ok {
			break
		}
		queue := e.book.SellOrders[bestAskPrice]
		if queue.IsEmpty() {
			e.book.RemoveAskPrice(bestAskPrice)
			continue
		}
		bestAskNode := queue.GetHead()
		bestAskOrder := bestAskNode.Order
		sellQtyLeft := bestAskOrder.RemainingSize
		tradeSize := min(buyQtyLeft, sellQtyLeft)
		if tradeSize == 0 {
			break
		}
		trades = append(trades, &domain.Trade{
			ID:          uuid.New().String(),
			Market:      e.book.Market,
			BuyOrderID:  order.ID,
			SellOrderID: bestAskOrder.ID,
			Price:       bestAskPrice,
			Size:        tradeSize,
			Timestamp:   time.Now(),
		})
		order.FilledSize += tradeSize
		bestAskOrder.FilledSize += tradeSize
		order.RemainingSize -= tradeSize
		bestAskOrder.RemainingSize -= tradeSize
		if sellQtyLeft == tradeSize {
			bestAskOrder.Status = domain.OrderStatusFilled
			queue.RemoveOrder(bestAskNode)
			e.book.RemoveOrderByID(bestAskOrder.ID)
			if queue.IsEmpty() {
				e.book.RemoveAskPrice(bestAskPrice)
			}
		}
	}

	if order.FilledSize == order.Size {
		order.Status = domain.OrderStatusFilled
	} else if order.FilledSize > 0 {
		order.Status = domain.OrderStatusPartiallyFilled
	} else {
		order.Status = domain.OrderStatusCancelled
	}
	return trades, nil
}

func (e *Engine) processSellMarketOrder(order *domain.Order) ([]*domain.Trade, error) {
	var trades []*domain.Trade

	for {
		sellQtyLeft := order.RemainingSize
		if sellQtyLeft <= 0 {
			break
		}
		bestBidPrice, ok := e.book.BestBidPrice()
		if !ok {
			break
		}
		queue := e.book.BuyOrders[bestBidPrice]
		if queue.IsEmpty() {
			e.book.RemoveBidPrice(bestBidPrice)
			continue
		}
		bestBidNode := queue.GetHead()
		bestBidOrder := bestBidNode.Order
		buyQtyLeft := bestBidOrder.RemainingSize
		tradeSize := min(sellQtyLeft, buyQtyLeft)
		if tradeSize == 0 {
			break
		}

		trades = append(trades, &domain.Trade{
			ID:          uuid.New().String(),
			Market:      e.book.Market,
			BuyOrderID:  bestBidOrder.ID,
			SellOrderID: order.ID,
			Price:       bestBidPrice,
			Size:        tradeSize,
			Timestamp:   time.Now(),
		})
		order.FilledSize += tradeSize
		order.RemainingSize -= tradeSize
		bestBidOrder.RemainingSize -= tradeSize
		bestBidOrder.FilledSize += tradeSize
		if buyQtyLeft == tradeSize {
			bestBidOrder.Status = domain.OrderStatusFilled
			queue.RemoveOrder(bestBidNode)
			e.book.RemoveOrderByID(bestBidOrder.ID)
			if queue.IsEmpty() {
				e.book.RemoveBidPrice(bestBidPrice)
			}
		}
	}

	if order.FilledSize == order.Size {
		order.Status = domain.OrderStatusFilled
	} else if order.FilledSize > 0 {
		order.Status = domain.OrderStatusPartiallyFilled
	} else {
		order.Status = domain.OrderStatusCancelled
	}
	return trades, nil
}

func (e *Engine) processBuyOrder(order *domain.Order) []*domain.Trade {
	var trades []*domain.Trade

	for {
		buyQtyLeft := order.RemainingSize
		if buyQtyLeft <= 0 {
			break
		}
		bestAskPrice, ok := e.book.BestAskPrice()
		if !ok {
			break
		}
		if order.Price < bestAskPrice {
			break
		}
		queue := e.book.SellOrders[bestAskPrice]
		if queue.IsEmpty() {
			e.book.RemoveAskPrice(bestAskPrice)
			continue
		}
		bestAskNode := queue.GetHead()
		bestAskOrder := bestAskNode.Order
		sellQtyLeft := bestAskOrder.RemainingSize
		tradeSize := min(buyQtyLeft, sellQtyLeft)
		if tradeSize == 0 {
			break
		}

		trades = append(trades, &domain.Trade{
			ID:          uuid.New().String(),
			Market:      e.book.Market,
			BuyOrderID:  order.ID,
			SellOrderID: bestAskOrder.ID,
			Price:       bestAskPrice,
			Size:        tradeSize,
			Timestamp:   time.Now(),
		})
		order.FilledSize += tradeSize
		bestAskOrder.FilledSize += tradeSize
		order.RemainingSize -= tradeSize
		bestAskOrder.RemainingSize -= tradeSize
		if sellQtyLeft == tradeSize {
			bestAskOrder.Status = domain.OrderStatusFilled
			queue.RemoveOrder(bestAskNode)
			e.book.RemoveOrderByID(bestAskOrder.ID)
			if queue.IsEmpty() {
				e.book.RemoveAskPrice(bestAskPrice)
			}
		}
	}

	if order.FilledSize == order.Size {
		order.Status = domain.OrderStatusFilled
	} else {
		if order.Type == domain.OrderTypeIOC {
			order.Status = domain.OrderStatusCancelled
		} else {
			if order.FilledSize > 0 {
				order.Status = domain.OrderStatusPartiallyFilled
			} else {
				order.Status = domain.OrderStatusOpen
			}
			e.book.AddOrder(order)
		}
	}

	return trades
}

func (e *Engine) processSellOrder(order *domain.Order) []*domain.Trade {
	var trades []*domain.Trade

	for {
		sellQtyLeft := order.RemainingSize
		if sellQtyLeft <= 0 {
			break
		}
		bestBidPrice, ok := e.book.BestBidPrice()
		if !ok {
			break
		}
		if order.Price > bestBidPrice {
			break
		}
		queue := e.book.BuyOrders[bestBidPrice]
		if queue.IsEmpty() {
			e.book.RemoveBidPrice(bestBidPrice)
			continue
		}
		bestBidNode := queue.GetHead()
		bestBidOrder := bestBidNode.Order
		buyQtyLeft := bestBidOrder.RemainingSize
		tradeSize := min(sellQtyLeft, buyQtyLeft)
		if tradeSize == 0 {
			break
		}

		trades = append(trades, &domain.Trade{
			ID:          uuid.New().String(),
			Market:      e.book.Market,
			BuyOrderID:  bestBidOrder.ID,
			SellOrderID: order.ID,
			Price:       bestBidPrice,
			Size:        tradeSize,
			Timestamp:   time.Now(),
		})
		order.FilledSize += tradeSize
		bestBidOrder.FilledSize += tradeSize
		order.RemainingSize -= tradeSize
		bestBidOrder.RemainingSize -= tradeSize
		if buyQtyLeft == tradeSize {
			bestBidOrder.Status = domain.OrderStatusFilled
			queue.RemoveOrder(bestBidNode)
			e.book.RemoveOrderByID(bestBidOrder.ID)
			if queue.IsEmpty() {
				e.book.RemoveBidPrice(bestBidPrice)
			}
		}
	}

	if order.FilledSize == order.Size {
		order.Status = domain.OrderStatusFilled
	} else {
		if order.Type == domain.OrderTypeIOC {
			order.Status = domain.OrderStatusCancelled
		} else {
			if order.FilledSize > 0 {
				order.Status = domain.OrderStatusPartiallyFilled
			} else {
				order.Status = domain.OrderStatusOpen
			}
			e.book.AddOrder(order)
		}
	}

	return trades
}

func (e *Engine) CancelOrder(orderId string) error {
	existing := e.book.GetOrder(orderId)
	if existing == nil {
		return errors.New("order not found")
	}
	if existing.Status == domain.OrderStatusFilled {
		return errors.New("cannot cancel a filled order")
	}
	e.book.DeleteOrder(orderId)
	return nil
}

// #genai GetBookSnapshot returns an immutable view of the current book state.
func (e *Engine) GetBookSnapshot() BookSnapshot {
	snap := BookSnapshot{Market: e.book.Market}
	for price, level := range e.book.BuyOrders {
		var qty int64
		node := level.GetHead()
		for node != nil {
			qty += node.Order.RemainingSize
			node = node.Next
		}
		if qty > 0 {
			snap.Bids = append(snap.Bids, PriceQuantity{Price: price, Quantity: qty})
		}
	}
	for price, level := range e.book.SellOrders {
		var qty int64
		node := level.GetHead()
		for node != nil {
			qty += node.Order.RemainingSize
			node = node.Next
		}
		if qty > 0 {
			snap.Asks = append(snap.Asks, PriceQuantity{Price: price, Quantity: qty})
		}
	}
	return snap
}
