package main

import (
	"fmt"
	"math/rand"

	"sync"
	"time"
)

type Stock struct {
	Symbol         string
	CurrentPrice   float64
	TradingVolume  int
	OrderBookDepth int
	mu             sync.Mutex
}

type Portfolio struct {
	Cash   float64
	Stocks map[string]*Stock
	mu     sync.Mutex
}

func NewStock(symbol string, initialPrice float64, initialVolume int, orderBookDepth int) *Stock {
	return &Stock{
		Symbol:         symbol,
		CurrentPrice:   initialPrice,
		TradingVolume:  initialVolume,
		OrderBookDepth: orderBookDepth,
	}
}

func NewPortfolio(initialCash float64) *Portfolio {
	return &Portfolio{
		Cash:   initialCash,
		Stocks: make(map[string]*Stock),
	}
}

func (p *Portfolio) BuyStock(symbol string, price float64, quantity int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Cash < price*float64(quantity) {
		return fmt.Errorf("insufficient funds")
	}
	if p.Stocks[symbol] == nil {
		p.Stocks[symbol] = NewStock(symbol, price, 0, 0)
	}
	stock := p.Stocks[symbol]
	stock.mu.Lock()
	defer stock.mu.Unlock()

	stock.CurrentPrice = price
	stock.TradingVolume += quantity
	p.Cash -= price * float64(quantity)
	return nil
}

func (p *Portfolio) SellStock(symbol string, price float64, quantity int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	stock := p.Stocks[symbol]
	if stock == nil || stock.TradingVolume < quantity {
		return fmt.Errorf("insufficient stocks to sell")
	}

	stock.mu.Lock()
	defer stock.mu.Unlock()

	stock.CurrentPrice = price
	stock.TradingVolume -= quantity
	p.Cash += price * float64(quantity)

	if stock.TradingVolume == 0 {
		delete(p.Stocks, symbol)
	}
	return nil
}

func (s *Stock) SimulateMarket() {
	for {
		time.Sleep(time.Second)
		s.mu.Lock()
		s.CurrentPrice = generateRandomPriceUpdate(s.CurrentPrice)
		s.TradingVolume = generateRandomVolumeUpdate(s.TradingVolume)
		s.mu.Unlock()
		fmt.Printf("Stock %s - Price: %.2f, Volume: %d\n", s.Symbol, s.CurrentPrice, s.TradingVolume)

	}
}

func generateRandomPriceUpdate(currentPrice float64) float64 {
	priceChange := (rand.Float64() - 0.5) * 0.04 * currentPrice
	return currentPrice + priceChange
}

func generateRandomVolumeUpdate(currentVolume int) int {
	volumeChange := rand.Intn(currentVolume/10+1) - currentVolume
	return currentVolume + volumeChange
}

func main() {
	rand.Seed(time.Now().UnixNano())

	stocks := make(map[string]*Stock)

	symbols := []string{
		"YNDX", "TATN", "MGNT", "TCSG",
	}

	for _, symbol := range symbols {
		stocks[symbol] = NewStock(symbol, 100, 1000, 10)

		go stocks[symbol].SimulateMarket()
	}
	userPortfolios := make(map[string]*Portfolio)

	var wg sync.WaitGroup
	usernames := []string{
		"trader1", "trader2", "trader3",
	}
	for _, username := range usernames {
		wg.Add(1)
		go func(user string) {
			defer wg.Done()
			if userPortfolios[user] == nil {
				userPortfolios[user] = NewPortfolio(1000)
			}
			for i := 0; i < 5; i++ {
				symbol := symbols[rand.Intn(len(symbols))]
				price := stocks[symbol].CurrentPrice
				quantity := rand.Intn(10) + 1
				action := rand.Intn(2)
				if action == 0 {
					err := userPortfolios[user].BuyStock(symbol, price, quantity)
					if err == nil {
						fmt.Printf("%s bought %d shares of %s at %.2f\n", user, quantity, symbol, price)
					}
				} else {
					err := userPortfolios[user].SellStock(symbol, price, quantity)
					if err == nil {
						fmt.Printf("%s sell %d shares of %s at %.2f\n", user, quantity, symbol, price)
					}
				}
				time.Sleep(time.Second)
			}
			fmt.Printf("%s's Portfolio:\n", user)
			fmt.Println("Cash:", userPortfolios[user].Cash)
			fmt.Println("Stocks:")
			for symbol, stock := range userPortfolios[user].Stocks {
				fmt.Printf("%s: Price: %.2f, Volume: %d\n", symbol, stock.CurrentPrice, stock.TradingVolume)
			}
		}(username)
	}
	wg.Wait()
}
