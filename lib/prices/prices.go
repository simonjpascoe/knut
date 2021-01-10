// Copyright 2020 Silvio Böhler
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prices

import (
	"fmt"

	"github.com/sboehler/knut/lib/ledger"
	"github.com/sboehler/knut/lib/model/commodities"

	"github.com/shopspring/decimal"
)

// Prices stores the price for a commodity to a target commodity
// Outer map: target commodity
// Inner map: commodity
// value: price in (target commodity / commodity)
type Prices map[*commodities.Commodity]map[*commodities.Commodity]float64

// New creates prices.
func New() Prices {
	return map[*commodities.Commodity]map[*commodities.Commodity]float64{}
}

// Insert inserts a new price.
func (p Prices) Insert(pr *ledger.Price) {
	p.addPrice(pr.Target, pr.Commodity, pr.Price)
	p.addPrice(pr.Commodity, pr.Target, 1/pr.Price)
}

func (p Prices) addPrice(target, commodity *commodities.Commodity, pr float64) {
	i, ok := p[target]
	if !ok {
		i = map[*commodities.Commodity]float64{}
		p[target] = i
	}
	i[commodity] = pr
}

// Normalize creates a normalized price map for the given commodity.
func (p Prices) Normalize(c *commodities.Commodity) NormalizedPrices {
	// prices in (target commodity / commodity)
	todo := NormalizedPrices{c: 1}
	done := NormalizedPrices{}

	var (
		currentC *commodities.Commodity
		currentP float64
	)

	for len(todo) > 0 {
		// we're interested in an arbitrary element of the map
		for currentC, currentP = range todo {
			break
		}
		done[currentC] = currentP
		for neighbor, price := range p[currentC] {
			if _, ok := done[neighbor]; ok {
				continue
			}
			todo[neighbor] = price * currentP
		}
		delete(todo, currentC)
	}
	return done
}

// Copy creates a deep copy.
func (p Prices) Copy() Prices {
	pr := New()
	for tc, ps := range p {
		for c, v := range ps {
			pr.addPrice(tc, c, v)
		}
	}
	return pr
}

// NormalizedPrices is a map representing the price of
// commodities in some base commodity.
type NormalizedPrices map[*commodities.Commodity]float64

// Valuate valuates the given amount.
func (n NormalizedPrices) Valuate(c *commodities.Commodity, a decimal.Decimal) (decimal.Decimal, error) {
	price, ok := n[c]
	if !ok {
		return decimal.Zero, fmt.Errorf("No price found for %v in %v", c, n)
	}
	amount, _ := a.Float64()
	value := amount * price
	return decimal.NewFromFloat(value), nil
}
