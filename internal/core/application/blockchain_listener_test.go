package application

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"os"
	"testing"
)

func TestUpdateUnspentsForAddress(t *testing.T) {
	dbManager, err := dbbadger.NewDbManager("testdb", nil)
	if err != nil {
		panic(err)
	}
	unspentRepository := dbbadger.NewUnspentRepositoryImpl(dbManager)
	tx := dbManager.Store.Badger().NewTransaction(true)
	ctx = context.WithValue(
		context.Background(),
		"tx",
		tx,
	)
	l := NewBlockchainListener(
		unspentRepository,
		nil,
		nil,
		nil,
		nil,
		dbManager)

	unspents := []domain.Unspent{
		{
			TxID:         "1",
			VOut:         1,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "2",
			VOut:         2,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
	}

	err = l.UpdateUnspentsForAddress(
		ctx,
		unspents,
		"a",
	)
	if err != nil {
		t.Fatal(err)
	}

	unsp, err := unspentRepository.GetUnspentsForAddresses(
		ctx,
		[]string{"a"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(unsp))

	unspents = []domain.Unspent{
		{
			TxID:         "1",
			VOut:         1,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "4",
			VOut:         2,
			Value:        0,
			AssetHash:    "",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
	}

	err = l.UpdateUnspentsForAddress(
		ctx,
		unspents,
		"a",
	)
	if err != nil {
		t.Fatal(err)
	}

	unsp, err = unspentRepository.GetUnspentsForAddresses(
		ctx,
		[]string{"a"},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(unsp))

	count := 0
	for _, v := range unsp {
		if v.IsSpent() == true {
			count++
		}
	}

	assert.Equal(t, 1, count)

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
	dbManager.Store.Close()

	err = os.RemoveAll("testdb")
	if err != nil {
		panic(err)
	}
}
