package domain_test

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	baseAsset  = "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestFundMarket(t *testing.T) {
	t.Parallel()

	market := newTestMarket()
	outpoints := []domain.OutpointWithAsset{
		{
			Asset: baseAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		{
			Asset: quoteAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	}

	err := market.FundMarket(outpoints, baseAsset)
	require.NoError(t, err)
	require.Equal(t, baseAsset, market.BaseAsset)
	require.Equal(t, quoteAsset, market.QuoteAsset)
	require.True(t, market.IsFunded())
}

func TestFailingFundMarket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		baseAsset     string
		outpoints     []domain.OutpointWithAsset
		expectedError error
	}{
		{
			name:      "missing_quote_asset",
			market:    newTestMarket(),
			baseAsset: "0000000000000000000000000000000000000000000000000000000000000000",
			outpoints: []domain.OutpointWithAsset{
				{
					Asset: "0000000000000000000000000000000000000000000000000000000000000000",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  0,
				},
			},
			expectedError: domain.ErrMarketMissingQuoteAsset,
		},
		{
			name:      "missing_base_asset",
			market:    newTestMarket(),
			baseAsset: "0000000000000000000000000000000000000000000000000000000000000000",
			outpoints: []domain.OutpointWithAsset{
				{
					Asset: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  1,
				},
			},
			expectedError: domain.ErrMarketMissingBaseAsset,
		},
		{
			name:      "to_many_assets",
			market:    newTestMarket(),
			baseAsset: "0000000000000000000000000000000000000000000000000000000000000000",
			outpoints: []domain.OutpointWithAsset{
				{
					Asset: "0000000000000000000000000000000000000000000000000000000000000000",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  0,
				},
				{
					Asset: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  1,
				},
				{
					Asset: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  2,
				},
			},
			expectedError: domain.ErrMarketTooManyAssets,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.FundMarket(tt.outpoints, tt.baseAsset)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestMakeTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()

	err := m.MakeTradable()
	require.NoError(t, err)
	require.True(t, m.IsTradable())
}

func TestFailingMakeTradable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		expectedError error
	}{
		{
			name:          "not_funded",
			market:        newTestMarket(),
			expectedError: domain.ErrMarketNotFunded,
		},
		{
			name:          "not_priced",
			market:        newTestMarketFundedWithPluggableStrategy(),
			expectedError: domain.ErrMarketNotPriced,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.MakeTradable()
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestMakeNotTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeNotTradable()
	require.NoError(t, err)
	require.False(t, m.IsTradable())
}

func TestFailingMakeNotTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.MakeNotTradable()
	require.EqualError(t, err, domain.ErrMarketNotFunded.Error())
}

func TestMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()

	err := m.MakeStrategyPluggable()
	require.NoError(t, err)
	require.True(t, m.IsStrategyPluggable())
	require.False(t, m.IsStrategyPluggableInitialized())
}

func TestFailingMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyPluggable()
	require.EqualError(t, err, domain.ErrMarketMustBeClosed.Error())
}

func TestMakeStrategyBalanced(t *testing.T) {
	t.Parallel()

	m := newTestMarketFundedWithPluggableStrategy()

	err := m.MakeStrategyBalanced()
	require.NoError(t, err)
	require.False(t, m.IsStrategyPluggable())
}

func TestFailingMakeStrategyBalanced(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyBalanced()
	require.EqualError(t, err, domain.ErrMarketMustBeClosed.Error())
}

func TestChangeFeeBasisPoint(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()
	newFee := int64(50)

	err := m.ChangeFeeBasisPoint(newFee)
	require.NoError(t, err)
	require.Equal(t, newFee, m.Fee)
}

func TestFailingChangeFeeBasisPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		marketFee     int64
		expectedError error
	}{
		{
			name:          "not_funded",
			market:        newTestMarket(),
			marketFee:     50,
			expectedError: domain.ErrMarketNotFunded,
		},
		{
			name:          "must_be_closed",
			market:        newTestMarketTradable(),
			marketFee:     50,
			expectedError: domain.ErrMarketMustBeClosed,
		},
		{
			name:          "fee_too_low",
			market:        newTestMarketFunded(),
			marketFee:     -1,
			expectedError: domain.ErrMarketFeeTooLow,
		},
		{
			name:          "fee_too_high",
			market:        newTestMarketFunded(),
			marketFee:     10000,
			expectedError: domain.ErrMarketFeeTooHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeFeeBasisPoint(tt.marketFee)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestChangeFixedFee(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()
	baseFee := int64(100)
	quoteFee := int64(200000)

	err := m.ChangeFixedFee(baseFee, quoteFee)
	require.NoError(t, err)
	require.Equal(t, baseFee, m.FixedFee.BaseFee)
	require.Equal(t, quoteFee, m.FixedFee.QuoteFee)
}

func TestFailingChangeFixedFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		market            *domain.Market
		baseFee, quoteFee int64
		expectedError     error
	}{
		{
			name:          "not_funded",
			market:        newTestMarket(),
			expectedError: domain.ErrMarketNotFunded,
		},
		{
			name:          "must_be_closed",
			market:        newTestMarketTradable(),
			expectedError: domain.ErrMarketMustBeClosed,
		},
		{
			name:          "invalid_fixed_base_fee",
			market:        newTestMarketFunded(),
			baseFee:       -1,
			quoteFee:      1000,
			expectedError: domain.ErrInvalidFixedFee,
		},
		{
			name:          "invalid_fixed_quote_fee",
			market:        newTestMarketFunded(),
			baseFee:       100,
			quoteFee:      -1,
			expectedError: domain.ErrInvalidFixedFee,
		},
		{
			name:          "missing_fixed_base_fee",
			market:        newTestMarketFunded(),
			quoteFee:      1000,
			expectedError: domain.ErrMissingFixedFee,
		},
		{
			name:          "missing_fixed_quote_fee",
			market:        newTestMarketFunded(),
			baseFee:       1000,
			expectedError: domain.ErrMissingFixedFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeFixedFee(tt.baseFee, tt.quoteFee)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestChangeMarketPrices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		market     *domain.Market
		basePrice  decimal.Decimal
		quotePrice decimal.Decimal
	}{
		{
			name:       "change_prices_with_balanced_strategy",
			market:     newTestMarketFunded(),
			basePrice:  decimal.NewFromFloat(0.00002),
			quotePrice: decimal.NewFromFloat(50000),
		},
		{
			name:       "change_prices_with_pluggable_strategy",
			market:     newTestMarketFundedWithPluggableStrategy(),
			basePrice:  decimal.NewFromFloat(0.00002),
			quotePrice: decimal.NewFromFloat(50000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeBasePrice(tt.basePrice)
			require.NoError(t, err)
			require.Equal(t, tt.basePrice, tt.market.BaseAssetPrice())

			err = tt.market.ChangeQuotePrice(tt.quotePrice)
			require.NoError(t, err)
			require.Equal(t, tt.quotePrice, tt.market.QuoteAssetPrice())
			require.True(t, tt.market.IsStrategyPluggableInitialized())
		})
	}
}

func TestFailingChangeBasePrice(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.ChangeBasePrice(decimal.NewFromFloat(0.0002))
	require.EqualError(t, err, domain.ErrMarketNotFunded.Error())
}

func TestFailingChangeQuotePrice(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.ChangeQuotePrice(decimal.NewFromFloat(50000))
	require.EqualError(t, err, domain.ErrMarketNotFunded.Error())
}

func TestPreview(t *testing.T) {
	t.Parallel()

	t.Run("market with balanced strategy", func(t *testing.T) {
		market := newTestMarketFunded()
		market.ChangeFeeBasisPoint(100)
		market.ChangeFixedFee(650, 20000000)
		market.MakeTradable()

		tests := []struct {
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expected     *domain.PreviewInfo
		}{
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       2000,
				isBaseAsset:  true,
				isBuy:        true,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
					},
					Amount: 102448966,
					Asset:  quoteAsset,
				},
			},
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000000,
				isBaseAsset:  false,
				isBuy:        true,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
					},
					Amount: 1765,
					Asset:  baseAsset,
				},
			},
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       2000,
				isBaseAsset:  true,
				isBuy:        false,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
					},
					Amount: 57662280,
					Asset:  quoteAsset,
				},
			},
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000000,
				isBaseAsset:  false,
				isBuy:        false,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
					},
					Amount: 3239,
					Asset:  baseAsset,
				},
			},
		}

		for _, tt := range tests {
			preview, err := market.Preview(tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy)
			require.NoError(t, err)
			require.NotNil(t, preview)
			require.Equal(t, tt.expected.Price.BasePrice.String(), preview.Price.BasePrice.String())
			require.Equal(t, tt.expected.Price.QuotePrice.String(), preview.Price.QuotePrice.String())
			require.Equal(t, int(tt.expected.Amount), int(preview.Amount))
			require.Equal(t, tt.expected.Asset, preview.Asset)
		}
	})

	t.Run("market with pluggable strategy", func(t *testing.T) {
		market := newTestMarketFundedWithPluggableStrategy()
		market.MakeNotTradable()
		market.ChangeFeeBasisPoint(100)
		market.ChangeFixedFee(650, 20000000)
		market.ChangeBasePrice(decimal.NewFromFloat(0.000028571429))
		market.ChangeQuotePrice(decimal.NewFromFloat(35000))
		market.MakeTradable()

		tests := []struct {
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expected     *domain.PreviewInfo
		}{
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       2000,
				isBaseAsset:  true,
				isBuy:        true,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
					},
					Amount: 90700000,
					Asset:  quoteAsset,
				},
			},
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000000,
				isBaseAsset:  false,
				isBuy:        true,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
					},
					Amount: 2178,
					Asset:  baseAsset,
				},
			},
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       2000,
				isBaseAsset:  true,
				isBuy:        false,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
					},
					Amount: 49300000,
					Asset:  quoteAsset,
				},
			},
			{
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000000,
				isBaseAsset:  false,
				isBuy:        false,
				expected: &domain.PreviewInfo{
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
					},
					Amount: 3535,
					Asset:  baseAsset,
				},
			},
		}

		for _, tt := range tests {
			preview, err := market.Preview(tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy)
			require.NoError(t, err)
			require.NotNil(t, preview)
			require.Equal(t, tt.expected.Price.BasePrice.String(), preview.Price.BasePrice.String())
			require.Equal(t, tt.expected.Price.QuotePrice.String(), preview.Price.QuotePrice.String())
			require.Equal(t, tt.expected.Asset, preview.Asset)
			require.Equal(t, int(tt.expected.Amount), int(preview.Amount))
		}
	})
}

