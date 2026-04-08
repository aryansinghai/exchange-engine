package matching

import (
	"testing"
	"trade-exchange-engine/internal/domain"
	"trade-exchange-engine/internal/orderbook"
)

// #genai

func newLimitOrder(id string, side domain.OrderSide, price, size int64) *domain.Order {
	return &domain.Order{
		ID:            id,
		Market:        "BTC/USD",
		Type:          domain.OrderTypeLimit,
		Side:          side,
		Size:          size,
		Price:         price,
		RemainingSize: size,
	}
}

func newMarketOrder(id string, side domain.OrderSide, size int64) *domain.Order {
	return &domain.Order{
		ID:            id,
		Market:        "BTC/USD",
		Type:          domain.OrderTypeMarket,
		Side:          side,
		Size:          size,
		RemainingSize: size,
	}
}

func newIOCOrder(id string, side domain.OrderSide, price, size int64) *domain.Order {
	return &domain.Order{
		ID:            id,
		Market:        "BTC/USD",
		Type:          domain.OrderTypeIOC,
		Side:          side,
		Size:          size,
		Price:         price,
		RemainingSize: size,
	}
}

func setup() (*orderbook.OrderBook, *Engine) {
	book := orderbook.NewOrderBook("BTC/USD")
	engine := NewEngine(book)
	return book, engine
}

// --- Invalid order type ---

func TestSubmitOrder_InvalidType(t *testing.T) {
	_, engine := setup()
	order := &domain.Order{ID: "x1", Type: "STOP_LOSS", Size: 1, RemainingSize: 1, Side: domain.OrderSideBuy}
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for invalid order type")
	}
}

// --- Limit buy into empty book ---

func TestLimitBuy_EmptyBook_RestsOnBook(t *testing.T) {
	book, engine := setup()
	order := newLimitOrder("b1", domain.OrderSideBuy, 100, 10)
	trades, err := engine.SubmitOrder(order)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 0 {
		t.Fatal("no trades expected against empty book")
	}
	if order.Status != domain.OrderStatusOpen {
		t.Fatalf("expected OPEN, got %s", order.Status)
	}
	best, ok := book.BestBidPrice()
	if !ok || best != 100 {
		t.Fatal("order should rest on bid side")
	}
}

// --- Limit sell into empty book ---

func TestLimitSell_EmptyBook_RestsOnBook(t *testing.T) {
	book, engine := setup()
	order := newLimitOrder("s1", domain.OrderSideSell, 200, 10)
	trades, err := engine.SubmitOrder(order)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 0 {
		t.Fatal("no trades expected against empty book")
	}
	if order.Status != domain.OrderStatusOpen {
		t.Fatalf("expected OPEN, got %s", order.Status)
	}
	best, ok := book.BestAskPrice()
	if !ok || best != 200 {
		t.Fatal("order should rest on ask side")
	}
}

// --- Exact match limit orders ---

func TestLimitBuy_ExactMatch(t *testing.T) {
	book, engine := setup()
	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 5)
	engine.SubmitOrder(sell)

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	trades, err := engine.SubmitOrder(buy)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Size != 5 {
		t.Fatalf("expected trade size 5, got %d", trades[0].Size)
	}
	if trades[0].Price != 100 {
		t.Fatalf("expected trade price 100, got %d", trades[0].Price)
	}
	if buy.Status != domain.OrderStatusFilled {
		t.Fatalf("buy should be FILLED, got %s", buy.Status)
	}
	if sell.Status != domain.OrderStatusFilled {
		t.Fatalf("sell should be FILLED, got %s", sell.Status)
	}
	_, hasBid := book.BestBidPrice()
	_, hasAsk := book.BestAskPrice()
	if hasBid || hasAsk {
		t.Fatal("book should be empty after full match")
	}
}

func TestLimitSell_ExactMatch(t *testing.T) {
	_, engine := setup()
	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	engine.SubmitOrder(buy)

	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 5)
	trades, _ := engine.SubmitOrder(sell)

	if len(trades) != 1 || trades[0].Size != 5 {
		t.Fatal("expected exact match trade of size 5")
	}
	if buy.Status != domain.OrderStatusFilled || sell.Status != domain.OrderStatusFilled {
		t.Fatal("both orders should be FILLED")
	}
}

