package orderbook

import "trade-exchange-engine/internal/domain"

// #genai OrderNode is the in-memory orderbook node wrapper.
type OrderNode struct {
	Order *domain.Order
	Prev  *OrderNode
	Next  *OrderNode
	Level *PriceLevel
}