func TestFailingPreview(t *testing.T) {
	t.Parallel()

	t.Run("market with balanced strategy", func(t *testing.T) {
		market := newTestMarketFunded()
		market.ChangeFeeBasisPoint(100)
		market.MakeTradable()

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       40384,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       1,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       39979,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       4000000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				if preview != nil {
					fmt.Println(preview)
				}
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with pluggable strategy", func(t *testing.T) {
		market := newTestMarketFundedWithPluggableStrategy()
		market.MakeNotTradable()
		market.ChangeFeeBasisPoint(100)
		market.ChangeBasePrice(decimal.NewFromFloat(0.000028571429))
		market.ChangeQuotePrice(decimal.NewFromFloat(35000))
		market.MakeTradable()

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       69999,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "buy with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       3535384947,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       34999,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       115441,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  10000,
				quoteBalance: 40000000,
				amount:       40000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				if preview != nil {
					fmt.Println(preview)
				}
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with balanced strategy and fixed fees", func(t *testing.T) {
		market := newTestMarketFunded()
		market.ChangeFeeBasisPoint(100)
		market.ChangeFixedFee(650, 20000000)
		market.MakeTradable()

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       649,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       26475364,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       649,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       19999999,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       4000000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				if preview != nil {
					fmt.Println(preview)
				}
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with pluggable strategy and fixed fees", func(t *testing.T) {
		t.Parallel()

		market := newTestMarketFundedWithPluggableStrategy()
		market.MakeNotTradable()
		market.ChangeFeeBasisPoint(100)
		market.ChangeFixedFee(650, 20000000)
		market.ChangeBasePrice(decimal.NewFromFloat(0.000028571429))
		market.ChangeQuotePrice(decimal.NewFromFloat(35000))
		market.MakeTradable()

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       649,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       23029999,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "buy with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       3558344947,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       649,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       19999999,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       116018,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       400000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				if preview != nil {
					fmt.Println(preview)
				}
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})
}

func newTestMarket() *domain.Market {
	m, _ := domain.NewMarket(0, 25)
	return m
}

func newTestMarketFunded() *domain.Market {
	outpoints := []domain.OutpointWithAsset{
		{
			Asset: baseAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		{
			Asset: quoteAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	}

	m := newTestMarket()
	m.FundMarket(outpoints, baseAsset)
	return m
}

func newTestMarketTradable() *domain.Market {
	m := newTestMarketFunded()
	m.MakeTradable()
	return m
}

func newTestMarketFundedWithPluggableStrategy() *domain.Market {
	m := newTestMarketFunded()
	m.MakeStrategyPluggable()
	return m
}
