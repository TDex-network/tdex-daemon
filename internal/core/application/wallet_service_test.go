package application

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

func TestNewWalletService(t *testing.T) {
	ws, _, close := newTestWallet(nil)
	defer close()
	assert.Equal(t, false, ws.walletInitialized)
	assert.Equal(t, false, ws.walletIsSyncing)
}

func TestGenSeed(t *testing.T) {
	walletSvc, ctx, close := newTestWallet(nil)
	defer close()

	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 24, len(seed))
}

func TestInitWalletWrongSeed(t *testing.T) {
	walletSvc, ctx, close := newTestWallet(nil)
	defer close()

	wrongSeed := []string{"test"}
	err := walletSvc.InitWallet(ctx, wrongSeed, "pass")
	assert.Error(t, err)
}

func TestInitEmptyWallet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	walletSvc, ctx, close := newTestWallet(emptyWallet)
	defer close()
	// If the vault repository is not empty when the wallet service is
	// instantiated, this behaves like it  it was shut down and restarted again.
	// Therefore, the service restores its previous state and "marks" the wallet
	// as initialized by setting the internal walletInitialized bool field to
	// true. InitWallet, on its side, does not perform any operation if the
	// wallet looks already initialized.
	// In this test and in the next one, the walletInitialized field is manually
	// set to false because a mocked Vault repository is used that would cause
	// the bool field to be set to true when at service instantiation.
	walletSvc.walletInitialized = false

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: emptyWallet.mnemonic,
	})
	firstWalletAccountAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/0",
		Network:        config.GetNetwork(),
	})

	err := walletSvc.InitWallet(ctx, emptyWallet.mnemonic, emptyWallet.password)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, walletSvc.walletInitialized)

	if err := walletSvc.UnlockWallet(ctx, emptyWallet.password); err != nil {
		t.Fatal(err)
	}
	addr, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, firstWalletAccountAddr, addr)
}

func TestInitUsedWallet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	walletSvc, ctx, close := newTestWallet(usedWallet)
	defer close()
	walletSvc.walletInitialized = false

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: usedWallet.mnemonic,
	})
	mockedLastDerivedAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/15",
		Network:        config.GetNetwork(),
	})
	if _, err := walletSvc.explorerService.Faucet(mockedLastDerivedAddr); err != nil {
		t.Fatal(err)
	}
	firstWalletAccountAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/16",
		Network:        config.GetNetwork(),
	})

	err := walletSvc.InitWallet(ctx, usedWallet.mnemonic, usedWallet.password)
	if err != nil {
		t.Fatal(err)
	}
	if err := walletSvc.UnlockWallet(ctx, usedWallet.password); err != nil {
		t.Fatal(err)
	}
	addr, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, firstWalletAccountAddr, addr)
}

func TestWalletUnlock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	walletSvc, ctx, close := newTestWallet(dryLockedWallet)
	defer close()

	address, blindingKey, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	assert.Equal(t, domain.ErrMustBeUnlocked, err)

	err = walletSvc.UnlockWallet(ctx, dryLockedWallet.password)
	if err != nil {
		t.Fatal(err)
	}

	address, blindingKey, err = walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, len(address) > 0)
	assert.Equal(t, true, len(blindingKey) > 0)
}

func TestWalletChangePass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	walletSvc, ctx, close := newTestWallet(dryLockedWallet)
	defer close()

	err := walletSvc.ChangePassword(ctx, "wrongPass", "newPass")
	assert.Equal(t, domain.ErrInvalidPassphrase, err)

	err = walletSvc.ChangePassword(ctx, dryLockedWallet.password, "newPass")
	assert.NoError(t, err)

	err = walletSvc.UnlockWallet(ctx, dryLockedWallet.password)
	assert.Equal(t, wallet.ErrInvalidPassphrase, err)
}

func TestWalletBalance(t *testing.T) {
	walletSvc, ctx, close := newTestWallet(dryWallet)
	defer close()

	address, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = walletSvc.explorerService.Faucet(address)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	balance, err := walletSvc.GetWalletBalance(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		true,
		int(balance[config.GetString(config.BaseAssetKey)].ConfirmedBalance) >= 100000000,
	)
}
