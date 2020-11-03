package application

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
)

type OperatorService interface {
	DepositMarket(
		ctx context.Context,
		quoteAsset string,
	) (string, error)
	DepositFeeAccount(
		ctx context.Context,
	) (address string, blindingKey string, err error)
	OpenMarket(
		ctx context.Context,
		baseAsset string,
		quoteAsset string,
	) error
	CloseMarket(
		ctx context.Context,
		baseAsset string,
		quoteAsset string,
	) error
	UpdateMarketFee(
		ctx context.Context,
		req MarketWithFee,
	) (*MarketWithFee, error)
	UpdateMarketPrice(
		ctx context.Context,
		req MarketWithPrice,
	) error
	UpdateMarketStrategy(
		ctx context.Context,
		req MarketStrategy,
	) error
	ListSwaps(
		ctx context.Context,
	) (*pb.ListSwapsReply, error)
	WithdrawMarketFunds(
		ctx context.Context,
		req WithdrawMarketReq,
	) (
		[]byte,
		error,
	)
	FeeAccountBalance(ctx context.Context) (
		int64,
		error,
	)
	ListMarket(
		ctx context.Context,
	) ([]MarketInfo, error)
}

type operatorService struct {
	marketRepository  domain.MarketRepository
	vaultRepository   domain.VaultRepository
	tradeRepository   domain.TradeRepository
	unspentRepository domain.UnspentRepository
	explorerSvc       explorer.Service
	crawlerSvc        crawler.Service
}

// NewOperatorService is a constructor function for OperatorService.
func NewOperatorService(
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	tradeRepository domain.TradeRepository,
	unspentRepository domain.UnspentRepository,
	explorerSvc explorer.Service,
	crawlerSvc crawler.Service,
) OperatorService {
	return &operatorService{
		marketRepository:  marketRepository,
		vaultRepository:   vaultRepository,
		tradeRepository:   tradeRepository,
		unspentRepository: unspentRepository,
		explorerSvc:       explorerSvc,
		crawlerSvc:        crawlerSvc,
	}
}

func (o *operatorService) DepositMarket(
	ctx context.Context,
	quoteAsset string,
) (string, error) {

	var address string

	var accountIndex int
	_, a, err := o.marketRepository.GetMarketByAsset(
		ctx,
		quoteAsset,
	)
	if err != nil {
		return "", err
	}
	accountIndex = a

	if accountIndex == -1 {
		_, latestAccountIndex, err := o.marketRepository.GetLatestMarket(
			ctx,
		)
		if err != nil {
			return "", err
		}

		accountIndex = latestAccountIndex + 1
		_, err = o.marketRepository.GetOrCreateMarket(ctx, accountIndex)
		if err != nil {
			return "", err
		}
	}

	err = o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(accountIndex)
			if err != nil {
				return nil, err
			}

			address = addr

			o.crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: accountIndex,
				Address:      addr,
				BlindingKey:  blindingKey,
			})

			return v, nil
		})
	if err != nil {
		return "", err
	}

	return address, nil
}

func (o *operatorService) DepositFeeAccount(
	ctx context.Context,
) (address string, blindingKey string, err error) {
	err = o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			addr, _, blindKey, err := v.DeriveNextExternalAddressForAccount(
				domain.FeeAccount,
			)
			if err != nil {
				return nil, err
			}

			address = addr
			blindingKey = hex.EncodeToString(blindKey)

			o.crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: domain.FeeAccount,
				Address:      addr,
				BlindingKey:  blindKey,
			})

			return v, nil
		})
	return
}

func (o *operatorService) OpenMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return errors.New("invalid base asset")
	}

	_, marketAccountIndex, err := o.marketRepository.GetMarketByAsset(ctx, quoteAsset)
	if err != nil {
		return err
	}

	var outpoints []domain.OutpointWithAsset
	if marketAccountIndex < 0 {
		_, marketAccountIndex, err = o.marketRepository.GetLatestMarket(ctx)
		if err != nil {
			return err
		}

		addresses, _, err :=
			o.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, marketAccountIndex)
		if err != nil {
			return err
		}
		unspents, err := o.unspentRepository.GetUnspentsForAddresses(ctx, addresses)
		if err != nil {
			return err
		}

		outpoints = make([]domain.OutpointWithAsset, 0, len(unspents))
		for _, u := range unspents {
			outpoints = append(outpoints, domain.OutpointWithAsset{
				Txid:  u.TxID,
				Vout:  int(u.VOut),
				Asset: u.AssetHash,
			})
		}
	}

	if err := o.marketRepository.UpdateMarket(ctx, marketAccountIndex, func(m *domain.Market) (*domain.Market, error) {
		if m.IsTradable() {
			return m, nil
		}

		if len(outpoints) > 0 {
			if err := m.FundMarket(outpoints); err != nil {
				return nil, err
			}
		}

		if err := m.MakeTradable(); err != nil {
			return nil, err
		}
		return m, nil
	}); err != nil {
		return err
	}

	return nil
}