// --- Partial fills ---

func TestLimitBuy_PartialFill_BuyLarger(t *testing.T) {
	book, engine := setup()
	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 3)
	engine.SubmitOrder(sell)

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 10)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 1 || trades[0].Size != 3 {
		t.Fatalf("expected trade size 3, got %d trades", len(trades))
	}
	if buy.Status != domain.OrderStatusPartiallyFilled {
		t.Fatalf("buy should be PARTIALLY_FILLED, got %s", buy.Status)
	}
	if buy.FilledSize != 3 || buy.RemainingSize != 7 {
		t.Fatalf("buy filled=%d remaining=%d", buy.FilledSize, buy.RemainingSize)
	}
	best, ok := book.BestBidPrice()
	if !ok || best != 100 {
		t.Fatal("remaining buy should rest on book")
	}
}

func TestLimitBuy_PartialFill_SellLarger(t *testing.T) {
	book, engine := setup()
	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 10)
	engine.SubmitOrder(sell)

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 3)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 1 || trades[0].Size != 3 {
		t.Fatal("expected trade size 3")
	}
	if buy.Status != domain.OrderStatusFilled {
		t.Fatalf("buy should be FILLED, got %s", buy.Status)
	}
	if sell.RemainingSize != 7 {
		t.Fatalf("sell remaining should be 7, got %d", sell.RemainingSize)
	}
	_, hasBid := book.BestBidPrice()
	if hasBid {
		t.Fatal("fully filled buy should not rest on book")
	}
	best, ok := book.BestAskPrice()
	if !ok || best != 100 {
		t.Fatal("partially filled sell should remain")
	}
}

// --- Price priority ---

func TestLimitBuy_PricePriority(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))
	engine.SubmitOrder(newLimitOrder("s2", domain.OrderSideSell, 90, 5))

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Price != 90 {
		t.Fatalf("should match best ask (90), got %d", trades[0].Price)
	}
	if trades[0].SellOrderID != "s2" {
		t.Fatalf("should match s2 first, got %s", trades[0].SellOrderID)
	}
}

func TestLimitSell_PricePriority(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("b1", domain.OrderSideBuy, 100, 5))
	engine.SubmitOrder(newLimitOrder("b2", domain.OrderSideBuy, 110, 5))

	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 5)
	trades, _ := engine.SubmitOrder(sell)

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Price != 110 {
		t.Fatalf("should match best bid (110), got %d", trades[0].Price)
	}
	if trades[0].BuyOrderID != "b2" {
		t.Fatalf("should match b2 first, got %s", trades[0].BuyOrderID)
	}
}

// --- Time priority (FIFO at same price) ---

func TestLimitBuy_TimePriority_FIFO(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))
	engine.SubmitOrder(newLimitOrder("s2", domain.OrderSideSell, 100, 5))

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	trades, _ := engine.SubmitOrder(buy)

	if trades[0].SellOrderID != "s1" {
		t.Fatalf("FIFO: should match s1 first, got %s", trades[0].SellOrderID)
	}
}

// --- No match when price doesn't cross ---

func TestLimitBuy_NoMatchWhenPriceBelowAsk(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 200, 5))

	buy := newLimitOrder("b1", domain.OrderSideBuy, 150, 5)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 0 {
		t.Fatal("should not match when buy price < ask price")
	}
	if buy.Status != domain.OrderStatusOpen {
		t.Fatalf("expected OPEN, got %s", buy.Status)
	}
	best, _ := book.BestBidPrice()
	if best != 150 {
		t.Fatal("unmatched buy should rest at 150")
	}
}

func TestLimitSell_NoMatchWhenPriceAboveBid(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("b1", domain.OrderSideBuy, 100, 5))

	sell := newLimitOrder("s1", domain.OrderSideSell, 150, 5)
	trades, _ := engine.SubmitOrder(sell)

	if len(trades) != 0 {
		t.Fatal("should not match when sell price > bid price")
	}
	if sell.Status != domain.OrderStatusOpen {
		t.Fatalf("expected OPEN, got %s", sell.Status)
	}
	best, _ := book.BestAskPrice()
	if best != 150 {
		t.Fatal("unmatched sell should rest at 150")
	}
}

