package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestNewMarket(t *testing.T) {
	t.Parallel()

	accountIndex := 0
	fee := int64(25)

	m, err := domain.NewMarket(accountIndex, fee)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, accountIndex, m.AccountIndex)
	require.Equal(t, fee, m.Fee.BasisPoint)
	require.Zero(t, m.Fee.FixedBaseFee)
	require.Zero(t, m.Fee.FixedQuoteFee)
	require.False(t, m.IsStrategyPluggable())
}

func TestFailingNewMarket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accountIndex  int
		fee           int64
		expectedError error
	}{
		{"invalid_account", -1, 25, domain.ErrInvalidAccount},
		{"fee_too_low", 0, -1, domain.ErrMarketFeeTooLow},
		{"fee_too_high", 0, 10000, domain.ErrMarketFeeTooHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewMarket(tt.accountIndex, tt.fee)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}
