package grpchandler

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operatorHandler struct {
	pb.UnimplementedOperatorServer
	operatorSvc application.OperatorService
}

// NewOperatorHandler is a constructor function returning an protobuf OperatorServer.
func NewOperatorHandler(operatorSvc application.OperatorService) pb.OperatorServer {
	return &operatorHandler{
		operatorSvc: operatorSvc,
	}
}

func (o operatorHandler) DepositMarket(
	ctx context.Context,
	req *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	address, err := o.operatorSvc.DepositMarket(ctx, req.GetMarket().GetBaseAsset(), req.GetMarket().GetQuoteAsset())
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}
	return &pb.DepositMarketReply{
		Address: address,
	}, nil
}

func (o operatorHandler) DepositFeeAccount(
	ctx context.Context,
	req *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {
	address, blindingKey, err := o.operatorSvc.DepositFeeAccount(ctx)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.DepositFeeAccountReply{
		Address:  address,
		Blinding: blindingKey,
	}, nil
}

func (o operatorHandler) OpenMarket(
	ctx context.Context,
	req *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.OpenMarket(
		ctx,
		market.GetBaseAsset(),
		market.GetQuoteAsset(),
	); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OpenMarketReply{}, nil
}

func (o operatorHandler) CloseMarket(
	ctx context.Context,
	req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.CloseMarket(
		ctx,
		market.GetBaseAsset(),
		market.GetQuoteAsset(),
	); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CloseMarketReply{}, nil
}

func (o operatorHandler) UpdateMarketFee(
	ctx context.Context,
	req *pb.UpdateMarketFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	marketWithFee := req.GetMarketWithFee()
	if err := validateMarketWithFee(marketWithFee); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwf := application.MarketWithFee{
		Market: application.Market{
			BaseAsset:  marketWithFee.GetMarket().GetBaseAsset(),
			QuoteAsset: marketWithFee.GetMarket().GetQuoteAsset(),
		},
		Fee: application.Fee{
			FeeAsset:   marketWithFee.GetFee().GetAsset(),
			BasisPoint: marketWithFee.GetFee().GetBasisPoint(),
		},
	}
	res, err := o.operatorSvc.UpdateMarketFee(ctx, mwf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketFeeReply{
		MarketWithFee: &pbtypes.MarketWithFee{
			Market: &pbtypes.Market{
				BaseAsset:  res.BaseAsset,
				QuoteAsset: res.QuoteAsset,
			},
			Fee: &pbtypes.Fee{
				Asset:      res.FeeAsset,
				BasisPoint: res.BasisPoint,
			},
		},
	}, nil
}

func (o operatorHandler) UpdateMarketPrice(
	ctx context.Context,
	req *pb.UpdateMarketPriceRequest,
) (*pb.UpdateMarketPriceReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	price := req.GetPrice()
	if err := validatePrice(price); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwp := application.MarketWithPrice{
		Market: application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Price: application.Price{
			BasePrice:  decimal.NewFromFloat32(price.GetBasePrice()),
			QuotePrice: decimal.NewFromFloat32(price.GetQuotePrice()),
		},
	}
	if err := o.operatorSvc.UpdateMarketPrice(ctx, mwp); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketPriceReply{}, nil
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context,
	req *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	strategyType := req.GetStrategyType()
	if err := validateStrategyType(strategyType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ms := application.MarketStrategy{
		Market: application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Strategy: domain.StrategyType(strategyType),
	}
	if err := o.operatorSvc.UpdateMarketStrategy(ctx, ms); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketStrategyReply{}, nil
}

func (o operatorHandler) ListSwaps(
	ctx context.Context,
	req *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
	swapInfos, err := o.operatorSvc.ListSwaps(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbSwapInfos := make([]*pb.SwapInfo, len(swapInfos), len(swapInfos))

	for index, swapInfo := range swapInfos {
		pbSwapInfos[index] = &pb.SwapInfo{
			Status: pb.SwapStatus(swapInfo.Status),
			AmountP: swapInfo.AmountP,
			AssetP: swapInfo.AssetP,
			AmountR: swapInfo.AmountR,
			AssetR: swapInfo.AssetR,
			MarketFee: &pbtypes.Fee{
				Asset: swapInfo.MarketFee.FeeAsset,
				BasisPoint: swapInfo.MarketFee.BasisPoint,
			},
			RequestTimeUnix: swapInfo.RequestTimeUnix,
			AcceptTimeUnix: swapInfo.AcceptTimeUnix,
			CompleteTimeUnix: swapInfo.RequestTimeUnix,
			ExpiryTimeUnix: swapInfo.ExpiryTimeUnix,
		}
	}
	
	return &pb.ListSwapsReply{Swaps: pbSwapInfos}, nil
}

func (o operatorHandler) WithdrawMarket(
	ctx context.Context,
	req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rawTx, err := o.operatorSvc.WithdrawMarketFunds(
		ctx,
		application.WithdrawMarketReq{
			Market: application.Market{
				BaseAsset:  req.GetMarket().GetBaseAsset(),
				QuoteAsset: req.GetMarket().GetQuoteAsset(),
			},
			BalanceToWithdraw: application.Balance{
				BaseAmount:  req.GetBalanceToWithdraw().GetBaseAmount(),
				QuoteAmount: req.GetBalanceToWithdraw().GetQuoteAmount(),
			},
			MillisatPerByte: req.GetMillisatPerByte(),
			Address:         req.GetAddress(),
			Push:            req.GetPush(),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.WithdrawMarketReply{
		RawTx: rawTx,
	}, nil
}

func (o operatorHandler) BalanceFeeAccount(
	ctx context.Context,
	req *pb.BalanceFeeAccountRequest,
) (*pb.BalanceFeeAccountReply, error) {

	balance, err := o.operatorSvc.FeeAccountBalance(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.BalanceFeeAccountReply{
		Balance: balance,
	}, nil
}

// ListMarket returns the result of the ListMarket method of the operator service.
func (o operatorHandler) ListMarket(
	ctx context.Context,
	req *pb.ListMarketRequest,
) (*pb.ListMarketReply, error) {
	marketInfos, err := o.operatorSvc.ListMarket(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbMarketInfos := make([]*pb.MarketInfo, len(marketInfos), len(marketInfos))

	for index, marketInfo := range marketInfos {
		pbMarketInfos[index] = &pb.MarketInfo{
			Fee: &pbtypes.Fee{
				BasisPoint: marketInfo.Fee.BasisPoint,
				Asset:      marketInfo.Fee.FeeAsset,
			},
			Market: &pbtypes.Market{
				BaseAsset:  marketInfo.Market.BaseAsset,
				QuoteAsset: marketInfo.Market.QuoteAsset,
			},
			Tradable:     marketInfo.Tradable,
			StrategyType: pb.StrategyType(marketInfo.StrategyType),
		}
	}

	return &pb.ListMarketReply{Markets: pbMarketInfos}, nil
}

func (o operatorHandler) ReportMarketFee(
	ctx context.Context,
	req *pb.ReportMarketFeeRequest,
) (*pb.ReportMarketFeeReply, error) {
	if err := validateMarket(req.GetMarket()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	report, err := o.operatorSvc.GetCollectedMarketFee(
		ctx,
		application.Market{
			BaseAsset:  req.GetMarket().GetBaseAsset(),
			QuoteAsset: req.GetMarket().GetQuoteAsset(),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	collectedFees := make([]*pbtypes.Fee, 0)
	for _, v := range report.CollectedFees {
		collectedFees = append(collectedFees, &pbtypes.Fee{
			Asset:      v.FeeAsset,
			BasisPoint: v.BasisPoint,
		})
	}

	return &pb.ReportMarketFeeReply{
		CollectedFees:              collectedFees,
		TotalCollectedFeesPerAsset: report.TotalCollectedFeesPerAsset,
	}, nil
}

func validateMarketWithFee(marketWithFee *pbtypes.MarketWithFee) error {
	if marketWithFee == nil {
		return errors.New("market with fee is null")
	}

	if err := validateMarket(marketWithFee.GetMarket()); err != nil {
		return err
	}

	if err := validateFee(marketWithFee.GetFee()); err != nil {
		return err
	}

	return nil
}

func validateMarket(market *pbtypes.Market) error {
	if market == nil {
		return errors.New("market is null")
	}
	if len(market.GetBaseAsset()) <= 0 || len(market.GetQuoteAsset()) <= 0 {
		return errors.New("base asset or quote asset are null")
	}
	return nil
}

func validateFee(fee *pbtypes.Fee) error {
	if fee == nil {
		return errors.New("fee is null")
	}
	if len(fee.GetAsset()) <= 0 {
		return errors.New("fee asset is null")
	}
	if fee.GetBasisPoint() <= 0 {
		return errors.New("fee basis point is too low")
	}
	return nil
}

func validatePrice(price *pbtypes.Price) error {
	if price == nil {
		return errors.New("price is null")
	}
	if price.GetBasePrice() <= 0 || price.GetQuotePrice() <= 0 {
		return errors.New("base or quote price are too low")
	}
	return nil
}

func validateStrategyType(sType pb.StrategyType) error {
	if domain.StrategyType(sType) < domain.StrategyTypePluggable ||
		domain.StrategyType(sType) > domain.StrategyTypeUnbalanced {
		return errors.New("strategy type is unknown")
	}
	return nil
}
