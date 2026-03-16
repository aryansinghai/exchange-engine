package domain

type PriceLevel struct {
	Price int64

	Head *Order
	Tail *Order
}

func NewPriceLevel(price int64) *PriceLevel {
	return &PriceLevel{
		Price: price,
		Head:  nil,
		Tail:  nil,
	}
}

func (p *PriceLevel) AddOrder(order *Order) {
	order.Prev = p.Tail
	order.Next = nil
	order.PriceLevel = p

	if p.Tail != nil {
		p.Tail.Next = order
	} else {
		p.Head = order
	}

	p.Tail = order
}

func (p *PriceLevel) RemoveOrder(order *Order) {
	if order.Prev != nil {
		order.Prev.Next = order.Next
	} else {
		p.Head = order.Next
	}

	if order.Next != nil {
		order.Next.Prev = order.Prev
	} else {
		p.Tail = order.Prev
	}

	order.Prev = nil
	order.Next = nil
	order.PriceLevel = nil
}

func (p *PriceLevel) GetHead() *Order {
	return p.Head
}

func (p *PriceLevel) GetTail() *Order {
	return p.Tail
}

func (p *PriceLevel) PopHead() *Order {
	order := p.Head
	if order == nil {
		return nil
	}
	p.RemoveOrder(order)
	return order
}

func (p *PriceLevel) IsEmpty() bool {
	return p.Head == nil
}