// --- Multi-level matching ---

func TestLimitBuy_MatchesAcrossMultiplePriceLevels(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 3))
	engine.SubmitOrder(newLimitOrder("s2", domain.OrderSideSell, 110, 3))

	buy := newLimitOrder("b1", domain.OrderSideBuy, 110, 5)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}
	if trades[0].Price != 100 || trades[0].Size != 3 {
		t.Fatalf("first trade should be at 100 for 3, got price=%d size=%d", trades[0].Price, trades[0].Size)
	}
	if trades[1].Price != 110 || trades[1].Size != 2 {
		t.Fatalf("second trade should be at 110 for 2, got price=%d size=%d", trades[1].Price, trades[1].Size)
	}
	if buy.Status != domain.OrderStatusFilled {
		t.Fatalf("expected FILLED, got %s", buy.Status)
	}
}

// --- Market orders ---

func TestMarketBuy_FullFill(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 10))

	buy := newMarketOrder("b1", domain.OrderSideBuy, 10)
	trades, err := engine.SubmitOrder(buy)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 || trades[0].Size != 10 {
		t.Fatal("market buy should fully fill")
	}
	if buy.Status != domain.OrderStatusFilled {
		t.Fatalf("expected FILLED, got %s", buy.Status)
	}
	_, hasAsk := book.BestAskPrice()
	if hasAsk {
		t.Fatal("sell should be consumed")
	}
}

func TestMarketSell_FullFill(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("b1", domain.OrderSideBuy, 100, 10))

	sell := newMarketOrder("s1", domain.OrderSideSell, 10)
	trades, _ := engine.SubmitOrder(sell)

	if len(trades) != 1 || trades[0].Size != 10 {
		t.Fatal("market sell should fully fill")
	}
	if sell.Status != domain.OrderStatusFilled {
		t.Fatalf("expected FILLED, got %s", sell.Status)
	}
	_, hasBid := book.BestBidPrice()
	if hasBid {
		t.Fatal("buy should be consumed")
	}
}

func TestMarketBuy_PartialFill_InsufficientLiquidity(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 3))

	buy := newMarketOrder("b1", domain.OrderSideBuy, 10)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 1 || trades[0].Size != 3 {
		t.Fatal("should fill 3 of 10")
	}
	if buy.Status != domain.OrderStatusPartiallyFilled {
		t.Fatalf("expected PARTIALLY_FILLED, got %s", buy.Status)
	}
	if buy.FilledSize != 3 || buy.RemainingSize != 7 {
		t.Fatal("filled/remaining mismatch")
	}
}

func TestMarketBuy_NoLiquidity_Cancelled(t *testing.T) {
	_, engine := setup()

	buy := newMarketOrder("b1", domain.OrderSideBuy, 10)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 0 {
		t.Fatal("no trades on empty book")
	}
	if buy.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", buy.Status)
	}
}

func TestMarketSell_NoLiquidity_Cancelled(t *testing.T) {
	_, engine := setup()

	sell := newMarketOrder("s1", domain.OrderSideSell, 10)
	trades, _ := engine.SubmitOrder(sell)

	if len(trades) != 0 {
		t.Fatal("no trades on empty book")
	}
	if sell.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", sell.Status)
	}
}

func TestMarketBuy_MatchesAcrossMultipleLevels(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))
	engine.SubmitOrder(newLimitOrder("s2", domain.OrderSideSell, 110, 5))

	buy := newMarketOrder("b1", domain.OrderSideBuy, 8)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}
	if trades[0].Price != 100 || trades[0].Size != 5 {
		t.Fatal("first trade at best ask 100")
	}
	if trades[1].Price != 110 || trades[1].Size != 3 {
		t.Fatal("second trade at next ask 110")
	}
}

// --- IOC orders ---

