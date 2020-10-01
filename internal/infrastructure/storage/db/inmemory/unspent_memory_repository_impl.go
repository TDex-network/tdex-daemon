package inmemory

import (
	"context"
	"fmt"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
)

// UnspentRepositoryImpl represents an in memory storage
type UnspentRepositoryImpl struct {
	unspents map[domain.UnspentKey]domain.Unspent
	lock     *sync.RWMutex
}

//NewUnspentRepositoryImpl returns a new empty MarketRepositoryImpl
func NewUnspentRepositoryImpl() *UnspentRepositoryImpl {
	return &UnspentRepositoryImpl{
		unspents: map[domain.UnspentKey]domain.Unspent{},
		lock:     &sync.RWMutex{},
	}
}

//AddUnspents method is used by crawler to add unspent's to the memory,
//it assumes that all unspent's belongs to the same address,
//it assumes that each time it is invoked by crawler,
//it assumes that it will receive all unspent's for specific address
//it adds non exiting unspent's to the memory
//in case that unspent's, passed to the function, are not already in memory
//it will mark unspent in memory, as spent
func (r UnspentRepositoryImpl) AddUnspents(ctx context.Context, unspents []domain.Unspent) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return addUnspents(r.storageByContext(ctx), unspents)
}

// GetAllUnspents returns all the unspents stored
func (r UnspentRepositoryImpl) GetAllUnspents(ctx context.Context) []domain.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAllUnspents(r.storageByContext(ctx), false)
}

// GetAllSpents returns all the unspents that have been spent
func (r UnspentRepositoryImpl) GetAllSpents(ctx context.Context) []domain.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAllUnspents(r.storageByContext(ctx), true)
}

// GetBalance returns the balance of the given asset for the given address
func (r UnspentRepositoryImpl) GetBalance(
	ctx context.Context,
	address, assetHash string,
) uint64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getBalance(r.storageByContext(ctx), address, assetHash)
}

// GetUnlockedBalance returns the total amount of unlocked unspents for the
// given asset and address
func (r UnspentRepositoryImpl) GetUnlockedBalance(
	ctx context.Context,
	address, assetHash string,
) uint64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getUnlockedBalance(r.storageByContext(ctx), address, assetHash)
}

// GetAvailableUnspents returns the list of unlocked unspents
func (r UnspentRepositoryImpl) GetAvailableUnspents(ctx context.Context) []domain.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAvailableUnspents(r.storageByContext(ctx), nil)
}

// GetAvailableUnspentsForAddresses returns the list of unlocked unspents for
// the given list of addresses
func (r UnspentRepositoryImpl) GetAvailableUnspentsForAddresses(
	ctx context.Context,
	addresses []string,
) []domain.Unspent {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return getAvailableUnspents(r.storageByContext(ctx), addresses)
}

// LockUnspents locks the given unspents associating them with the trade where
// they'are currently used as inputs
func (r UnspentRepositoryImpl) LockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return lockUnspents(r.storageByContext(ctx), unspentKeys, tradeID)
}

// UnlockUnspents unlocks the given locked unspents
func (r UnspentRepositoryImpl) UnlockUnspents(
	ctx context.Context,
	unspentKeys []domain.UnspentKey,
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return unlockUnspents(r.storageByContext(ctx), unspentKeys)
}

func (r UnspentRepositoryImpl) GetBalanceInfoForAsset(
	unspents []domain.Unspent,
) map[string]domain.BalanceInfo {
	balances := map[string]domain.BalanceInfo{}
	for _, unspent := range unspents {
		if _, ok := balances[unspent.AssetHash()]; !ok {
			balances[unspent.AssetHash()] = domain.BalanceInfo{}
		}

		balance := balances[unspent.AssetHash()]
		balance.TotalBalance += unspent.Value()
		if unspent.IsConfirmed() {
			balance.ConfirmedBalance += unspent.Value()
		} else {
			balance.UnconfirmedBalance += unspent.Value()
		}
		balances[unspent.AssetHash()] = balance
	}
	return balances
}

// Begin returns a new UnspentRepositoryTxImpl
func (r UnspentRepositoryImpl) Begin() (uow.Tx, error) {
	tx := &UnspentRepositoryTxImpl{
		root:     r,
		unspents: map[domain.UnspentKey]domain.Unspent{},
	}

	// copy the current state of the repo into the transaction
	for k, v := range r.unspents {
		tx.unspents[k] = v
	}
	return tx, nil
}

