// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

//go:build !live

package rpcserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"decred.org/dcrdex/client/asset"
	"decred.org/dcrdex/client/core"
	"decred.org/dcrdex/client/db"
	"decred.org/dcrdex/client/mnemonic"
	"decred.org/dcrdex/client/orderbook"
	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/msgjson"
)

func init() {
	log = dex.StdOutLogger("TEST", dex.LevelTrace)
}

var (
	tCtx context.Context
)

type TCore struct {
	dexExchange              *core.Exchange
	getDEXConfigErr          error
	balanceErr               error
	syncErr                  error
	createWalletErr          error
	newWalletForm            *core.WalletForm
	openWalletErr            error
	rescanWalletErr          error
	walletState              *core.WalletState
	closeWalletErr           error
	walletStatusErr          error
	wallets                  []*core.WalletState
	initializeClientErr      error
	postBondResult           *core.PostBondResult
	postBondErr              error
	bondOptsErr              error
	exchanges                map[string]*core.Exchange
	loginErr                 error
	order                    *core.Order
	tradeErr                 error
	cancelErr                error
	coin                     asset.Coin
	sendErr                  error
	logoutErr                error
	book                     *core.OrderBook
	bookErr                  error
	exportSeed               string
	exportSeedErr            error
	discoverAcctErr          error
	archivedRecords          int
	deleteArchivedRecordsErr error
	setVSPErr                error
	purchaseTicketsErr       error
	stakeStatus              *asset.TicketStakingStatus
	stakeStatusErr           error
	setVotingPrefErr         error
}