func TestIOCBuy_FullMatch(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))

	ioc := newIOCOrder("b1", domain.OrderSideBuy, 100, 5)
	trades, _ := engine.SubmitOrder(ioc)

	if len(trades) != 1 || trades[0].Size != 5 {
		t.Fatal("IOC should fully match")
	}
	if ioc.Status != domain.OrderStatusFilled {
		t.Fatalf("expected FILLED, got %s", ioc.Status)
	}
}

func TestIOCBuy_PartialMatch_CancelledRemainder(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 3))

	ioc := newIOCOrder("b1", domain.OrderSideBuy, 100, 10)
	trades, _ := engine.SubmitOrder(ioc)

	if len(trades) != 1 || trades[0].Size != 3 {
		t.Fatal("should fill 3")
	}
	if ioc.Status != domain.OrderStatusCancelled {
		t.Fatalf("IOC remainder should be CANCELLED, got %s", ioc.Status)
	}
	_, hasBid := book.BestBidPrice()
	if hasBid {
		t.Fatal("IOC should NOT rest on book")
	}
}

func TestIOCBuy_NoMatch_Cancelled(t *testing.T) {
	book, engine := setup()

	ioc := newIOCOrder("b1", domain.OrderSideBuy, 50, 10)
	trades, _ := engine.SubmitOrder(ioc)

	if len(trades) != 0 {
		t.Fatal("no match expected")
	}
	if ioc.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", ioc.Status)
	}
	_, hasBid := book.BestBidPrice()
	if hasBid {
		t.Fatal("IOC should NOT rest on book")
	}
}

func TestIOCSell_PartialMatch_CancelledRemainder(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("b1", domain.OrderSideBuy, 100, 3))

	ioc := newIOCOrder("s1", domain.OrderSideSell, 100, 10)
	trades, _ := engine.SubmitOrder(ioc)

	if len(trades) != 1 || trades[0].Size != 3 {
		t.Fatal("should fill 3")
	}
	if ioc.Status != domain.OrderStatusCancelled {
		t.Fatalf("IOC sell remainder should be CANCELLED, got %s", ioc.Status)
	}
	_, hasAsk := book.BestAskPrice()
	if hasAsk {
		t.Fatal("IOC sell should NOT rest on book")
	}
}

func TestIOCSell_NoMatch_Cancelled(t *testing.T) {
	book, engine := setup()

	ioc := newIOCOrder("s1", domain.OrderSideSell, 500, 10)
	trades, _ := engine.SubmitOrder(ioc)

	if len(trades) != 0 {
		t.Fatal("no match expected")
	}
	if ioc.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", ioc.Status)
	}
	_, hasAsk := book.BestAskPrice()
	if hasAsk {
		t.Fatal("IOC should NOT rest on book")
	}
}

// --- Trade fields ---

func TestTrade_FieldsCorrect(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))
	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	trades, _ := engine.SubmitOrder(buy)

	trade := trades[0]
	if trade.Market != "BTC/USD" {
		t.Fatalf("expected market BTC/USD, got %s", trade.Market)
	}
	if trade.BuyOrderID != "b1" {
		t.Fatalf("expected buy order b1, got %s", trade.BuyOrderID)
	}
	if trade.SellOrderID != "s1" {
		t.Fatalf("expected sell order s1, got %s", trade.SellOrderID)
	}
	if trade.ID == "" {
		t.Fatal("trade ID should not be empty")
	}
	if trade.Timestamp.IsZero() {
		t.Fatal("trade timestamp should be set")
	}
}

// --- Order book state after cancel ---

func TestDeleteOrder_ThenMatchSkipsIt(t *testing.T) {
	book, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))
	engine.SubmitOrder(newLimitOrder("s2", domain.OrderSideSell, 100, 5))

	book.DeleteOrder("s1")

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].SellOrderID != "s2" {
		t.Fatalf("should match s2 (s1 was cancelled), got %s", trades[0].SellOrderID)
	}
}

// --- Aggressive limit order (buy price > ask) ---