// ContextKey returns the context key shared between in-memory repositories
func (r UnspentRepositoryImpl) ContextKey() interface{} {
	return uow.InMemoryContextKey
}

func (r UnspentRepositoryImpl) storageByContext(ctx context.Context) (
	unspents map[domain.UnspentKey]domain.Unspent,
) {
	unspents = r.unspents
	if tx, ok := ctx.Value(r).(*UnspentRepositoryTxImpl); ok {
		unspents = tx.unspents
	}
	return
}

func addUnspents(storage map[domain.UnspentKey]domain.Unspent, unspents []domain.Unspent) error {
	addr := unspents[0].Address()
	for _, u := range unspents {
		if u.Address() != addr {
			return fmt.Errorf("all unspent's must belong to the same address")
		}
	}

	//add new unspent
	for _, newUnspent := range unspents {
		if _, ok := storage[newUnspent.GetKey()]; !ok {
			storage[domain.UnspentKey{
				TxID: newUnspent.TxID(),
				VOut: newUnspent.VOut(),
			}] = newUnspent
		}
	}

	//update spent
	for key, oldUnspent := range storage {
		if oldUnspent.Address() == addr {
			exist := false
			for _, newUnspent := range unspents {
				if newUnspent.IsKeyEqual(oldUnspent.GetKey()) {
					exist = true
				}
			}
			if !exist {
				oldUnspent.Spend()
				storage[key] = oldUnspent
			}
		}
	}

	return nil
}

func getAllUnspents(storage map[domain.UnspentKey]domain.Unspent, spent bool) []domain.Unspent {
	unspents := make([]domain.Unspent, 0)
	for _, u := range storage {
		if u.IsSpent() == spent {
			unspents = append(unspents, u)
		}
	}
	return unspents
}

func getBalance(storage map[domain.UnspentKey]domain.Unspent, address, assetHash string) uint64 {
	var balance uint64
	for _, u := range storage {
		if u.Address() == address && u.AssetHash() == assetHash && !u.IsSpent() && u.IsConfirmed() {
			balance += u.Value()
		}
	}
	return balance
}

func getUnlockedBalance(storage map[domain.UnspentKey]domain.Unspent, address, assetHash string) uint64 {
	var balance uint64
	for _, u := range storage {
		if u.Address() == address && u.AssetHash() == assetHash &&
			!u.IsSpent() && !u.IsLocked() && u.IsConfirmed() {
			balance += u.Value()
		}
	}
	return balance
}

func getAvailableUnspents(storage map[domain.UnspentKey]domain.Unspent, addresses []string) []domain.Unspent {
	unspents := make([]domain.Unspent, 0)
	for _, u := range storage {
		if !u.IsSpent() && !u.IsLocked() && u.IsConfirmed() {
			if len(addresses) == 0 {
				unspents = append(unspents, u)
			} else {
				for _, addr := range addresses {
					if addr == u.Address() {
						unspents = append(unspents, u)
						break
					}
				}
			}
		}
	}
	return unspents
}

func lockUnspents(
	storage map[domain.UnspentKey]domain.Unspent,
	unspentKeys []domain.UnspentKey,
	tradeID uuid.UUID,
) error {
	for _, key := range unspentKeys {
		if _, ok := storage[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := storage[key]
		unspent.Lock(&tradeID)
		storage[key] = unspent
	}
	return nil
}

func unlockUnspents(
	storage map[domain.UnspentKey]domain.Unspent,
	unspentKeys []domain.UnspentKey,
) error {
	for _, key := range unspentKeys {
		if _, ok := storage[key]; !ok {
			return fmt.Errorf("unspent not found for key %v", key)
		}
	}

	for _, key := range unspentKeys {
		unspent := storage[key]
		unspent.UnLock()
		storage[key] = unspent
	}
	return nil
}

// UnspentRepositoryTxImpl allows to make transactional read/write operation
// on the in-memory repository
type UnspentRepositoryTxImpl struct {
	root     UnspentRepositoryImpl
	unspents map[domain.UnspentKey]domain.Unspent
}

// Commit applies the updates made to the state of the transaction to its root
func (tx *UnspentRepositoryTxImpl) Commit() error {
	for k, v := range tx.unspents {
		tx.root.unspents[k] = v
	}
	return nil
}

// Rollback resets the state of the transaction to the state of its root
func (tx *UnspentRepositoryTxImpl) Rollback() error {
	tx.unspents = map[domain.UnspentKey]domain.Unspent{}
	for k, v := range tx.root.unspents {
		tx.unspents[k] = v
	}
	return nil
}