func (o *operatorService) CloseMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return errors.New("invalid base asset")
	}

	err := o.marketRepository.CloseMarket(
		ctx,
		quoteAsset,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateMarketFee changes the Liquidity Provider fee for the given market.
// MUST be expressed as basis point.
// Eg. To change the fee on each swap from 0.25% to 1% you need to pass down 100
// The Market MUST be closed before doing this change.
func (o *operatorService) UpdateMarketFee(
	ctx context.Context,
	req MarketWithFee,
) (*MarketWithFee, error) {

	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return nil, domain.ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, domain.ErrMarketNotExist
	}

	//Updates the fee and the fee asset
	err = o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			if err := m.ChangeFee(req.BasisPoint); err != nil {
				return nil, err
			}

			if err := m.ChangeFeeAsset(req.FeeAsset); err != nil {
				return nil, err
			}

			return m, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Ignore errors. If we reached this point it must exists.
	mkt, _ := o.marketRepository.GetOrCreateMarket(
		ctx,
		accountIndex,
	)

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			FeeAsset:   mkt.FeeAsset,
			BasisPoint: mkt.Fee,
		},
	}, nil
}

// UpdateMarketPrice rpc updates the price for the given market
func (o *operatorService) UpdateMarketPrice(
	ctx context.Context,
	req MarketWithPrice,
) error {
	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return domain.ErrMarketNotExist
	}

	//Updates the base price and the quote price
	return o.marketRepository.UpdatePrices(
		ctx,
		accountIndex,
		domain.Prices{
			BasePrice:  req.Price.BasePrice,
			QuotePrice: req.Price.QuotePrice,
		},
	)
}

// UpdateMarketStrategy changes the current market making strategy, either using an automated
// market making formula or a pluggable price feed
func (o *operatorService) UpdateMarketStrategy(
	ctx context.Context,
	req MarketStrategy,
) error {
	// Checks if base asset is correct
	if req.BaseAsset != config.GetString(config.BaseAssetKey) {
		return domain.ErrMarketNotExist
	}
	//Checks if market exist
	_, accountIndex, err := o.marketRepository.GetMarketByAsset(
		ctx,
		req.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return domain.ErrMarketNotExist
	}

	//For now we support only BALANCED or PLUGGABLE (ie. price feed)
	requestStrategy := req.Strategy
	//Updates the strategy
	return o.marketRepository.UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {

			switch requestStrategy {

			case domain.StrategyTypePluggable:
				if err := m.MakeStrategyPluggable(); err != nil {
					return nil, err
				}

			case domain.StrategyTypeBalanced:
				if err := m.MakeStrategyBalanced(); err != nil {
					return nil, err
				}

			default:
				return nil, errors.New("strategy not supported")
			}

			return m, nil
		},
	)
}

// ListSwaps returns the list of all swaps processed by the daemon
func (o *operatorService) ListSwaps(
	ctx context.Context,
) (*pb.ListSwapsReply, error) {
	trades, err := o.tradeRepository.GetAllTrades(ctx)
	if err != nil {
		return nil, err
	}

	markets, err := o.getMarketsForTrades(ctx, trades)
	if err != nil {
		return nil, err
	}

	swaps := tradesToSwapInfo(markets, trades)
	return &pb.ListSwapsReply{
		Swaps: swaps,
	}, nil
}

//ListMarket a set of informations about all the markets.
func (o *operatorService) ListMarket(
	ctx context.Context,
) ([]MarketInfo, error) {
	markets, err := o.marketRepository.GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketInfos := make([]MarketInfo, len(markets), len(markets))

	for index, market := range markets {
		marketInfos[index] = MarketInfo{
			Market: Market{
				BaseAsset:  market.BaseAsset,
				QuoteAsset: market.QuoteAsset,
			},
			Fee: Fee{
				BasisPoint: market.Fee,
				FeeAsset:   market.FeeAsset,
			},
			Tradable:     market.Tradable,
			StrategyType: market.Strategy.Type,
		}
	}

	return marketInfos, nil
}

func (o *operatorService) getMarketsForTrades(
	ctx context.Context,
	trades []*domain.Trade,
) (map[string]*domain.Market, error) {
	markets := map[string]*domain.Market{}
	for _, trade := range trades {
		market, accountIndex, err := o.marketRepository.GetMarketByAsset(
			ctx,
			trade.MarketQuoteAsset,
		)
		if err != nil {
			return nil, err
		}
		if accountIndex < 0 {
			return nil, domain.ErrMarketNotExist
		}
		if _, ok := markets[trade.MarketQuoteAsset]; !ok {
			markets[trade.MarketQuoteAsset] = market
		}
	}
	return markets, nil
}