func (c *TCore) Balance(uint32) (uint64, error) {
	return 0, c.balanceErr
}
func (c *TCore) Book(dex string, base, quote uint32) (*core.OrderBook, error) {
	return c.book, c.bookErr
}
func (c *TCore) AckNotes(ids []dex.Bytes) {}
func (c *TCore) AssetBalance(uint32) (*core.WalletBalance, error) {
	return nil, c.balanceErr
}
func (c *TCore) Cancel(oid dex.Bytes) error {
	return c.cancelErr
}
func (c *TCore) CreateWallet(appPW, walletPW []byte, form *core.WalletForm) error {
	c.newWalletForm = form
	return c.createWalletErr
}
func (c *TCore) CloseWallet(assetID uint32) error {
	return c.closeWalletErr
}
func (c *TCore) Exchanges() (exchanges map[string]*core.Exchange) { return c.exchanges }
func (c *TCore) Exchange(host string) (*core.Exchange, error) {
	exchange, ok := c.exchanges[host]
	if !ok {
		return nil, fmt.Errorf("no exchange at %v", host)
	}
	return exchange, nil
}
func (c *TCore) InitializeClient(pw []byte, seed *string) (string, error) {
	var mnemonicSeed string
	if seed == nil {
		_, mnemonicSeed = mnemonic.New()
	}
	return mnemonicSeed, c.initializeClientErr
}
func (c *TCore) Login(appPass []byte) error {
	return c.loginErr
}
func (c *TCore) Logout() error {
	return c.logoutErr
}
func (c *TCore) OpenWallet(assetID uint32, pw []byte) error {
	return c.openWalletErr
}
func (c *TCore) ToggleWalletStatus(assetID uint32, disable bool) error {
	if c.walletStatusErr != nil {
		return c.walletStatusErr
	}
	if c.walletState != nil {
		c.walletState.Disabled = disable
	}
	return c.walletStatusErr
}
func (c *TCore) RescanWallet(assetID uint32, force bool) error {
	return c.rescanWalletErr
}
func (c *TCore) GetDEXConfig(dexAddr string, certI any) (*core.Exchange, error) {
	return c.dexExchange, c.getDEXConfigErr
}
func (c *TCore) PostBond(*core.PostBondForm) (*core.PostBondResult, error) {
	return c.postBondResult, c.postBondErr
}
func (c *TCore) UpdateBondOptions(form *core.BondOptionsForm) error {
	return c.bondOptsErr
}
func (c *TCore) SyncBook(dex string, base, quote uint32) (*orderbook.OrderBook, core.BookFeed, error) {
	return nil, &tBookFeed{}, c.syncErr
}
func (c *TCore) Trade(appPass []byte, form *core.TradeForm) (order *core.Order, err error) {
	return c.order, c.tradeErr
}
func (c *TCore) Wallets() []*core.WalletState {
	return c.wallets
}
func (c *TCore) WalletState(assetID uint32) *core.WalletState {
	return c.walletState
}
func (c *TCore) Send(pw []byte, assetID uint32, value uint64, addr string, subtract bool) (asset.Coin, error) {
	return c.coin, c.sendErr
}
func (c *TCore) ExportSeed(pw []byte) (string, error) {
	return c.exportSeed, c.exportSeedErr
}
func (c *TCore) DiscoverAccount(dexAddr string, pass []byte, certI any) (*core.Exchange, bool, error) {
	return c.dexExchange, false, c.discoverAcctErr
}
func (c *TCore) DeleteArchivedRecords(olderThan *time.Time, matchesFileStr, ordersFileStr string) (int, error) {
	return c.archivedRecords, c.deleteArchivedRecordsErr
}
func (c *TCore) AssetHasActiveOrders(uint32) bool {
	return false
}
func (c *TCore) WalletPeers(assetID uint32) ([]*asset.WalletPeer, error) {
	return nil, nil
}
func (c *TCore) AddWalletPeer(assetID uint32, address string) error {
	return nil
}
func (c *TCore) RemoveWalletPeer(assetID uint32, address string) error {
	return nil
}
func (c *TCore) Notifications(n int) (notes, pokes []*db.Notification, _ error) {
	return nil, nil, nil
}
func (c *TCore) MultiTrade(appPass []byte, form *core.MultiTradeForm) []*core.MultiTradeResult {
	return nil
}
func (c *TCore) SetVSP(assetID uint32, addr string) error {
	return c.setVSPErr
}
func (c *TCore) PurchaseTickets(assetID uint32, pw []byte, n int) error {
	return c.purchaseTicketsErr
}
func (c *TCore) StakeStatus(assetID uint32) (*asset.TicketStakingStatus, error) {
	return c.stakeStatus, c.stakeStatusErr
}
func (c *TCore) SetVotingPreferences(assetID uint32, choices, tSpendPolicy, treasuryPolicy map[string]string) error {
	return c.setVotingPrefErr
}
func (c *TCore) TxHistory(assetID uint32, n int, refID *string, past bool) ([]*asset.WalletTransaction, error) {
	return nil, nil
}
func (c *TCore) WalletTransaction(assetID uint32, txID string) (*asset.WalletTransaction, error) {
	return nil, nil
}
func (c *TCore) GenerateBCHRecoveryTransaction(appPW []byte, recipient string) ([]byte, error) {
	return nil, nil
}
func (c *TCore) BridgeContractApprovalStatus(assetID uint32) (asset.ApprovalStatus, error) {
	return 0, nil
}
func (c *TCore) ApproveBridgeContract(assetID uint32) (string, error) {
	return "", nil
}
func (c *TCore) UnapproveBridgeContract(assetID uint32) (string, error) {
	return "", nil
}
func (c *TCore) Bridge(fromAssetID, toAssetID uint32, amt uint64) (txID string, err error) {
	return "", nil
}
func (c *TCore) BridgeHistory(fromAssetID uint32, n int, refID *string, past bool) ([]*asset.WalletTransaction, error) {
	return nil, nil
}
func (c *TCore) PendingBridges(fromAssetID uint32) ([]*asset.WalletTransaction, error) {
	return nil, nil
}

type tBookFeed struct{}

func (*tBookFeed) Next() <-chan *core.BookUpdate {
	return make(<-chan *core.BookUpdate)
}
func (*tBookFeed) Close() {}
func (*tBookFeed) Candles(dur string) error {
	return nil
}

func newTServer(t *testing.T, start bool, user, pass string) (*RPCServer, func()) {
	tSrv, fn, err := newTServerWErr(t, start, user, pass)
	if err != nil {
		t.Fatal(err)
	}
	return tSrv, fn
}
func newTServerWErr(t *testing.T, start bool, user, pass string) (*RPCServer, func(), error) {
	t.Helper()

	var shutdown func()
	ctx, killCtx := context.WithCancel(tCtx)
	tempDir := t.TempDir()

	cert, key := tempDir+"/cert.cert", tempDir+"/key.key"
	cfg := &Config{
		Core: &TCore{},
		Addr: "127.0.0.1:0",
		User: user,
		Pass: pass,
		Cert: cert,
		Key:  key,
	}
	s, err := New(cfg)
	if err != nil {
		killCtx()
		return nil, nil, fmt.Errorf("error creating server: %w", err)
	}
	if start {
		cm := dex.NewConnectionMaster(s)
		err := cm.Connect(ctx)
		if err != nil {
			killCtx()
			return nil, nil, fmt.Errorf("error starting RPCServer: %w", err)
		}
		shutdown = func() {
			killCtx()
			cm.Disconnect()
		}
	} else {
		shutdown = killCtx
	}
	return s, shutdown, nil
}

