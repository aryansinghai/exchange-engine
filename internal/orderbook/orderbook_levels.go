package orderbook

import "slices"

// #genai Price book helpers for top-level bid/ask arrays.
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

func (o *OrderBook) BestBidPrice() (int64, bool) {
	if len(o.bidPrices) == 0 {
		return 0, false
	}
	return o.bidPrices[0], true
}

func (o *OrderBook) BestAskPrice() (int64, bool) {
	if len(o.askPrices) == 0 {
		return 0, false
	}
	return o.askPrices[0], true
}

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
