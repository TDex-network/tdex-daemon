package storage

import (
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"sync"
)

// InMemoryUnspentRepository represents an in memory storage
type InMemoryUnspentRepository struct {
	unspents map[unspent.UnspentKey]unspent.Unspent
	lock     *sync.RWMutex
}

//NewInMemoryUnspentRepository returns a new empty InMemoryMarketRepository
func NewInMemoryUnspentRepository() *InMemoryUnspentRepository {
	return &InMemoryUnspentRepository{
		unspents: make(map[unspent.UnspentKey]unspent.Unspent),
		lock:     &sync.RWMutex{},
	}
}

func (i *InMemoryUnspentRepository) AddUnspent(unspents []unspent.Unspent) {

	//add new unspent
	for _, newUnspent := range unspents {
		if _, ok := i.unspents[newUnspent.GetKey()]; !ok {
			i.unspents[unspent.UnspentKey{
				TxID: newUnspent.TxID(),
				VOut: newUnspent.VOut(),
			}] = newUnspent
		}
	}

	//update spent
	for key, oldUnspent := range i.unspents {
		exist := false
		for _, newUnspent := range unspents {
			if newUnspent.IsKeyEqual(oldUnspent.GetKey()) {
				exist = true
			}
		}
		if !exist {
			oldUnspent.Spend()
			i.unspents[key] = oldUnspent
		}
	}
}

func (i *InMemoryUnspentRepository) GetAllUnspent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		if u.IsSpent() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func (i *InMemoryUnspentRepository) GetAllSpent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		if u.IsSpent() == true {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func (i *InMemoryUnspentRepository) GetBalance(
	address string,
	assetHash string,
) uint64 {
	var balance uint64

	for _, u := range i.unspents {
		if u.Address() == address && u.AssetHash() == assetHash && !u.IsSpent() {
			balance += u.Value()
		}
	}

	return balance
}

func (i *InMemoryUnspentRepository) GetUnlockedBalance(
	address string,
	assetHash string,
) uint64 {
	var balance uint64

	for _, u := range i.unspents {
		if u.Address() == address && u.AssetHash() == assetHash &&
			!u.IsSpent() && !u.IsLocked() {
			balance += u.Value()
		}
	}

	return balance
}

func (i *InMemoryUnspentRepository) GetAvailableUnspent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		if u.IsSpent() == false && u.IsLocked() == false {
			unspents = append(unspents, u)
		}
	}
	return unspents
}