func TestMain(m *testing.M) {
	var shutdown func()
	tCtx, shutdown = context.WithCancel(context.Background())
	doIt := func() int {
		defer shutdown()
		return m.Run()
	}
	os.Exit(doIt())
}

func TestConnectBindError(t *testing.T) {
	s0, shutdown := newTServer(t, true, "", "abc")
	defer shutdown()

	tempDir := t.TempDir()

	cert, key := tempDir+"/cert.cert", tempDir+"/key.key"
	cfg := &Config{
		Core: &TCore{},
		Addr: s0.addr,
		User: "",
		Pass: "abc",
		Cert: cert,
		Key:  key,
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}

	cm := dex.NewConnectionMaster(s)
	if err = cm.Connect(tCtx); err == nil {
		shutdown() // shutdown both servers with shared context
		cm.Disconnect()
		t.Fatal("should have failed to bind")
	}
}

type tResponseWriter struct {
	b    []byte
	code int
}

func (w *tResponseWriter) Header() http.Header {
	return make(http.Header)
}
func (w *tResponseWriter) Write(msg []byte) (int, error) {
	w.b = msg
	return len(msg), nil
}
func (w *tResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func TestParseHTTPRequest(t *testing.T) {
	s, shutdown := newTServer(t, false, "", "abc")
	defer shutdown()
	var r *http.Request

	ensureHTTPError := func(name string, wantCode int) {
		t.Helper()
		w := &tResponseWriter{}
		s.handleJSON(w, r)
		if w.code != wantCode {
			t.Fatalf("%s: Expected HTTP error %d, got %d",
				name, wantCode, w.code)
		}
	}

	ensureMsgErr := func(name string, wantCode int) {
		t.Helper()
		w := &tResponseWriter{}
		s.handleJSON(w, r)
		if w.code != 200 {
			t.Fatalf("HTTP error when expecting msgjson.Error")
		}
		resp := new(msgjson.Message)
		if err := json.Unmarshal(w.b, resp); err != nil {
			t.Fatalf("unable to unmarshal response: %v", err)
		}
		payload := new(msgjson.ResponsePayload)
		if err := json.Unmarshal(resp.Payload, payload); err != nil {
			t.Fatalf("unable to unmarshal payload: %v", err)
		}
		if payload.Error == nil {
			t.Fatalf("%s: no error", name)
		}
		if wantCode != payload.Error.Code {
			t.Fatalf("%s, wanted %d, got %d",
				name, wantCode, payload.Error.Code)
		}
	}
	ensureNoErr := func(name string) {
		t.Helper()
		w := &tResponseWriter{}
		s.handleJSON(w, r)
		if w.code != 200 {
			t.Fatalf("HTTP error when expecting no error")
		}
		resp := new(msgjson.Message)
		if err := json.Unmarshal(w.b, resp); err != nil {
			t.Fatalf("unable to unmarshal response: %v", err)
		}
		payload := new(msgjson.ResponsePayload)
		if err := json.Unmarshal(resp.Payload, payload); err != nil {
			t.Fatalf("unable to unmarshal payload: %v", err)
		}
		if payload.Error != nil {
			t.Fatalf("%s: errored", name)
		}
	}

	// Send a response, which is unsupported on the server.
	msg, _ := msgjson.NewResponse(1, nil, nil)
	b, _ := json.Marshal(msg)
	bbuff := bytes.NewBuffer(b)
	r, _ = http.NewRequest("GET", "", bbuff)
	ensureHTTPError("response", http.StatusMethodNotAllowed)

	// Unknown route.
	msg, _ = msgjson.NewRequest(1, "123", nil)
	b, _ = json.Marshal(msg)
	bbuff = bytes.NewBuffer(b)
	r, _ = http.NewRequest("GET", "", bbuff)
	ensureMsgErr("bad route", msgjson.RPCUnknownRoute)

	// Use real route.
	msg, _ = msgjson.NewRequest(1, "version", nil)
	b, _ = json.Marshal(msg)
	bbuff = bytes.NewBuffer(b)
	r, _ = http.NewRequest("GET", "", bbuff)
	ensureNoErr("good request")

	// Use real route with bad args.
	msg, _ = msgjson.NewRequest(1, "version", "something")
	b, _ = json.Marshal(msg)
	bbuff = bytes.NewBuffer(b)
	r, _ = http.NewRequest("GET", "", bbuff)
	ensureMsgErr("bad params", msgjson.RPCParseError)
}

func TestNew(t *testing.T) {
	authTests := []struct {
		name, user, pass, wantAuth string
		wantErr                    bool
	}{{
		name:     "ok",
		user:     "user",
		pass:     "pass",
		wantAuth: "AK+rg3mIGeouojwZwNRMjBjZouASr4mu4FWMTXQQcD0=",
	}, {
		name:     "ok various input",
		user:     `&!"#$%&'()~=`,
		pass:     `+<>*?,:.;/][{}`,
		wantAuth: "Te4g4+Ke9Q07MYo3iT1OCqq5qXX2ZcB47FBiVaT41hQ=",
	}, {
		name:    "no password",
		user:    "user",
		wantErr: true,
	}}
	for _, test := range authTests {
		s, shutdown, err := newTServerWErr(t, false, test.user, test.pass)
		if test.wantErr {
			if err == nil {
				t.Fatalf("expected error for test %s", test.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error for test %s: %v", test.name, err)
		}
		auth := base64.StdEncoding.EncodeToString((s.authSHA[:]))
		if auth != test.wantAuth {
			t.Fatalf("expected auth %s but got %s", test.wantAuth, auth)
		}
		shutdown()
	}
}

func TestAuthMiddleware(t *testing.T) {
	s, shutdown := newTServer(t, false, "", "abc")
	defer shutdown()
	am := s.authMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	r, _ := http.NewRequest("GET", "", nil)

	wantAuthError := func(name string, want bool) {
		t.Helper()
		w := &tResponseWriter{}
		am.ServeHTTP(w, r)
		if w.code != http.StatusUnauthorized && w.code != http.StatusOK {
			t.Fatalf("unexpected HTTP error %d for test \"%s\"",
				w.code, name)
		}
		switch want {
		case true:
			if w.code != http.StatusUnauthorized {
				t.Fatalf("Expected unauthorized HTTP error for test \"%s\"",
					name)
			}
		case false:
			if w.code != http.StatusOK {
				t.Fatalf("Expected OK HTTP status for test \"%s\"",
					name)
			}
		}
	}

	user, pass := "Which one is it?", "It's the one that says bmf on it."
	login := user + ":" + pass
	h := "Basic "
	auth := h + base64.StdEncoding.EncodeToString([]byte(login))
	s.authSHA = sha256.Sum256([]byte(auth))

	tests := []struct {
		name, user, pass, header string
		hasAuth, wantErr         bool
	}{{
		name:    "auth ok",
		user:    user,
		pass:    pass,
		header:  h,
		hasAuth: true,
		wantErr: false,
	}, {
		name:    "wrong pass",
		user:    user,
		pass:    "password123",
		header:  h,
		hasAuth: true,
		wantErr: true,
	}, {
		name:    "unknown user",
		user:    "Jules",
		pass:    pass,
		header:  h,
		hasAuth: true,
		wantErr: true,
	}, {
		name:    "no header",
		user:    user,
		pass:    pass,
		header:  h,
		hasAuth: false,
		wantErr: true,
	}, {
		name:    "malformed header",
		user:    user,
		pass:    pass,
		header:  "basic ",
		hasAuth: true,
		wantErr: true,
	}}
	for _, test := range tests {
		login = test.user + ":" + test.pass
		auth = test.header + base64.StdEncoding.EncodeToString([]byte(login))
		requestHeader := make(http.Header)
		if test.hasAuth {
			requestHeader.Add("Authorization", auth)
		}
		r.Header = requestHeader
		wantAuthError(test.name, test.wantErr)
	}
}