func tradesToSwapInfo(
	markets map[string]*domain.Market,
	trades []*domain.Trade,
) []*pb.SwapInfo {
	info := make([]*pb.SwapInfo, 0, len(trades))
	for _, trade := range trades {
		requestMsg := trade.SwapRequestMessage()
		fee := &pbtypes.Fee{
			Asset:      markets[trade.MarketQuoteAsset].FeeAsset,
			BasisPoint: markets[trade.MarketQuoteAsset].Fee,
		}
		i := &pb.SwapInfo{
			Status:           trade.Status.Code,
			AmountP:          requestMsg.GetAmountP(),
			AssetP:           requestMsg.GetAssetP(),
			AmountR:          requestMsg.GetAmountR(),
			AssetR:           requestMsg.GetAssetR(),
			MarketFee:        fee,
			RequestTimeUnix:  trade.SwapRequestTime(),
			AcceptTimeUnix:   trade.SwapAcceptTime(),
			CompleteTimeUnix: trade.SwapCompleteTime(),
			ExpiryTimeUnix:   trade.SwapExpiryTime(),
		}
		info = append(info, i)
	}
	return info
}

func (o *operatorService) WithdrawMarketFunds(
	ctx context.Context,
	req WithdrawMarketReq,
) (
	[]byte,
	error,
) {
	var rawTx []byte

	market, _, err := o.marketRepository.GetMarketByAsset(ctx, req.QuoteAsset)
	if err != nil {
		return nil, err
	}

	if market == nil {
		return nil, errors.New("market does not exists")
	}

	outs := make([]TxOut, 0)
	if req.BalanceToWithdraw.BaseAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.BaseAsset,
			Value:   req.BalanceToWithdraw.BaseAmount,
			Address: req.Address,
		})
	}
	if req.BalanceToWithdraw.QuoteAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.QuoteAsset,
			Value:   req.BalanceToWithdraw.QuoteAmount,
			Address: req.Address,
		})
	}
	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs)
	if err != nil {
		return nil, err
	}

	marketUnspents, err := o.getAllUnspentsForAccount(ctx, market.AccountIndex)
	if err != nil {
		return nil, err
	}
	if len(marketUnspents) <= 0 {
		return nil, ErrWalletNotFunded
	}

	feeUnspents, err := o.getAllUnspentsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return nil, err
	}
	if len(feeUnspents) <= 0 {
		return nil, ErrWalletNotFunded
	}

	var addressesToObserve []*crawler.AddressObservable
	err = o.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			mnemonic, err := v.GetMnemonicSafe()
			if err != nil {
				return nil, err
			}
			marketAccount, err := v.AccountByIndex(market.AccountIndex)
			if err != nil {
				return nil, err
			}
			feeAccount, err := v.AccountByIndex(domain.FeeAccount)
			if err != nil {
				return nil, err
			}

			changePathsByAsset := map[string]string{}
			feeChangePathByAsset := map[string]string{}
			for _, asset := range getAssetsOfOutputs(outputs) {
				addr, script, blindkey, err := v.DeriveNextInternalAddressForAccount(
					market.AccountIndex,
				)
				if err != nil {
					return nil, err
				}

				derivationPath, _ := marketAccount.DerivationPathByScript[script]
				changePathsByAsset[asset] = derivationPath
				addressesToObserve = append(addressesToObserve, &crawler.AddressObservable{
					AccountIndex: market.AccountIndex,
					Address:      addr,
					BlindingKey:  blindkey,
				})
			}

			feeAddress, script, feeBlindkey, err := v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
			if err != nil {
				return nil, err
			}
			feeChangePathByAsset[config.GetNetwork().AssetID] = feeAccount.DerivationPathByScript[script]

			addressesToObserve = append(addressesToObserve, &crawler.AddressObservable{
				AccountIndex: market.AccountIndex,
				Address:      feeAddress,
				BlindingKey:  feeBlindkey,
			})

			txHex, _, err := sendToMany(sendToManyOpts{
				mnemonic:             mnemonic,
				unspents:             marketUnspents,
				feeUnspents:          feeUnspents,
				outputs:              outputs,
				outputsBlindingKeys:  outputsBlindingKeys,
				milliSatPerByte:      int(req.MillisatPerByte),
				changePathsByAsset:   changePathsByAsset,
				feeChangePathByAsset: feeChangePathByAsset,
			})
			if err != nil {
				return nil, err
			}

			if req.Push {
				if _, err := o.explorerSvc.BroadcastTransaction(txHex); err != nil {
					return nil, err
				}
			}

			rawTx, _ = hex.DecodeString(txHex)

			return v, nil
		},
	)
	if err != nil {
		return nil, err
	}

	for _, obs := range addressesToObserve {
		o.crawlerSvc.AddObservable(obs)
	}

	return rawTx, nil
}

func (o *operatorService) FeeAccountBalance(ctx context.Context) (
	int64,
	error,
) {
	addresses, _, err := o.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return 0, err
	}

	baseAssetAmount, err := o.unspentRepository.GetBalance(
		ctx,
		addresses,
		config.GetString(config.BaseAssetKey),
	)
	if err != nil {
		return -1, err
	}
	return int64(baseAssetAmount), nil
}

func (o *operatorService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
) ([]explorer.Utxo, error) {
	addresses, _, err := o.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	unspents, err := o.unspentRepository.GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, u := range unspents {
		utxos = append(utxos, u.ToUtxo())
	}
	return utxos, nil
}