func TestLimitBuy_AggressivePrice_MatchesAtAskPrice(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 5))

	buy := newLimitOrder("b1", domain.OrderSideBuy, 150, 5)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 1 {
		t.Fatal("aggressive buy should match")
	}
	if trades[0].Price != 100 {
		t.Fatalf("trade should execute at resting ask price 100, got %d", trades[0].Price)
	}
}

func TestLimitSell_AggressivePrice_MatchesAtBidPrice(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("b1", domain.OrderSideBuy, 200, 5))

	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 5)
	trades, _ := engine.SubmitOrder(sell)

	if len(trades) != 1 {
		t.Fatal("aggressive sell should match")
	}
	if trades[0].Price != 200 {
		t.Fatalf("trade should execute at resting bid price 200, got %d", trades[0].Price)
	}
}

// --- Multiple fills from single order ---

func TestLimitBuy_FillsMultipleOrdersAtSameLevel(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 100, 3))
	engine.SubmitOrder(newLimitOrder("s2", domain.OrderSideSell, 100, 4))

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 7)
	trades, _ := engine.SubmitOrder(buy)

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}
	if trades[0].Size != 3 || trades[0].SellOrderID != "s1" {
		t.Fatal("first fill should be s1 for 3")
	}
	if trades[1].Size != 4 || trades[1].SellOrderID != "s2" {
		t.Fatal("second fill should be s2 for 4")
	}
	if buy.Status != domain.OrderStatusFilled {
		t.Fatalf("expected FILLED, got %s", buy.Status)
	}
}

// --- Market order does NOT rest on book ---

func TestMarketBuy_DoesNotRestOnBook(t *testing.T) {
	book, engine := setup()
	buy := newMarketOrder("b1", domain.OrderSideBuy, 10)
	engine.SubmitOrder(buy)

	_, hasBid := book.BestBidPrice()
	if hasBid {
		t.Fatal("unfilled market order should NOT rest on book")
	}
}

func TestMarketSell_DoesNotRestOnBook(t *testing.T) {
	book, engine := setup()
	sell := newMarketOrder("s1", domain.OrderSideSell, 10)
	engine.SubmitOrder(sell)

	_, hasAsk := book.BestAskPrice()
	if hasAsk {
		t.Fatal("unfilled market order should NOT rest on book")
	}
}

// --- Validation tests ---

func TestSubmitOrder_ZeroSize_Rejected(t *testing.T) {
	_, engine := setup()
	order := newLimitOrder("v1", domain.OrderSideBuy, 100, 0)
	order.Size = 0
	order.RemainingSize = 0
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for zero size")
	}
}

func TestSubmitOrder_NegativeSize_Rejected(t *testing.T) {
	_, engine := setup()
	order := &domain.Order{
		ID:   "v2",
		Type: domain.OrderTypeLimit,
		Side: domain.OrderSideBuy,
		Size: -5, Price: 100, RemainingSize: -5,
		Market: "BTC/USD",
	}
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for negative size")
	}
}

func TestSubmitOrder_LimitZeroPrice_Rejected(t *testing.T) {
	_, engine := setup()
	order := &domain.Order{
		ID:   "v3",
		Type: domain.OrderTypeLimit,
		Side: domain.OrderSideBuy,
		Size: 5, Price: 0, RemainingSize: 5,
		Market: "BTC/USD",
	}
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for limit order with zero price")
	}
}

func TestSubmitOrder_LimitNegativePrice_Rejected(t *testing.T) {
	_, engine := setup()
	order := &domain.Order{
		ID:   "v4",
		Type: domain.OrderTypeLimit,
		Side: domain.OrderSideBuy,
		Size: 5, Price: -100, RemainingSize: 5,
		Market: "BTC/USD",
	}
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for limit order with negative price")
	}
}

func TestSubmitOrder_EmptyID_Rejected(t *testing.T) {
	_, engine := setup()
	order := &domain.Order{
		ID:   "",
		Type: domain.OrderTypeLimit,
		Side: domain.OrderSideBuy,
		Size: 5, Price: 100, RemainingSize: 5,
		Market: "BTC/USD",
	}
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for empty order ID")
	}
}

