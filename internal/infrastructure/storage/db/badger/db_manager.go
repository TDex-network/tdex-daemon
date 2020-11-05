package dbbadger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold/v2"
)

// DbManager holds all the badgerhold stores in a single data structure.
type DbManager struct {
	Store        *badgerhold.Store
	PriceStore   *badgerhold.Store
	UnspentStore *badgerhold.Store
}

// NewDbManager opens (or creates if not exists) the badger store on disk. It expects a base data dir and an optional logger.
// It creates a dedicated directory for main, price and unspent.
func NewDbManager(baseDbDir string, logger badger.Logger) (*DbManager, error) {
	mainDb, err := createDb(filepath.Join(baseDbDir, "main"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}

	priceDb, err := createDb(filepath.Join(baseDbDir, "prices"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening prices db: %w", err)
	}

	unspentDb, err := createDb(filepath.Join(baseDbDir, "unspents"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	return &DbManager{
		Store:        mainDb,
		PriceStore:   priceDb,
		UnspentStore: unspentDb,
	}, nil
}

// NewTransaction implements the DbManager interface
func (d DbManager) NewTransaction() ports.Transaction {
	return d.Store.Badger().NewTransaction(true)
}

// NewPricesTransaction implements the DbManager interface
func (d DbManager) NewPricesTransaction() ports.Transaction {
	return d.PriceStore.Badger().NewTransaction(true)
}

// NewUnspentsTransaction implements the DbManager interface
func (d DbManager) NewUnspentsTransaction() ports.Transaction {
	return d.UnspentStore.Badger().NewTransaction(true)
}

// IsTransactionConflict returns wheter the error occured when commiting a
// transacton is a conflict
func (d DbManager) IsTransactionConflict(err error) bool {
	return err == badger.ErrConflict
}

// JSONEncode is a custom JSON based encoder for badger
func JSONEncode(value interface{}) ([]byte, error) {
	var buff bytes.Buffer

	en := json.NewEncoder(&buff)

	err := en.Encode(value)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// JSONDecode is a custom JSON based decoder for badger
func JSONDecode(data []byte, value interface{}) error {
	var buff bytes.Buffer
	de := json.NewDecoder(&buff)

	_, err := buff.Write(data)
	if err != nil {
		return err
	}

	return de.Decode(value)
}

func createDb(dbDir string, logger badger.Logger) (db *badgerhold.Store, err error) {
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger

	db, err = badgerhold.Open(badgerhold.Options{
		Encoder:          JSONEncode,
		Decoder:          JSONDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})

	return
}
