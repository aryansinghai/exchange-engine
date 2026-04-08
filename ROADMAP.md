# `ROADMAP.md`

## Production Roadmap For `trade-exchange-engine`

### Summary
This project should become a **production-minded paper trading exchange** that demonstrates strong backend design, matching engine correctness, realtime data flow, and a polished retail web experience.

The goal is not to build a real-money exchange in one month. The goal is to build a **credible, interview-ready, production-style system** that shows:
- strong system design
- thoughtful product scoping
- clean architecture
- operational maturity
- a working end-to-end demo

Recommended v1 positioning:
- **Paper trading only**
- **Retail web users**
- **Single-region MVP**
- **Modular monolith architecture**
- **Go backend + React/Next.js frontend + PostgreSQL**

### Current State
What exists today:
- In-memory order book and matching logic in [`/Users/aryan.singhai/Desktop/personal projects/trade-exchange-engine/internal/orderbook/orderbook.go`](/Users/aryan.singhai/Desktop/personal%20projects/trade-exchange-engine/internal/orderbook/orderbook.go)
- Book internals (`OrderNode`, `PriceLevel`) in [`internal/orderbook`](/Users/aryan.singhai/Desktop/personal%20projects/trade-exchange-engine/internal/orderbook)
- Basic domain models in [`/Users/aryan.singhai/Desktop/personal projects/trade-exchange-engine/internal/domain/order.go`](/Users/aryan.singhai/Desktop/personal%20projects/trade-exchange-engine/internal/domain/order.go)
- Placeholder HTTP server in [`/Users/aryan.singhai/Desktop/personal projects/trade-exchange-engine/cmd/api/main.go`](/Users/aryan.singhai/Desktop/personal%20projects/trade-exchange-engine/cmd/api/main.go)
- Minimal tests in [`/Users/aryan.singhai/Desktop/personal projects/trade-exchange-engine/internal/orderbook/orderbook_test.go`](/Users/aryan.singhai/Desktop/personal%20projects/trade-exchange-engine/internal/orderbook/orderbook_test.go)

What is missing:
- real API surface
- persistence
- auth and user accounts
- balances/portfolio model
- realtime feeds
- frontend product
- observability and deployability
- robust test coverage

### Architecture Direction
Use a **modular monolith** with clear internal boundaries.

Core modules:
- `matching`: matching engine and market event loop
- `orders`: order submission, validation, cancellation, status tracking
- `markets`: market metadata and market-level orchestration
- `wallet` or `accounts`: paper balances, reservations, portfolio state
- `trades`: trade persistence and trade history
- `users`: auth and profile
- `server`: HTTP and WebSocket delivery
- `config` and `logger`: shared runtime infrastructure

Recommended runtime model:
- one in-memory engine loop per market
- HTTP for commands and queries
- WebSocket for live market data
- PostgreSQL for persistence
- internal event flow for order/trade/balance updates

### V1 Product Scope
The first version should let a reviewer:
1. register or sign in
2. receive seeded paper balances
3. open a market page like `BTC/USD`
4. place buy and sell orders
5. cancel open orders
6. see live order book and recent trades
7. see portfolio and order history
8. restart the app without losing persisted orders/trades

Out of scope for v1:
- real money movement
- external exchange connectivity
- KYC/compliance
- multi-region architecture
- microservices
- advanced charting and pro-trader tooling

### Roadmap
#### Phase 1: Matching Engine Foundation
- harden order lifecycle rules
- add strong engine tests
- isolate engine behind a service interface
- make matching deterministic and easier to reason about

#### Phase 2: API And Persistence
- replace the placeholder server with versioned JSON APIs
- persist users, balances, orders, and trades
- add order submission and cancellation endpoints
- support startup recovery

#### Phase 3: Retail Product Experience
- build frontend market view
- add portfolio and order history pages
- stream book/trade updates over WebSocket
- make the product demo-friendly and visually polished

#### Phase 4: Production Polish
- add logs, metrics, tracing, health checks
- add CI and deployment setup
- add benchmark and recovery stories
- document architecture and scaling path