func TestSubmitOrder_InvalidSide_Rejected(t *testing.T) {
	_, engine := setup()
	order := &domain.Order{
		ID:   "v5",
		Type: domain.OrderTypeLimit,
		Side: "SHORT",
		Size: 5, Price: 100, RemainingSize: 5,
		Market: "BTC/USD",
	}
	_, err := engine.SubmitOrder(order)
	if err == nil {
		t.Fatal("expected error for invalid order side")
	}
}

// --- Cancel edge case tests ---

func TestCancelOrder_OpenOrder(t *testing.T) {
	_, engine := setup()
	order := newLimitOrder("c1", domain.OrderSideBuy, 100, 10)
	engine.SubmitOrder(order)

	err := engine.CancelOrder("c1")
	if err != nil {
		t.Fatalf("expected cancel to succeed, got %v", err)
	}
	if order.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", order.Status)
	}
}

func TestCancelOrder_PartiallyFilledOrder(t *testing.T) {
	_, engine := setup()
	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 3)
	engine.SubmitOrder(sell)

	buy := newLimitOrder("c2", domain.OrderSideBuy, 100, 10)
	engine.SubmitOrder(buy)

	if buy.Status != domain.OrderStatusPartiallyFilled {
		t.Fatalf("precondition: expected PARTIALLY_FILLED, got %s", buy.Status)
	}

	err := engine.CancelOrder("c2")
	if err != nil {
		t.Fatalf("expected cancel of partially filled order to succeed, got %v", err)
	}
	if buy.Status != domain.OrderStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", buy.Status)
	}
}

func TestCancelOrder_UnknownOrder_ReturnsError(t *testing.T) {
	_, engine := setup()
	err := engine.CancelOrder("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown order")
	}
}

func TestCancelOrder_FilledOrder_ReturnsError(t *testing.T) {
	_, engine := setup()
	sell := newLimitOrder("s1", domain.OrderSideSell, 100, 5)
	engine.SubmitOrder(sell)

	buy := newLimitOrder("b1", domain.OrderSideBuy, 100, 5)
	engine.SubmitOrder(buy)

	if sell.Status != domain.OrderStatusFilled {
		t.Fatalf("precondition: expected sell to be FILLED, got %s", sell.Status)
	}

	err := engine.CancelOrder("s1")
	if err == nil {
		t.Fatal("expected error when cancelling a filled order")
	}
}

// --- BookSnapshot tests ---

func TestGetBookSnapshot_EmptyBook(t *testing.T) {
	_, engine := setup()
	snap := engine.GetBookSnapshot()

	if snap.Market != "BTC/USD" {
		t.Fatalf("expected market BTC/USD, got %s", snap.Market)
	}
	if len(snap.Bids) != 0 {
		t.Fatal("expected no bids on empty book")
	}
	if len(snap.Asks) != 0 {
		t.Fatal("expected no asks on empty book")
	}
}

func TestGetBookSnapshot_WithOrders(t *testing.T) {
	_, engine := setup()
	engine.SubmitOrder(newLimitOrder("b1", domain.OrderSideBuy, 100, 5))
	engine.SubmitOrder(newLimitOrder("b2", domain.OrderSideBuy, 100, 3))
	engine.SubmitOrder(newLimitOrder("s1", domain.OrderSideSell, 200, 10))

	snap := engine.GetBookSnapshot()

	if len(snap.Bids) != 1 {
		t.Fatalf("expected 1 bid level, got %d", len(snap.Bids))
	}
	if snap.Bids[0].Price != 100 || snap.Bids[0].Quantity != 8 {
		t.Fatalf("expected bid {100, 8}, got {%d, %d}", snap.Bids[0].Price, snap.Bids[0].Quantity)
	}
	if len(snap.Asks) != 1 {
		t.Fatalf("expected 1 ask level, got %d", len(snap.Asks))
	}
	if snap.Asks[0].Price != 200 || snap.Asks[0].Quantity != 10 {
		t.Fatalf("expected ask {200, 10}, got {%d, %d}", snap.Asks[0].Price, snap.Asks[0].Quantity)
	}
}
