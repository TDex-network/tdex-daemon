package domain

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

// MakeTradable ...
func (m *Market) MakeTradable() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsStrategyPluggable() && !m.IsStrategyPluggableInitialized() {
		return ErrNotPriced
	}

	m.Tradable = true
	return nil
}

// MakeNotTradable ...
func (m *Market) MakeNotTradable() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	m.Tradable = false
	return nil
}

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.Tradable
}

func validateFee(basisPoint int64) error {
	if basisPoint < 1 || basisPoint > 9999 {
		return errors.New("percentage of the fee on each swap must be > 0.01 and < 99")
	}

	return nil
}

// IsFunded method returns true if the market contains a non empty funding tx outpoint for each asset
func (m *Market) IsFunded() bool {
	return m.BaseAsset != "" && m.QuoteAsset != ""
}

// FundMarket adds the assets of market from the given array of outpoints.
// Since the list of outpoints can contain an infinite number of utxos with
// different  assets, they're fistly indexed by their asset, then the market's
// base asset is updated if found in the list, otherwise only the very first
// asset type is used as the market's quote asset, discarding the others that
// should be manually transferred to some other address because they won't be
// used by the daemon.
func (m *Market) FundMarket(fundingTxs []OutpointWithAsset) error {
	if m.IsFunded() {
		return nil
	}

	baseAssetHash := config.GetString(config.BaseAssetKey)
	assetCount := make(map[string]int)
	for _, o := range fundingTxs {
		assetCount[o.Asset]++
	}

	if len(assetCount) > 2 {
		return fmt.Errorf(
			"outpoints must be at most of 2 different type of assets, but %d were "+
				"found and it's not possible to determine what's the correct asset "+
				"pair of the market! It's mandatory to move funds from market's "+
				"addresses so that they own only utxos of 2 different asset types "+
				"(base asset included).", len(assetCount),
		)
	}

	for asset := range assetCount {
		if asset == baseAssetHash {
			if !m.IsFunded() {
				m.BaseAsset = baseAssetHash
			}
		} else {
			if !m.IsFunded() {
				m.QuoteAsset = asset
			}
		}
	}

	return nil
}

// ChangeFee ...
func (m *Market) ChangeFee(fee int64) error {

	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClose
	}

	if err := validateFee(fee); err != nil {
		return err
	}

	m.Fee = fee
	return nil
}

// ChangeFeeAsset ...
func (m *Market) ChangeFeeAsset(asset string) error {
	// In case of empty asset hash, no updates happens and therefore it exit without error
	if asset == "" {
		return nil
	}

	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClose
	}

	if asset != m.BaseAsset && asset != m.QuoteAsset {
		return errors.New("the given asset must be either the base or quote" +
			" asset in the pair")
	}

	m.FeeAsset = asset
	return nil
}

// BaseAssetPrice returns the latest price for the base asset
func (m *Market) BaseAssetPrice() decimal.Decimal {
	_, price := getLatestPrice(m.BasePrice)

	return decimal.Decimal(price)
}

// QuoteAssetPrice returns the latest price for the quote asset
func (m *Market) QuoteAssetPrice() decimal.Decimal {
	_, price := getLatestPrice(m.QuotePrice)

	return decimal.Decimal(price)
}

// ChangeBasePrice ...
func (m *Market) ChangeBasePrice(price decimal.Decimal) error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	// TODO add logic to be sure that the price do not change to much from the latest one

	timestamp := uint64(time.Now().Unix())
	if _, ok := m.BasePrice[timestamp]; ok {
		return ErrPriceExists
	}

	m.BasePrice[timestamp] = Price(price)
	return nil
}

// ChangeQuotePrice ...
func (m *Market) ChangeQuotePrice(price decimal.Decimal) error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	//TODO check if the previous price is changing too much as security measure

	timestamp := uint64(time.Now().Unix())
	if _, ok := m.QuotePrice[timestamp]; ok {
		return ErrPriceExists
	}

	m.QuotePrice[timestamp] = Price(price)
	return nil
}

// IsZero ...
func (pt PriceByTime) IsZero() bool {
	return len(pt) == 0
}

// IsZero ...
func (p Price) IsZero() bool {
	return decimal.Decimal(p).Equal(decimal.NewFromInt(0))
}

func getLatestPrice(keyValue PriceByTime) (uint64, Price) {
	if keyValue.IsZero() {
		return uint64(0), Price(decimal.NewFromInt(0))
	}

	keys := make([]uint64, 0, len(keyValue))
	for k := range keyValue {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	latestKey := keys[len(keys)-1]
	latestValue := keyValue[latestKey]
	return latestKey, latestValue
}

// IsStrategyPluggable returns true if the the startegy isn't automated.
func (m *Market) IsStrategyPluggable() bool {
	return m.Strategy.IsZero()
}

// IsStrategyPluggableInitialized returns true if the prices have been set.
func (m *Market) IsStrategyPluggableInitialized() bool {
	return !m.BasePrice.IsZero() && !m.QuotePrice.IsZero()
}

// MakeStrategyPluggable makes the current market using a given price
//(ie. set via UpdateMarketPrice rpc either manually or a price feed plugin)
func (m *Market) MakeStrategyPluggable() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketMustBeClose
	}

	m.Strategy = mm.MakingStrategy{}

	return nil
}

// MakeStrategyBalanced makes the current market using a balanced AMM formula 50/50
func (m *Market) MakeStrategyBalanced() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketMustBeClose
	}

	m.Strategy = mm.NewStrategyFromFormula(formula.BalancedReserves{})

	return nil
}
