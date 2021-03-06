package domain_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/vulpemventures/go-elements/network"
)

/*
 * SwapParser
 */
type mockSwapParser struct {
	mock.Mock
}

func (m mockSwapParser) SerializeRequest(req domain.SwapRequest) ([]byte, *domain.SwapError) {
	args := m.Called(req)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}

	var err *domain.SwapError
	if a := args.Get(1); a != nil {
		err = a.(*domain.SwapError)
	}
	return res, err
}

func (m mockSwapParser) SerializeAccept(acc domain.AcceptArgs) (string, []byte, *domain.SwapError) {
	args := m.Called(acc)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	var err *domain.SwapError
	if args.Get(2) != nil {
		err = args.Get(2).(*domain.SwapError)
	}

	return sres, bres, err
}

func (m mockSwapParser) SerializeComplete(accMsg []byte, tx string) (string, []byte, *domain.SwapError) {
	args := m.Called(accMsg, tx)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	var err *domain.SwapError
	if args.Get(2) != nil {
		err = args.Get(2).(*domain.SwapError)
	}

	return sres, bres, err
}

func (m mockSwapParser) SerializeFail(id string, errCode int, errMsg string) (string, []byte) {
	args := m.Called(id, errCode, errMsg)

	var sres string
	if a := args.Get(0); a != nil {
		sres = a.(string)
	}

	var bres []byte
	if a := args.Get(1); a != nil {
		bres = a.([]byte)
	}

	return sres, bres
}

func (m mockSwapParser) DeserializeRequest(msg []byte) (domain.SwapRequest, error) {
	args := m.Called(msg)
	var res domain.SwapRequest
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapRequest)
	}

	return res, args.Error(1)
}

func (m mockSwapParser) DeserializeAccept(msg []byte) (domain.SwapAccept, error) {
	args := m.Called(msg)
	var res domain.SwapAccept
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapAccept)
	}
	return res, args.Error(1)
}

func (m mockSwapParser) DeserializeComplete(msg []byte) (domain.SwapComplete, error) {
	args := m.Called(msg)
	var res domain.SwapComplete
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapComplete)
	}
	return res, args.Error(1)
}

func (m mockSwapParser) DeserializeFail(msg []byte) (domain.SwapFail, error) {
	args := m.Called(msg)
	var res domain.SwapFail
	if a := args.Get(0); a != nil {
		res = a.(domain.SwapFail)
	}
	return res, args.Error(1)
}

/*
 * PsetParser
 */
type mockPsetParser struct {
	mock.Mock
}

func (m mockPsetParser) GetTxID(psetBase64 string) (string, error) {
	args := m.Called(psetBase64)
	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

func (m mockPsetParser) GetTxHex(psetBase64 string) (string, error) {
	args := m.Called(psetBase64)
	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

/*
 * SwapRequest
 */
type mockSwapRequest struct {
	id string
}

func newMockedSwapRequest() domain.SwapRequest {
	return &mockSwapRequest{uuid.New().String()}
}

func (m mockSwapRequest) GetId() string {
	return m.id
}

func (m mockSwapRequest) GetAssetP() string {
	return network.Regtest.AssetID
}

func (m mockSwapRequest) GetAmountP() uint64 {
	return 10000000
}

func (m mockSwapRequest) GetAssetR() string {
	return randomHex(32)
}

func (m mockSwapRequest) GetAmountR() uint64 {
	return 300000000000
}

func (m mockSwapRequest) GetTransaction() string {
	return randomBase64(100)
}

func (m mockSwapRequest) GetInputBlindingKey() map[string][]byte {
	mm := map[string][]byte{}
	mm[randomHex(20)] = randomBytes(32)
	return mm
}

func (m mockSwapRequest) GetOutputBlindingKey() map[string][]byte {
	mm := map[string][]byte{}
	mm[randomHex(20)] = randomBytes(32)
	mm[randomHex(20)] = randomBytes(32)
	return mm
}

/*
 * SwapAccept
 */
type mockSwapAccept struct {
	id string
}

func newMockedSwapAccept() domain.SwapAccept {
	return mockSwapAccept{uuid.New().String()}
}

func (m mockSwapAccept) GetId() string {
	return m.id
}

func (m mockSwapAccept) GetRequestId() string {
	return uuid.New().String()
}

func (m mockSwapAccept) GetTransaction() string {
	return randomBase64(100)
}

func (m mockSwapAccept) GetInputBlindingKey() map[string][]byte {
	mm := map[string][]byte{}
	mm[randomHex(20)] = randomBytes(32)
	mm[randomHex(20)] = randomBytes(32)
	return mm
}

func (m mockSwapAccept) GetOutputBlindingKey() map[string][]byte {
	mm := map[string][]byte{}
	mm[randomHex(20)] = randomBytes(32)
	mm[randomHex(20)] = randomBytes(32)
	mm[randomHex(20)] = randomBytes(32)
	mm[randomHex(20)] = randomBytes(32)
	return mm
}

/*
 * SwapComplete
 */
type mockSwapComplete struct {
	id string
}

func newMockedSwapComplete() domain.SwapComplete {
	return &mockSwapComplete{uuid.New().String()}
}

func (m mockSwapComplete) GetId() string {
	return m.id
}

func (m mockSwapComplete) GetAcceptId() string {
	return uuid.New().String()
}

func (m mockSwapComplete) GetTransaction() string {
	return randomBase64(100)
}
