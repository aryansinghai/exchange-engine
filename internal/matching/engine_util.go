package matching

import (
	"time"
	"trade-exchange-engine/internal/domain"
)

// #genai
func isValidOrder(order *domain.Order) bool {
	return order.Validate() == nil
}

func initializeOrder(order *domain.Order) {
	order.Status = domain.OrderStatusOpen
	order.RemainingSize = order.Size
	order.FilledSize = 0
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
}
