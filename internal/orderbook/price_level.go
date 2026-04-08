package orderbook

import "trade-exchange-engine/internal/domain"

// #genai PriceLevel is a single price bucket in the in-memory order book.
type PriceLevel struct {
	Price int64

	Head *OrderNode
	Tail *OrderNode
}

func NewPriceLevel(price int64) *PriceLevel {
	return &PriceLevel{
		Price: price,
		Head:  nil,
		Tail:  nil,
	}
}

func (p *PriceLevel) AddOrder(node *OrderNode) {
	node.Prev = p.Tail
	node.Next = nil
	node.Level = p

	if p.Tail != nil {
		p.Tail.Next = node
	} else {
		p.Head = node
	}

	p.Tail = node
}

func (p *PriceLevel) RemoveOrder(node *OrderNode) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		p.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		p.Tail = node.Prev
	}

	node.Prev = nil
	node.Next = nil
	node.Level = nil
}

func (p *PriceLevel) GetHead() *OrderNode {
	return p.Head
}

func (p *PriceLevel) GetTail() *OrderNode {
	return p.Tail
}

func (p *PriceLevel) PopHead() *OrderNode {
	node := p.Head
	if node == nil {
		return nil
	}
	p.RemoveOrder(node)
	return node
}

func (p *PriceLevel) IsEmpty() bool {
	return p.Head == nil
}

var _ = domain.Order{}