### Public Interfaces To Add
Suggested v1 API surface:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/markets`
- `GET /api/v1/markets/:symbol/orderbook`
- `POST /api/v1/orders`
- `DELETE /api/v1/orders/:id`
- `GET /api/v1/orders`
- `GET /api/v1/trades`
- `GET /api/v1/portfolio`
- `GET /ws/markets/:symbol`

Suggested internal service interfaces:
- `SubmitOrder`
- `CancelOrder`
- `GetBookSnapshot`
- `GetUserOrders`
- `GetTrades`
- `GetPortfolio`

### Next Step You Should Work On
## Feature: Engine Correctness And Order Lifecycle Foundation

This is the highest-leverage next step.

Why this comes first:
- every later feature depends on correct matching behavior
- a polished UI cannot compensate for incorrect fills or broken order states
- this gives you the strongest engineering signal early
- your repo already has the beginnings of this in `internal/orderbook`, while the API and frontend are still mostly empty

Goal:
Build a clean, well-tested foundation for submitting, matching, and cancelling orders so the engine can be safely exposed through an API in the next step.

What to improve next:
- ensure `RemainingSize` is initialized and always consistent with `Size - FilledSize`
- define clear validation rules for market, side, type, size, and price
- make status transitions explicit: `OPEN`, `PARTIALLY_FILLED`, `FILLED`, `CANCELLED`
- make IOC behavior explicit and testable
- verify cancel behavior for open and partially filled orders
- ensure no invalid order can be inserted into the book
- keep deterministic matching based on price-time priority
- remove ambiguity between "book storage" and "application service" responsibilities

Suggested implementation shape:
- keep `internal/orderbook` focused on book mechanics
- use `internal/matching` as the application/service layer for commands like submit and cancel
- use `internal/domain` for strict data rules and state transitions
- keep `cmd/api` and `internal/server` out of engine internals

Suggested responsibilities:
- `domain`: order fields, enums, invariants, helper constructors/validators
- `orderbook`: price levels, queues, matching mechanics, delete/remove rules
- `matching`: command handlers, validation, orchestration, returned events/results

Deliverables for this step:
- a documented order lifecycle
- a small matching service interface
- comprehensive unit tests for matching scenarios
- a clear separation between engine core and future API layer
- confidence that later persistence/API work will sit on correct behavior

Acceptance criteria:
- limit buy and sell matching obey price-time priority
- market orders consume best available liquidity correctly
- IOC orders never rest on the book
- partially filled orders keep correct remaining quantity
- filled orders are removed from the book
- cancelled orders are removed and cannot be matched again
- invalid inputs are rejected before mutation
- test coverage includes both happy paths and edge cases

Recommended test scenarios:
- add resting buy then matching sell
- add resting sell then matching buy
- partial fill across one level
- market order across multiple price levels
- IOC partial execution with leftover cancelled
- cancellation of open order
- cancellation of partially filled order
- cancel unknown order
- empty book market order behavior
- order with zero or negative size
- limit order with invalid price
- multiple orders at same price verifying FIFO behavior

What not to do in this step:
- do not build auth yet
- do not build frontend yet
- do not introduce WebSockets yet
- do not over-optimize for extreme performance yet
- do not jump to microservices

Manager review checklist for this step:
- can I explain the order lifecycle in one minute?
- can I point to tests that prove the important matching rules?
- is the engine usable without HTTP concerns leaking in?
- if I add persistence next, do I know exactly where it belongs?
- if a reviewer reads this code, will they see intentional architecture rather than just algorithms?

### After That
Once this step is complete, the next feature should be:

## Feature: Minimal Trading API With Persistence
Build the thinnest useful API around the hardened engine:
- `POST /orders`
- `DELETE /orders/:id`
- `GET /markets/:symbol/orderbook`
- `GET /orders`
- PostgreSQL tables for orders and trades
- startup recovery for open orders

That will be the bridge from "good algorithm project" to "real production-style system."

### Test Plan
- unit tests for matching logic
- scenario tests for lifecycle transitions
- integration tests for API to engine consistency
- restart/recovery tests once persistence is added
- benchmark tests for basic throughput story

### Assumptions And Defaults
- project remains a paper trading platform
- primary audience is hiring managers and interviewers
- one-month horizon means disciplined scope matters more than completeness
- modular monolith is the correct architecture for v1
- the next step is correctness and service boundaries, not UI breadth
- recommended doc path when you leave Plan Mode: `/Users/aryan.singhai/Desktop/personal projects/trade-exchange-engine/ROADMAP.md`
