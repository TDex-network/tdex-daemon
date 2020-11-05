package application

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
)

const marketRepoIsEmpty = true
const tradeRepoIsEmpty = true

var baseAsset = config.GetString(config.BaseAssetKey)

func TestListMarket(t *testing.T) {
	t.Run("ListMarket should return an empty list and a nil error if market repository is empty", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)
		marketInfos, err := operatorService.ListMarket(ctx)
		close()
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(marketInfos))
	})

	t.Run("ListMarket should return the number of markets in the market repository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty, tradeRepoIsEmpty)
		marketInfos, err := operatorService.ListMarket(ctx)
		close()
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(marketInfos))
	})
}

func TestDepositMarket(t *testing.T) {

	t.Run("DepositMarket with new market", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		address, err := operatorService.DepositMarket(ctx, "", "")
		assert.Equal(t, nil, err)

		assert.Equal(
			t,
			"el1qqvead5fpxkjyyl3zwukr7twqrnag40ls0y052s547smxdyeus209ppkmtdyemgkz4rjn8ss8fhjrzc3q9evt7atrgtpff2thf",
			address,
		)

		close()
	})

	t.Run("DepositMarket with invalid base asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		validQuoteAsset := "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"
		emptyAddress, err := operatorService.DepositMarket(ctx, "", validQuoteAsset)
		assert.Equal(t, domain.ErrInvalidBaseAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})

	t.Run("DepositMarket with valid base asset and empty quote asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, baseAsset, "")
		assert.Equal(t, domain.ErrInvalidQuoteAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})

	t.Run("DepositMarket with valid base asset and invalid quote asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, baseAsset, "ldjbwjkbfjksdbjkvcsbdjkbcdsjkb")
		assert.Equal(t, domain.ErrInvalidQuoteAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})
}

func TestDepositMarketWithCrawler(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("Get address to deposit, fund market and get next address for the market", func(t *testing.T) {

		startNigiriAndWait()

		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		address, err := operatorService.DepositMarket(ctx, "", "")
		assert.Equal(t, nil, err)

		assert.Equal(t, nil, err)
		assert.Equal(
			t,
			"el1qqvead5fpxkjyyl3zwukr7twqrnag40ls0y052s547smxdyeus209ppkmtdyemgkz4rjn8ss8fhjrzc3q9evt7atrgtpff2thf",
			address,
		)

		// Let's depsoit both assets on the same address
		explorerSvc := explorer.NewService(RegtestExplorerAPI)
		_, err = explorerSvc.Faucet(address)
		assert.Equal(t, nil, err)
		time.Sleep(1500 * time.Millisecond)

		_, quoteAsset, err := explorerSvc.Mint(address, 5)
		assert.Equal(t, nil, err)
		time.Sleep(1500 * time.Millisecond)

		// we try to get a child address for the quote asset. Since is not being expicitly initialized, should return ErrMarketNotExist
		failToGetChildAddress, err := operatorService.DepositMarket(ctx, baseAsset, quoteAsset)
		assert.Equal(t, domain.ErrMarketNotExist, err)
		assert.Equal(
			t,
			"",
			failToGetChildAddress,
		)

		// Now we try to intialize (ie. fund) the market by opening it
		err = operatorService.OpenMarket(ctx, baseAsset, quoteAsset)
		assert.Equal(t, nil, err)

		// Now we can derive a childAddress
		childAddress, err := operatorService.DepositMarket(ctx, baseAsset, quoteAsset)
		assert.Equal(t, nil, err)
		assert.Equal(
			t,
			"el1qqfzjp0y057j60avxqgmj9aycqhlq7ke20v20c8dkml68jjs0fu09u9sn55uduay46yyt25tcny0rfqejly5x6dgjw44uk9p8r",
			childAddress,
		)

		close()
		stopNigiri()
	})
}

func TestUpdateMarketPrice(t *testing.T) {
	operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty, tradeRepoIsEmpty)
	defer close()
	
	updateMarketPriceRequest := func(basePrice int, quotePrice int) error {
		args := MarketWithPrice{
			Market: Market{
				BaseAsset: network.Regtest.AssetID, 
				QuoteAsset: marketUnspents[1].AssetHash,
			},
			Price: Price{
				BasePrice: decimal.NewFromInt(int64(basePrice)), 
				QuotePrice: decimal.NewFromInt(int64(quotePrice)),
			},
		}
		return operatorService.UpdateMarketPrice(ctx, args)
	}
	
	t.Run("should not return an error if the price is valid and market is found", func (t *testing.T) {
		err := updateMarketPriceRequest(10, 1000)
		assert.Equal(t, nil, err)
	})

	t.Run("should return an error if the prices are <= 0", func(t *testing.T) {
		err := updateMarketPriceRequest(-1, 10000)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should return an error if the prices are greater than 2099999997690000", func(t *testing.T) {
		err := updateMarketPriceRequest(1,  2099999997690000 + 1)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should return an error if one of the prices are equal to zero", func(t *testing.T) {
		err := updateMarketPriceRequest(102, 0)
		assert.NotEqual(t, nil, err)
	})
}
func TestListSwap(t *testing.T) {
	t.Run("ListSwap should return an empty list and a nil error if there is not trades in TradeRepository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)
		defer close()

		swapInfos, err := operatorService.ListSwaps(ctx)
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(swapInfos))
	})

	t.Run("ListSwap should return the SwapInfo according to the number of trades in the TradeRepository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty, !tradeRepoIsEmpty)
		defer close()
		
		swapInfos, err := operatorService.ListSwaps(ctx)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, len(swapInfos))
	})
}
