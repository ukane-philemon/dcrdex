// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package market

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/calc"
	"decred.org/dcrdex/dex/candles"
	"decred.org/dcrdex/dex/msgjson"
	"decred.org/dcrdex/dex/order"
	"decred.org/dcrdex/dex/order/test"
	"decred.org/dcrdex/server/account"
	"decred.org/dcrdex/server/asset"
	"decred.org/dcrdex/server/coinlock"
	"decred.org/dcrdex/server/db"
	"decred.org/dcrdex/server/matcher"
	"decred.org/dcrdex/server/swap"
)

type TArchivist struct {
	mtx                  sync.Mutex
	poisonEpochOrder     order.Order
	orderWithKnownCommit order.OrderID
	commitForKnownOrder  order.Commitment
	bookedOrders         []*order.LimitOrder
	canceledOrders       []*order.LimitOrder
	archivedCancels      []*order.CancelOrder
	epochInserted        chan struct{}
	revoked              order.Order
}

func (ta *TArchivist) Close() error           { return nil }
func (ta *TArchivist) LastErr() error         { return nil }
func (ta *TArchivist) Fatal() <-chan struct{} { return nil }
func (ta *TArchivist) Order(oid order.OrderID, base, quote uint32) (order.Order, order.OrderStatus, error) {
	return nil, order.OrderStatusUnknown, errors.New("boom")
}
func (ta *TArchivist) BookOrders(base, quote uint32) ([]*order.LimitOrder, error) {
	ta.mtx.Lock()
	defer ta.mtx.Unlock()
	return ta.bookedOrders, nil
}
func (ta *TArchivist) EpochOrders(base, quote uint32) ([]order.Order, error) {
	return nil, nil
}
func (ta *TArchivist) MarketMatches(base, quote uint32) ([]*db.MatchDataWithCoins, error) {
	return nil, nil
}
func (ta *TArchivist) FlushBook(base, quote uint32) (sells, buys []order.OrderID, err error) {
	ta.mtx.Lock()
	defer ta.mtx.Unlock()
	for _, lo := range ta.bookedOrders {
		if lo.Sell {
			sells = append(sells, lo.ID())
		} else {
			buys = append(buys, lo.ID())
		}
	}
	ta.bookedOrders = nil
	return
}
func (ta *TArchivist) NewArchivedCancel(ord *order.CancelOrder, epochID, epochDur int64) error {
	if ta.archivedCancels != nil {
		ta.archivedCancels = append(ta.archivedCancels, ord)
	}
	return nil
}
func (ta *TArchivist) ActiveOrderCoins(base, quote uint32) (baseCoins, quoteCoins map[order.OrderID][]order.CoinID, err error) {
	return make(map[order.OrderID][]order.CoinID), make(map[order.OrderID][]order.CoinID), nil
}
func (ta *TArchivist) UserOrders(ctx context.Context, aid account.AccountID, base, quote uint32) ([]order.Order, []order.OrderStatus, error) {
	return nil, nil, errors.New("boom")
}
func (ta *TArchivist) UserOrderStatuses(aid account.AccountID, base, quote uint32, oids []order.OrderID) ([]*db.OrderStatus, error) {
	return nil, errors.New("boom")
}
func (ta *TArchivist) ActiveUserOrderStatuses(aid account.AccountID) ([]*db.OrderStatus, error) {
	return nil, errors.New("boom")
}
func (ta *TArchivist) OrderWithCommit(ctx context.Context, commit order.Commitment) (found bool, oid order.OrderID, err error) {
	ta.mtx.Lock()
	defer ta.mtx.Unlock()
	if commit == ta.commitForKnownOrder {
		return true, ta.orderWithKnownCommit, nil
	}
	return
}
func (ta *TArchivist) CompletedUserOrders(aid account.AccountID, N int) (oids []order.OrderID, compTimes []int64, err error) {
	return nil, nil, nil
}
func (ta *TArchivist) ExecutedCancelsForUser(aid account.AccountID, N int) ([]*db.CancelRecord, error) {
	return nil, nil
}
func (ta *TArchivist) OrderStatus(order.Order) (order.OrderStatus, order.OrderType, int64, error) {
	return order.OrderStatusUnknown, order.UnknownOrderType, -1, errors.New("boom")
}
func (ta *TArchivist) NewEpochOrder(ord order.Order, epochIdx, epochDur int64, epochGap int32) error {
	ta.mtx.Lock()
	defer ta.mtx.Unlock()
	if ta.poisonEpochOrder != nil && ord.ID() == ta.poisonEpochOrder.ID() {
		return errors.New("barf")
	}
	return nil
}
func (ta *TArchivist) StorePreimage(ord order.Order, pi order.Preimage) error { return nil }
func (ta *TArchivist) failOnEpochOrder(ord order.Order) {
	ta.mtx.Lock()
	ta.poisonEpochOrder = ord
	ta.mtx.Unlock()
}
func (ta *TArchivist) InsertEpoch(ed *db.EpochResults) error {
	if ta.epochInserted != nil { // the test wants to know
		ta.epochInserted <- struct{}{}
	}
	return nil
}
func (ta *TArchivist) LastEpochRate(base, quote uint32) (rate uint64, err error) {
	return 1, nil
}
func (ta *TArchivist) BookOrder(lo *order.LimitOrder) error {
	ta.mtx.Lock()
	defer ta.mtx.Unlock()
	// Note that the other storage functions like ExecuteOrder and CancelOrder
	// do not change this order slice.
	ta.bookedOrders = append(ta.bookedOrders, lo)
	return nil
}
func (ta *TArchivist) ExecuteOrder(ord order.Order) error { return nil }
func (ta *TArchivist) CancelOrder(lo *order.LimitOrder) error {
	if ta.canceledOrders != nil {
		ta.canceledOrders = append(ta.canceledOrders, lo)
	}
	return nil
}
func (ta *TArchivist) RevokeOrder(ord order.Order) (order.OrderID, time.Time, error) {
	ta.revoked = ord
	return ord.ID(), time.Now(), nil
}
func (ta *TArchivist) RevokeOrderUncounted(order.Order) (order.OrderID, time.Time, error) {
	return order.OrderID{}, time.Now(), nil
}
func (ta *TArchivist) SetOrderCompleteTime(ord order.Order, compTime int64) error { return nil }
func (ta *TArchivist) FailCancelOrder(*order.CancelOrder) error                   { return nil }
func (ta *TArchivist) UpdateOrderFilled(*order.LimitOrder) error                  { return nil }
func (ta *TArchivist) UpdateOrderStatus(order.Order, order.OrderStatus) error     { return nil }

// SwapArchiver for Swapper
func (ta *TArchivist) ActiveSwaps() ([]*db.SwapDataFull, error) { return nil, nil }
func (ta *TArchivist) InsertMatch(match *order.Match) error     { return nil }
func (ta *TArchivist) MatchByID(mid order.MatchID, base, quote uint32) (*db.MatchData, error) {
	return nil, nil
}
func (ta *TArchivist) UserMatches(aid account.AccountID, base, quote uint32) ([]*db.MatchData, error) {
	return nil, nil
}
func (ta *TArchivist) CompletedAndAtFaultMatchStats(aid account.AccountID, lastN int) ([]*db.MatchOutcome, error) {
	return nil, nil
}
func (ta *TArchivist) PreimageStats(user account.AccountID, lastN int) ([]*db.PreimageResult, error) {
	return nil, nil
}
func (ta *TArchivist) ForgiveMatchFail(order.MatchID) (bool, error) { return false, nil }
func (ta *TArchivist) AllActiveUserMatches(account.AccountID) ([]*db.MatchData, error) {
	return nil, nil
}
func (ta *TArchivist) MatchStatuses(aid account.AccountID, base, quote uint32, matchIDs []order.MatchID) ([]*db.MatchStatus, error) {
	return nil, nil
}
func (ta *TArchivist) SwapData(mid db.MarketMatchID) (order.MatchStatus, *db.SwapData, error) {
	return 0, nil, nil
}
func (ta *TArchivist) SaveMatchAckSigA(mid db.MarketMatchID, sig []byte) error { return nil }
func (ta *TArchivist) SaveMatchAckSigB(mid db.MarketMatchID, sig []byte) error { return nil }

// Contract data.
func (ta *TArchivist) SaveContractA(mid db.MarketMatchID, contract []byte, coinID []byte, timestamp int64) error {
	return nil
}
func (ta *TArchivist) SaveAuditAckSigB(mid db.MarketMatchID, sig []byte) error { return nil }
func (ta *TArchivist) SaveContractB(mid db.MarketMatchID, contract []byte, coinID []byte, timestamp int64) error {
	return nil
}
func (ta *TArchivist) SaveAuditAckSigA(mid db.MarketMatchID, sig []byte) error { return nil }

// Redeem data.
func (ta *TArchivist) SaveRedeemA(mid db.MarketMatchID, coinID, secret []byte, timestamp int64) error {
	return nil
}
func (ta *TArchivist) SaveRedeemAckSigB(mid db.MarketMatchID, sig []byte) error {
	return nil
}
func (ta *TArchivist) SaveRedeemB(mid db.MarketMatchID, coinID []byte, timestamp int64) error {
	return nil
}
func (ta *TArchivist) SetMatchInactive(mid db.MarketMatchID, forgive bool) error { return nil }
func (ta *TArchivist) LoadEpochStats(uint32, uint32, []*candles.Cache) error     { return nil }

type TCollector struct{}

var collectorSpot = &msgjson.Spot{
	Stamp: rand.Uint64(),
}

func (tc *TCollector) ReportEpoch(base, quote uint32, epochIdx uint64, stats *matcher.MatchCycleStats) (*msgjson.Spot, error) {
	return collectorSpot, nil
}

type tFeeFetcher struct {
	maxFeeRate uint64
}

func (*tFeeFetcher) FeeRate(context.Context) uint64 {
	return 10
}

func (f *tFeeFetcher) MaxFeeRate() uint64 {
	return f.maxFeeRate
}

func (f *tFeeFetcher) LastRate() uint64 {
	return 10
}

func (f *tFeeFetcher) SwapFeeRate(context.Context) uint64 {
	return 10
}

type tBalancer struct {
	reqs map[string]int
}

func newTBalancer() *tBalancer {
	return &tBalancer{make(map[string]int)}
}

func (b *tBalancer) CheckBalance(acctAddr string, assetID, redeemAssetID uint32, qty, lots uint64, redeems int) bool {
	b.reqs[acctAddr]++
	return true
}

func randomOrderID() order.OrderID {
	pk := randomBytes(order.OrderIDSize)
	var id order.OrderID
	copy(id[:], pk)
	return id
}

const (
	tUserTier, tUserScore, tMaxScore = int64(1), int32(30), int32(60)
)

var parcelLimit = float64(calcParcelLimit(tUserTier, tUserScore, tMaxScore))

func newTestMarket(opts ...any) (*Market, *TArchivist, *TAuth, func(), error) {
	// The DEX will make MasterCoinLockers for each asset.
	masterLockerBase := coinlock.NewMasterCoinLocker()
	bookLockerBase := masterLockerBase.Book()
	swapLockerBase := masterLockerBase.Swap()

	masterLockerQuote := coinlock.NewMasterCoinLocker()
	bookLockerQuote := masterLockerQuote.Book()
	swapLockerQuote := masterLockerQuote.Swap()

	epochDurationMSec := uint64(500) // 0.5 sec epoch duration
	storage := &TArchivist{}
	var balancer Balancer

	baseAsset, quoteAsset := assetDCR, assetBTC

	for _, opt := range opts {
		switch optT := opt.(type) {
		case *TArchivist:
			storage = optT
		case [2]*asset.BackedAsset:
			baseAsset, quoteAsset = optT[0], optT[1]
			if baseAsset.ID == assetETH.ID || baseAsset.ID == assetMATIC.ID {
				bookLockerBase = nil
			}
			if quoteAsset.ID == assetETH.ID || quoteAsset.ID == assetMATIC.ID {
				bookLockerQuote = nil
			}
		case *tBalancer:
			balancer = optT
		}

	}

	authMgr := &TAuth{
		sends:            make([]*msgjson.Message, 0),
		preimagesByMsgID: make(map[uint64]order.Preimage),
		preimagesByOrdID: make(map[string]order.Preimage),
	}

	var swapDone func(ord order.Order, match *order.Match, fail bool)
	swapperCfg := &swap.Config{
		Assets: map[uint32]*swap.SwapperAsset{
			assetDCR.ID:   {BackedAsset: assetDCR, Locker: swapLockerBase},
			assetBTC.ID:   {BackedAsset: assetBTC, Locker: swapLockerQuote},
			assetETH.ID:   {BackedAsset: assetETH},
			assetMATIC.ID: {BackedAsset: assetMATIC},
		},
		Storage:          storage,
		AuthManager:      authMgr,
		BroadcastTimeout: 10 * time.Second,
		TxWaitExpiration: 5 * time.Second,
		LockTimeTaker:    dex.LockTimeTaker(dex.Testnet),
		LockTimeMaker:    dex.LockTimeMaker(dex.Testnet),
		SwapDone: func(ord order.Order, match *order.Match, fail bool) {
			swapDone(ord, match, fail)
		},
	}
	swapper, err := swap.NewSwapper(swapperCfg)
	if err != nil {
		panic(err.Error())
	}

	mbBuffer := 1.1
	mktInfo, err := dex.NewMarketInfo(baseAsset.ID, quoteAsset.ID,
		dcrLotSize, btcRateStep, epochDurationMSec, mbBuffer)
	if err != nil {
		return nil, nil, nil, func() {}, fmt.Errorf("dex.NewMarketInfo() failure: %w", err)
	}

	mkt, err := NewMarket(&Config{
		MarketInfo:      mktInfo,
		Storage:         storage,
		Swapper:         swapper,
		AuthManager:     authMgr,
		FeeFetcherBase:  &tFeeFetcher{baseAsset.MaxFeeRate},
		CoinLockerBase:  bookLockerBase,
		FeeFetcherQuote: &tFeeFetcher{quoteAsset.MaxFeeRate},
		CoinLockerQuote: bookLockerQuote,
		DataCollector:   new(TCollector),
		Balancer:        balancer,
		CheckParcelLimit: func(_ account.AccountID, f MarketParcelCalculator) bool {
			parcels := f(0)
			return parcels <= parcelLimit
		},
	})
	if err != nil {
		return nil, nil, nil, func() {}, fmt.Errorf("Failed to create test market: %w", err)
	}

	swapDone = mkt.SwapDone

	ssw := dex.NewStartStopWaiter(swapper)
	ssw.Start(testCtx)
	cleanup := func() {
		ssw.Stop()
		ssw.WaitForShutdown()
	}

	return mkt, storage, authMgr, cleanup, nil
}

func TestMarket_NewMarket_BookOrders(t *testing.T) {
	mkt, storage, _, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
	}

	// With no book orders in the DB, the market should have an empty book after
	// construction.
	_, buys, sells := mkt.Book()
	if len(buys) > 0 || len(sells) > 0 {
		cleanup()
		t.Fatalf("Fresh market had %d buys and %d sells, expected none.",
			len(buys), len(sells))
	}
	cleanup()

	rnd.Seed(12)

	randCoinDCR := func() []byte {
		coinID := make([]byte, 36)
		rnd.Read(coinID[:])
		return coinID
	}

	// Now store some book orders to verify NewMarket sees them.
	loBuy := makeLO(buyer3, mkRate3(0.8, 1.0), randLots(10), order.StandingTiF)
	loBuy.FillAmt = mkt.marketInfo.LotSize // partial fill to cover utxo check alt. path
	loSell := makeLO(seller3, mkRate3(1.0, 1.2), randLots(10)+1, order.StandingTiF)
	fundingCoinDCR := randCoinDCR()
	loSell.Coins = []order.CoinID{fundingCoinDCR}
	// let VerifyUnspentCoin find this coin as unspent
	oRig.dcr.addUTXO(&msgjson.Coin{ID: fundingCoinDCR}, 1234)

	_ = storage.BookOrder(loBuy)  // the stub does not error
	_ = storage.BookOrder(loSell) // the stub does not error

	mkt, storage, _, cleanup, err = newTestMarket(storage)
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
	}
	defer cleanup()

	_, buys, sells = mkt.Book()
	if len(buys) != 1 || len(sells) != 1 {
		t.Fatalf("Fresh market had %d buys and %d sells, expected 1 buy, 1 sell.",
			len(buys), len(sells))
	}
	if buys[0].ID() != loBuy.ID() {
		t.Errorf("booked buy order has incorrect ID. Expected %v, got %v",
			loBuy.ID(), buys[0].ID())
	}
	if sells[0].ID() != loSell.ID() {
		t.Errorf("booked sell order has incorrect ID. Expected %v, got %v",
			loSell.ID(), sells[0].ID())
	}

	// PurgeBook should clear the in memory book and those in storage.
	mkt.PurgeBook()
	_, buys, sells = mkt.Book()
	if len(buys) > 0 || len(sells) > 0 {
		t.Fatalf("purged market had %d buys and %d sells, expected none.",
			len(buys), len(sells))
	}

	los, _ := storage.BookOrders(mkt.marketInfo.Base, mkt.marketInfo.Quote)
	if len(los) != 0 {
		t.Errorf("stored book orders were not flushed")
	}

}

func TestMarket_Book(t *testing.T) {
	mkt, storage, auth, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
	}
	defer cleanup()

	rnd.Seed(0)

	// Fill the book.
	for i := 0; i < 8; i++ {
		// Buys
		lo := makeLO(buyer3, mkRate3(0.8, 1.0), randLots(10), order.StandingTiF)
		if !mkt.book.Insert(lo) {
			t.Fatalf("Failed to Insert order into book.")
		}
		//t.Logf("Inserted buy order  (rate=%10d, quantity=%d) onto book.", lo.Rate, lo.Quantity)

		// Sells
		lo = makeLO(seller3, mkRate3(1.0, 1.2), randLots(10), order.StandingTiF)
		if !mkt.book.Insert(lo) {
			t.Fatalf("Failed to Insert order into book.")
		}
		//t.Logf("Inserted sell order (rate=%10d, quantity=%d) onto book.", lo.Rate, lo.Quantity)
	}

	bestBuy, bestSell := mkt.book.Best()

	marketRate := mkt.MidGap()
	mktRateWant := (bestBuy.Rate + bestSell.Rate) / 2
	if marketRate != mktRateWant {
		t.Errorf("Market rate expected %d, got %d", mktRateWant, mktRateWant)
	}

	_, buys, sells := mkt.Book()
	if buys[0] != bestBuy {
		t.Errorf("Incorrect best buy order. Got %v, expected %v",
			buys[0], bestBuy)
	}
	if sells[0] != bestSell {
		t.Errorf("Incorrect best sell order. Got %v, expected %v",
			sells[0], bestSell)
	}

	// unbook something not on the book
	if mkt.Unbook(makeLO(buyer3, 100, 1, order.StandingTiF)) {
		t.Fatalf("unbooked and order that was not on the book")
	}

	// unbook the best buy order
	feed := mkt.OrderFeed()

	if !mkt.Unbook(bestBuy) {
		t.Fatalf("Failed to unbook order")
	}

	sig := <-feed
	if sig.action != unbookAction {
		t.Fatalf("did not receive unbookAction signal")
	}
	sigData, ok := sig.data.(sigDataUnbookedOrder)
	if !ok {
		t.Fatalf("incorrect sigdata type")
	}
	if sigData.epochIdx != -1 {
		t.Fatalf("expected epoch index -1, got %d", sigData.epochIdx)
	}
	loUnbooked, ok := sigData.order.(*order.LimitOrder)
	if !ok {
		t.Fatalf("incorrect unbooked order type")
	}
	if loUnbooked.ID() != bestBuy.ID() {
		t.Errorf("unbooked order %v, wanted %v", loUnbooked.ID(), bestBuy.ID())
	}

	if auth.canceledOrder != bestBuy.ID() {
		t.Errorf("revoke not recorded with auth manager")
	}

	if storage.revoked.ID() != bestBuy.ID() {
		t.Errorf("revoke not recorded in storage")
	}

	if lockedCoins, _ := mkt.coinsLocked(bestBuy); lockedCoins != nil {
		t.Errorf("unbooked order still has locked coins: %v", lockedCoins)
	}

	bestBuy2, _ := mkt.book.Best()
	if bestBuy2 == bestBuy {
		t.Errorf("failed to unbook order")
	}

}

func TestMarket_Suspend(t *testing.T) {
	// Create the market.
	mkt, _, _, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
		cleanup()
		return
	}
	defer cleanup()
	epochDurationMSec := int64(mkt.EpochDuration())

	// Suspend before market start.
	finalIdx, _ := mkt.Suspend(time.Now(), false)
	if finalIdx != -1 {
		t.Fatalf("not running market should not allow suspend")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startEpochIdx := 2 + time.Now().UnixMilli()/epochDurationMSec
	startEpochTime := time.UnixMilli(startEpochIdx * epochDurationMSec)
	midPrevEpochTime := startEpochTime.Add(time.Duration(-epochDurationMSec/2) * time.Millisecond)

	// ~----|-------|-------|-------|
	// ^now ^prev   ^start  ^next

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Start(ctx, startEpochIdx)
	}()

	feed := mkt.OrderFeed()
	go func() {
		for range feed {
		}
	}()

	// Wait until half way through the epoch prior to start, when we know Run is
	// running but the market hasn't started yet.
	<-time.After(time.Until(midPrevEpochTime))

	// This tests the case where m.activeEpochIdx == 0 but start is scheduled.
	// The suspend (final) epoch should be the one just prior to startEpochIdx.
	persist := true
	finalIdx, finalTime := mkt.Suspend(time.Now(), persist)
	if finalIdx != startEpochIdx-1 {
		t.Fatalf("finalIdx = %d, wanted %d", finalIdx, startEpochIdx-1)
	}
	if !startEpochTime.Equal(finalTime) {
		t.Errorf("got finalTime = %v, wanted %v", finalTime, startEpochTime)
	}

	if mkt.suspendEpochIdx != finalIdx {
		t.Errorf("got suspendEpochIdx = %d, wanted = %d", mkt.suspendEpochIdx, finalIdx)
	}

	// Set a new suspend time, in the future this time.
	nextEpochIdx := startEpochIdx + 1
	nextEpochTime := time.UnixMilli(nextEpochIdx * epochDurationMSec)

	// Just before second epoch start.
	finalIdx, finalTime = mkt.Suspend(nextEpochTime.Add(-1*time.Millisecond), persist)
	if finalIdx != nextEpochIdx-1 {
		t.Fatalf("finalIdx = %d, wanted %d", finalIdx, nextEpochIdx-1)
	}
	if !nextEpochTime.Equal(finalTime) {
		t.Errorf("got finalTime = %v, wanted %v", finalTime, nextEpochTime)
	}

	if mkt.suspendEpochIdx != finalIdx {
		t.Errorf("got suspendEpochIdx = %d, wanted = %d", mkt.suspendEpochIdx, finalIdx)
	}

	// Exactly at second epoch start, with same result.
	finalIdx, finalTime = mkt.Suspend(nextEpochTime, persist)
	if finalIdx != nextEpochIdx-1 {
		t.Fatalf("finalIdx = %d, wanted %d", finalIdx, nextEpochIdx-1)
	}
	if !nextEpochTime.Equal(finalTime) {
		t.Errorf("got finalTime = %v, wanted %v", finalTime, nextEpochTime)
	}

	if mkt.suspendEpochIdx != finalIdx {
		t.Errorf("got suspendEpochIdx = %d, wanted = %d", mkt.suspendEpochIdx, finalIdx)
	}

	mkt.waitForEpochOpen()

	// should be running
	if !mkt.Running() {
		t.Fatal("the market should have be running")
	}

	// Wait until after suspend time.
	<-time.After(time.Until(finalTime.Add(20 * time.Millisecond)))

	// should be stopped
	if mkt.Running() {
		t.Fatal("the market should have been suspended")
	}

	wg.Wait()
	mkt.FeedDone(feed)

	// Start up again (consumer resumes the Market manually)
	startEpochIdx = 1 + time.Now().UnixMilli()/epochDurationMSec
	// startEpochTime = time.UnixMilli(startEpochIdx * epochDurationMSec)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Start(ctx, startEpochIdx)
	}()

	feed = mkt.OrderFeed()
	go func() {
		for range feed {
		}
	}()

	mkt.waitForEpochOpen()

	// should be running
	if !mkt.Running() {
		t.Fatal("the market should have be running")
	}

	// Suspend asap.
	_, finalTime = mkt.SuspendASAP(persist)
	<-time.After(time.Until(finalTime.Add(40 * time.Millisecond)))

	// Should be stopped
	if mkt.Running() {
		t.Fatal("the market should have been suspended")
	}

	cancel()
	wg.Wait()
	mkt.FeedDone(feed)
}

func TestMarket_Suspend_Persist(t *testing.T) {
	// Create the market.
	mkt, storage, _, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
		cleanup()
		return
	}
	defer cleanup()
	epochDurationMSec := int64(mkt.EpochDuration())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startEpochIdx := 2 + time.Now().UnixMilli()/epochDurationMSec
	//startEpochTime := time.UnixMilli(startEpochIdx * epochDurationMSec)

	// ~----|-------|-------|-------|
	// ^now ^prev   ^start  ^next

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Start(ctx, startEpochIdx)
	}()

	startFeedRecv := func(feed <-chan *updateSignal) {
		go func() {
			for range feed {
			}
		}()
	}

	// Wait until after original start time.
	mkt.waitForEpochOpen()

	if !mkt.Running() {
		t.Fatal("the market should be running")
	}

	lo := makeLO(seller3, mkRate3(0.8, 1.0), randLots(10), order.StandingTiF)
	ok := mkt.book.Insert(lo)
	if !ok {
		t.Fatalf("Failed to insert an order into Market's Book")
	}
	_ = storage.BookOrder(lo)

	// Suspend asap with no resume.  The epoch with the limit order will be
	// processed and then the market will suspend.
	//wantClosedFeed = true // allow the feed receiver goroutine to return w/o error
	persist := true
	_, finalTime := mkt.SuspendASAP(persist)
	<-time.After(time.Until(finalTime.Add(40 * time.Millisecond)))

	// Wait for Run to return.
	wg.Wait()

	// Should be stopped
	if mkt.Running() {
		t.Fatal("the market should have been suspended")
	}

	// Verify the order is still there.
	los, _ := storage.BookOrders(mkt.marketInfo.Base, mkt.marketInfo.Quote)
	if len(los) == 0 {
		t.Errorf("stored book orders were flushed")
	}

	_, buys, sells := mkt.Book()
	if len(buys) != 0 {
		t.Errorf("buy side of book not empty")
	}
	if len(sells) != 1 {
		t.Errorf("sell side of book not equal to 1")
	}

	// Start it up again.
	feed := mkt.OrderFeed()
	startEpochIdx = 1 + time.Now().UnixMilli()/epochDurationMSec
	//startEpochTime = time.UnixMilli(startEpochIdx * epochDurationMSec)
	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Start(ctx, startEpochIdx)
	}()

	startFeedRecv(feed)

	mkt.waitForEpochOpen()

	if !mkt.Running() {
		t.Fatal("the market should be running")
	}

	persist = false
	_, finalTime = mkt.SuspendASAP(persist)
	<-time.After(time.Until(finalTime.Add(40 * time.Millisecond)))

	// Wait for Run to return.
	wg.Wait()
	mkt.FeedDone(feed)

	// Should be stopped
	if mkt.Running() {
		t.Fatal("the market should have been suspended")
	}

	// Verify the order is gone.
	los, _ = storage.BookOrders(mkt.marketInfo.Base, mkt.marketInfo.Quote)
	if len(los) != 0 {
		t.Errorf("stored book orders were not flushed")
	}

	_, buys, sells = mkt.Book()
	if len(buys) != 0 {
		t.Errorf("buy side of book not empty")
	}
	if len(sells) != 0 {
		t.Errorf("sell side of book not empty")
	}

	if t.Failed() {
		cancel()
		wg.Wait()
	}
}

func TestMarket_Run(t *testing.T) {
	// This test exercises the Market's main loop, which cycles the epochs and
	// queues (or not) incoming orders.

	// Create the market.
	mkt, storage, auth, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
		cleanup()
		return
	}
	epochDurationMSec := int64(mkt.EpochDuration())
	// This test wants to know when epoch order matching booking is done.
	storage.epochInserted = make(chan struct{}, 1)
	// and when handlePreimage is done.
	auth.handlePreimageDone = make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check that start is delayed by an unsynced backend. Tell the Market to
	// start
	atomic.StoreUint32(&oRig.dcr.synced, 0)
	nowEpochIdx := time.Now().UnixMilli()/epochDurationMSec + 1

	unsyncedEpochIdx := nowEpochIdx + 1
	unsyncedEpochTime := time.UnixMilli(unsyncedEpochIdx * epochDurationMSec)

	startEpochIdx := unsyncedEpochIdx + 1
	startEpochTime := time.UnixMilli(startEpochIdx * epochDurationMSec)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Start(ctx, unsyncedEpochIdx)
	}()

	// Make an order for the first epoch.
	clientTimeMSec := startEpochIdx*epochDurationMSec + 10 // 10 ms after epoch start
	lots := 1
	qty := uint64(dcrLotSize * lots)
	rate := uint64(1000) * dcrRateStep
	aid := test.NextAccount()
	pi := test.RandomPreimage()
	commit := pi.Commit()
	limit := &msgjson.LimitOrder{
		Prefix: msgjson.Prefix{
			AccountID:  aid[:],
			Base:       dcrID,
			Quote:      btcID,
			OrderType:  msgjson.LimitOrderNum,
			ClientTime: uint64(clientTimeMSec),
			Commit:     commit[:],
		},
		Trade: msgjson.Trade{
			Side:     msgjson.SellOrderNum,
			Quantity: qty,
			Coins:    []*msgjson.Coin{},
			Address:  btcAddr,
		},
		Rate: rate,
		TiF:  msgjson.StandingOrderNum,
	}

	newLimit := func() *order.LimitOrder {
		return &order.LimitOrder{
			P: order.Prefix{
				AccountID:  aid,
				BaseAsset:  limit.Base,
				QuoteAsset: limit.Quote,
				OrderType:  order.LimitOrderType,
				ClientTime: time.UnixMilli(clientTimeMSec),
				Commit:     commit,
			},
			T: order.Trade{
				Coins:    []order.CoinID{},
				Sell:     true,
				Quantity: limit.Quantity,
				Address:  limit.Address,
			},
			Rate:  limit.Rate,
			Force: order.StandingTiF,
		}
	}

	parcelQty := uint64(dcrLotSize)
	maxMakerQty := parcelQty * uint64(parcelLimit)
	maxTakerQty := maxMakerQty / 2

	var msgID uint64
	nextMsgID := func() uint64 { msgID++; return msgID }
	newOR := func() *orderRecord {
		return &orderRecord{
			msgID: nextMsgID(),
			req:   limit,
			order: newLimit(),
		}
	}

	storMsgPI := func(id uint64, pi order.Preimage) {
		auth.piMtx.Lock()
		auth.preimagesByMsgID[id] = pi
		auth.piMtx.Unlock()
	}

	oRecord := newOR()
	storMsgPI(oRecord.msgID, pi)
	//auth.Send will update preimagesByOrderID

	// Submit order before market starts running
	err = mkt.SubmitOrder(oRecord)
	if err == nil {
		t.Error("order successfully submitted to stopped market")
	}
	if !errors.Is(err, ErrMarketNotRunning) {
		t.Fatalf(`expected ErrMarketNotRunning ("%v"), got "%v"`, ErrMarketNotRunning, err)
	}

	mktStatus := mkt.Status()
	if mktStatus.Running {
		t.Fatalf("Market should not be running yet")
	}

	halfEpoch := time.Duration(epochDurationMSec/2) * time.Millisecond

	<-time.After(time.Until(unsyncedEpochTime.Add(halfEpoch)))

	if mkt.Running() {
		t.Errorf("market running with an unsynced backend")
	}

	atomic.StoreUint32(&oRig.dcr.synced, 1)

	<-time.After(time.Until(startEpochTime.Add(halfEpoch)))
	<-storage.epochInserted

	if !mkt.Running() {
		t.Errorf("market not running after backend sync finished")
	}

	// Submit again.
	limit.Quantity = dcrLotSize

	oRecord = newOR()
	storMsgPI(oRecord.msgID, pi)
	err = mkt.SubmitOrder(oRecord)
	if err != nil {
		t.Fatal(err)
	}

	// Let the epoch cycle and the fake client respond with its preimage
	// (handlePreimageResp done)...
	<-auth.handlePreimageDone
	// and for matching to complete (in processReadyEpoch).
	<-storage.epochInserted

	// Submit an immediate taker sell (taker) over user taker limit

	piSell := test.RandomPreimage()
	commitSell := piSell.Commit()
	oRecordSell := newOR()
	limit.Commit = commitSell[:]
	loSell := oRecordSell.order.(*order.LimitOrder)
	loSell.P.Commit = commitSell
	loSell.Force = order.ImmediateTiF // likely taker
	loSell.Quantity = maxTakerQty     // one lot already booked

	storMsgPI(oRecordSell.msgID, pi)
	err = mkt.SubmitOrder(oRecordSell)
	if err == nil {
		t.Fatal("should have rejected too large likely-taker")
	}

	// Submit a taker buy that is over user taker limit
	// loSell := oRecord.order.(*order.LimitOrder)
	piBuy := test.RandomPreimage()
	commitBuy := piBuy.Commit()
	oRecordBuy := newOR()
	limit.Commit = commitBuy[:]
	loBuy := oRecordBuy.order.(*order.LimitOrder)
	loBuy.P.Commit = commitBuy
	loBuy.Sell = false
	loBuy.Quantity = maxTakerQty // One lot already booked
	// rate matches with the booked sell = likely taker

	storMsgPI(oRecordBuy.msgID, piBuy)
	err = mkt.SubmitOrder(oRecordBuy)
	if err == nil {
		t.Fatal("should have rejected too large likely-taker")
	}

	// Submit a likely taker with an acceptable limit
	loSell.Quantity = maxTakerQty - dcrLotSize // the limit

	storMsgPI(oRecordSell.msgID, piSell)
	err = mkt.SubmitOrder(oRecordSell)
	if err != nil {
		t.Fatalf("should have allowed that likely-taker: %v", err)
	}

	// Another in the same epoch will push over the limit
	loBuy.Quantity = dcrLotSize // just one lot
	storMsgPI(oRecordBuy.msgID, pi)
	err = mkt.SubmitOrder(oRecordBuy)
	if err == nil {
		t.Fatalf("should have rejected too likely-taker that pushed the limit with existing epoch status takers")
	}

	// Submit a valid cancel order.
	loID := oRecord.order.ID()
	piCo := test.RandomPreimage()
	commit = piCo.Commit()
	cancelTime := time.Now().UnixMilli()
	cancelMsg := &msgjson.CancelOrder{
		Prefix: msgjson.Prefix{
			AccountID:  aid[:],
			Base:       dcrID,
			Quote:      btcID,
			OrderType:  msgjson.CancelOrderNum,
			ClientTime: uint64(cancelTime),
			Commit:     commit[:],
		},
		TargetID: loID[:],
	}

	newCancel := func() *order.CancelOrder {
		return &order.CancelOrder{
			P: order.Prefix{
				AccountID:  aid,
				BaseAsset:  limit.Base,
				QuoteAsset: limit.Quote,
				OrderType:  order.CancelOrderType,
				ClientTime: time.UnixMilli(cancelTime),
				Commit:     commit,
			},
			TargetOrderID: loID,
		}
	}
	co := newCancel()

	coRecord := orderRecord{
		msgID: nextMsgID(),
		req:   cancelMsg,
		order: co,
	}

	// Cancel order w/o permission to cancel target order (the limit order from
	// above that is now booked)
	cancelTime++
	otherAccount := test.NextAccount()
	cancelMsg.ClientTime = uint64(cancelTime)
	cancelMsg.AccountID = otherAccount[:]
	coWrongAccount := newCancel()
	piBadCo := test.RandomPreimage()
	commitBadCo := piBadCo.Commit()
	coWrongAccount.Commit = commitBadCo
	coWrongAccount.AccountID = otherAccount
	coWrongAccount.ClientTime = time.UnixMilli(cancelTime)
	cancelMsg.Commit = commitBadCo[:]
	coRecordWrongAccount := orderRecord{
		msgID: nextMsgID(),
		req:   cancelMsg,
		order: coWrongAccount,
	}

	// Submit the invalid cancel order first because it would be caught by the
	// duplicate check if we do it after the valid one is submitted.
	storMsgPI(coRecordWrongAccount.msgID, piBadCo)
	err = mkt.SubmitOrder(&coRecordWrongAccount)
	if err == nil {
		t.Errorf("An invalid order was processed, but it should not have been.")
	} else if !errors.Is(err, ErrCancelNotPermitted) {
		t.Errorf(`expected ErrCancelNotPermitted ("%v"), got "%v"`, ErrCancelNotPermitted, err)
	}

	// Valid cancel order
	storMsgPI(coRecord.msgID, piCo)
	err = mkt.SubmitOrder(&coRecord)
	if err != nil {
		t.Fatalf("Failed to submit order: %v", err)
	}

	// Duplicate cancel order
	piCoDup := test.RandomPreimage()
	commit = piCoDup.Commit()
	cancelTime++
	cancelMsg.ClientTime = uint64(cancelTime)
	cancelMsg.Commit = commit[:]
	coDup := newCancel()
	coDup.Commit = commit
	coDup.ClientTime = time.UnixMilli(cancelTime)
	coRecordDup := orderRecord{
		msgID: nextMsgID(),
		req:   cancelMsg,
		order: coDup,
	}
	storMsgPI(coRecordDup.msgID, piCoDup)
	err = mkt.SubmitOrder(&coRecordDup)
	if err == nil {
		t.Errorf("An duplicate cancel order was processed, but it should not have been.")
	} else if !errors.Is(err, ErrDuplicateCancelOrder) {
		t.Errorf(`expected ErrDuplicateCancelOrder ("%v"), got "%v"`, ErrDuplicateCancelOrder, err)
	}

	// Let the epoch cycle and the fake client respond with its preimage
	// (handlePreimageResp done)..
	<-auth.handlePreimageDone
	// and for matching to complete (in processReadyEpoch).
	<-storage.epochInserted

	cancel()
	wg.Wait()
	cleanup()

	// Test duplicate order (commitment) with a new Market.
	mkt, storage, auth, cleanup, err = newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
	}
	storage.epochInserted = make(chan struct{}, 1)
	auth.handlePreimageDone = make(chan struct{}, 1)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Run(ctx)
	}()
	mkt.waitForEpochOpen()

	// fresh oRecord
	oRecord = newOR()
	storMsgPI(oRecord.msgID, pi)
	err = mkt.SubmitOrder(oRecord)
	if err != nil {
		t.Error(err)
	}

	// Submit another order with the same Commitment in the same Epoch.
	oRecord = newOR()
	storMsgPI(oRecord.msgID, pi)
	err = mkt.SubmitOrder(oRecord)
	if err == nil {
		t.Errorf("A duplicate order was processed, but it should not have been.")
	} else if !errors.Is(err, ErrInvalidCommitment) {
		t.Errorf(`expected ErrInvalidCommitment ("%v"), got "%v"`, ErrInvalidCommitment, err)
	}

	// Send an order with a bad lot size.
	oRecord = newOR()
	oRecord.order.(*order.LimitOrder).Quantity += mkt.marketInfo.LotSize / 2
	storMsgPI(oRecord.msgID, pi)
	err = mkt.SubmitOrder(oRecord)
	if err == nil {
		t.Errorf("An invalid order was processed, but it should not have been.")
	} else if !errors.Is(err, ErrInvalidOrder) {
		t.Errorf(`expected ErrInvalidOrder ("%v"), got "%v"`, ErrInvalidOrder, err)
	}

	// Rate too low
	oRecord = newOR()
	mkt.minimumRate = oRecord.order.(*order.LimitOrder).Rate + 1
	storMsgPI(oRecord.msgID, pi)
	if err = mkt.SubmitOrder(oRecord); !errors.Is(err, ErrInvalidRate) {
		t.Errorf("An invalid rate was accepted, but it should not have been.")
	}
	mkt.minimumRate = 0

	// Let the epoch cycle and the fake client respond with its preimage
	// (handlePreimageResp done)..
	<-auth.handlePreimageDone
	// and for matching to complete (in processReadyEpoch).
	<-storage.epochInserted

	// Submit an order with a Commitment known to the DB.
	// NOTE: disabled since the OrderWithCommit check in Market.processOrder is disabled too.
	// oRecord = newOR()
	// oRecord.order.SetTime(time.Now()) // This will register a different order ID with the DB in the next statement.
	// storage.failOnCommitWithOrder(oRecord.order)
	// storMsgPI(oRecord.msgID, pi)
	// err = mkt.SubmitOrder(oRecord) // Will re-stamp the order, but the commit will be the same.
	// if err == nil {
	// 	t.Errorf("A duplicate order was processed, but it should not have been.")
	// } else if !errors.Is(err, ErrInvalidCommitment) {
	// 	t.Errorf(`expected ErrInvalidCommitment ("%v"), got "%v"`, ErrInvalidCommitment, err)
	// }

	// Submit an order with a zero commit.
	oRecord = newOR()
	oRecord.order.(*order.LimitOrder).Commit = order.Commitment{}
	storMsgPI(oRecord.msgID, pi)
	err = mkt.SubmitOrder(oRecord)
	if err == nil {
		t.Errorf("An order with a zero Commitment was processed, but it should not have been.")
	} else if !errors.Is(err, ErrInvalidCommitment) {
		t.Errorf(`expected ErrInvalidCommitment ("%v"), got "%v"`, ErrInvalidCommitment, err)
	}

	// Submit an order that breaks storage somehow.
	// tweak the order's commitment+preimage so it's not a dup.
	oRecord = newOR()
	pi = test.RandomPreimage()
	commit = pi.Commit()
	lo := oRecord.order.(*order.LimitOrder)
	lo.Commit = commit
	limit.Commit = commit[:] // oRecord.req
	storMsgPI(oRecord.msgID, pi)
	storage.failOnEpochOrder(lo) // force storage to fail on this order
	if err = mkt.SubmitOrder(oRecord); !errors.Is(err, ErrInternalServer) {
		t.Errorf(`expected ErrInternalServer ("%v"), got "%v"`, ErrInternalServer, err)
	}

	// NOTE: The Market is now stopping on its own because of the storage failure.

	wg.Wait()
	cleanup()
}

func TestMarket_enqueueEpoch(t *testing.T) {
	// This tests processing of a closed epoch by prepEpoch (for preimage
	// collection) and processReadyEpoch (for sending the expected book and
	// unbook messages to book subscribers registered via OrderFeed) via
	// enqueueEpoch and the epochPump.

	mkt, _, auth, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("Failed to create test market: %v", err)
		return
	}
	defer cleanup()

	rnd.Seed(0) // deterministic random data

	// Fill the book. Preimages not needed for these.
	for i := 0; i < 8; i++ {
		// Buys
		lo := makeLO(buyer3, mkRate3(0.8, 1.0), randLots(10), order.StandingTiF)
		if !mkt.book.Insert(lo) {
			t.Fatalf("Failed to Insert order into book.")
		}
		//t.Logf("Inserted buy order (rate=%10d, quantity=%d) onto book.", lo.Rate, lo.Quantity)

		// Sells
		lo = makeLO(seller3, mkRate3(1.0, 1.2), randLots(10), order.StandingTiF)
		if !mkt.book.Insert(lo) {
			t.Fatalf("Failed to Insert order into book.")
		}
		//t.Logf("Inserted sell order (rate=%10d, quantity=%d) onto book.", lo.Rate, lo.Quantity)
	}

	bestBuy, bestSell := mkt.book.Best()
	bestBuyRate := bestBuy.Rate
	bestBuyQuant := bestBuy.Quantity * 3 // tweak for new shuffle seed without changing csum
	bestSellID := bestSell.ID()

	var epochIdx, epochDur int64 = 123413513, int64(mkt.marketInfo.EpochDuration)
	eq := NewEpoch(epochIdx, epochDur)
	eID := order.EpochID{Idx: uint64(epochIdx), Dur: uint64(epochDur)}
	lo, loPI := makeLORevealed(seller3, bestBuyRate-dcrRateStep, bestBuyQuant, order.StandingTiF)
	co, coPI := makeCORevealed(buyer3, bestSellID)
	eq.Insert(lo)
	eq.Insert(co)

	cSum, _ := hex.DecodeString("4859aa186630c2b135074037a8db42f240bbbe81c1361d8783aa605ed3f0cf90")
	seed, _ := hex.DecodeString("e061777b09170c80ce7049439bef0d69649f361ed16b500b5e53b80920813c54")
	mp := &order.MatchProof{
		Epoch:     eID,
		Preimages: []order.Preimage{loPI, coPI},
		Misses:    nil,
		CSum:      cSum,
		Seed:      seed,
	}

	// Test with a missed preimage.
	eq2 := NewEpoch(epochIdx, epochDur)
	co2, co2PI := makeCORevealed(buyer3, randomOrderID())
	lo2, _ := makeLORevealed(seller3, bestBuyRate-dcrRateStep, bestBuyQuant, order.ImmediateTiF)
	eq2.Insert(co2)
	eq2.Insert(lo2) // lo2 will not be in preimage map (miss)

	cSum2, _ := hex.DecodeString("a64ee6372a49f9465910ca0b556818dbc765f3c7fa21d5f40ab25bf4b73f45ed") // includes both commitments, including the miss
	seed2, _ := hex.DecodeString("aba75140b1f6edf26955a97e1b09d7b17abdc9c0b099fc73d9729501652fbf66") // includes only the provided preimage
	mp2 := &order.MatchProof{
		Epoch:     eID,
		Preimages: []order.Preimage{co2PI},
		Misses:    []order.Order{lo2},
		CSum:      cSum2,
		Seed:      seed2,
	}

	auth.piMtx.Lock()
	auth.preimagesByOrdID[lo.UID()] = loPI
	auth.preimagesByOrdID[co.UID()] = coPI
	auth.preimagesByOrdID[co2.UID()] = co2PI
	// No lo2 (miss)
	auth.piMtx.Unlock()

	var bookSignals []*updateSignal
	var mtx sync.Mutex
	// intercept what would go to an OrderFeed() chan of Run were running.
	notifyChan := make(chan *updateSignal, 32)
	defer close(notifyChan) // quit bookSignals receiver, but not necessary
	go func() {
		for up := range notifyChan {
			//fmt.Println("received signal", up.action)
			mtx.Lock()
			bookSignals = append(bookSignals, up)
			mtx.Unlock()
		}
	}()

	var wg sync.WaitGroup
	defer wg.Wait() // wait for the following epoch pipeline goroutines

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // stop the following epoch pipeline goroutines

	// This test does not start the entire market, so manually start the epoch
	// queue pump, and a goroutine to receive ready (preimage collection
	// completed) epochs and start matching, etc.
	ePump := newEpochPump()
	wg.Add(1)
	go func() {
		defer wg.Done()
		ePump.Run(ctx)
	}()

	goForIt := make(chan struct{}, 1)

	wg.Add(1)
	go func() {
		defer close(goForIt)
		defer wg.Done()
		for ep := range ePump.ready {
			t.Logf("processReadyEpoch: %d orders revealed\n", len(ep.ordersRevealed))

			// prepEpoch has completed preimage collection.
			mkt.processReadyEpoch(ep, notifyChan) // notify is async!
			goForIt <- struct{}{}
		}
	}()

	// MatchProof for empty epoch queue.
	mp0 := &order.MatchProof{
		Epoch: eID,
		// everything else is nil
	}

	tests := []struct {
		name                string
		epoch               *EpochQueue
		expectedBookSignals []*updateSignal
	}{
		{
			"ok book unbook",
			eq,
			[]*updateSignal{
				{matchProofAction, sigDataMatchProof{mp}},
				{bookAction, sigDataBookedOrder{lo, epochIdx}},
				{unbookAction, sigDataUnbookedOrder{bestBuy, epochIdx}},
				{unbookAction, sigDataUnbookedOrder{bestSell, epochIdx}},
				{epochReportAction, sigDataEpochReport{epochIdx, epochDur, nil, nil, 10, 10, nil}},
			},
		},
		{
			"ok no matches or book updates, one miss",
			eq2,
			[]*updateSignal{
				{matchProofAction, sigDataMatchProof{mp2}},
				{epochReportAction, sigDataEpochReport{epochIdx, epochDur, nil, nil, 10, 10, nil}},
			},
		},
		{
			"ok empty queue",
			NewEpoch(epochIdx, epochDur),
			[]*updateSignal{
				{matchProofAction, sigDataMatchProof{mp0}},
				{epochReportAction, sigDataEpochReport{epochIdx, epochDur, nil, nil, 10, 10, nil}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mkt.enqueueEpoch(ePump, tt.epoch)
			// Wait for processReadyEpoch, which sends on buffered (async) book
			// order feed channels.
			<-goForIt
			// Preimage collection has completed, but notifications are asynchronous.
			runtime.Gosched()                  // defer to the notify goroutine in (*Market).Run, somewhat redundant with the following sleep
			time.Sleep(250 * time.Millisecond) // let the test goroutine receive the signals on notifyChan, updating bookSignals
			// TODO: if this sleep becomes a problem, a receive(expectedNotes int) function might be needed
			mtx.Lock()
			defer mtx.Unlock() // inside this closure
			defer func() { bookSignals = []*updateSignal{} }()
			if len(bookSignals) != len(tt.expectedBookSignals) {
				t.Fatalf("expected %d book update signals, got %d",
					len(tt.expectedBookSignals), len(bookSignals))
			}
			for i, s := range bookSignals {
				exp := tt.expectedBookSignals[i]
				if exp.action != s.action {
					t.Errorf("Book signal #%d has action %d, expected %d",
						i, s.action, exp.action)
				}

				switch sigData := s.data.(type) {
				case sigDataMatchProof:
					mp := sigData.matchProof
					wantMp := exp.data.(sigDataMatchProof).matchProof
					if !bytes.Equal(wantMp.CSum, mp.CSum) {
						t.Errorf("Book signal #%d (action %v), has CSum %x, expected %x",
							i, s.action, mp.CSum, wantMp.CSum)
					}
					if !bytes.Equal(wantMp.Seed, mp.Seed) {
						t.Errorf("Book signal #%d (action %v), has Seed %x, expected %x",
							i, s.action, mp.Seed, wantMp.Seed)
					}
					if wantMp.Epoch.Idx != mp.Epoch.Idx {
						t.Errorf("Book signal #%d (action %v), has Epoch Idx %d, expected %d",
							i, s.action, mp.Epoch.Idx, wantMp.Epoch.Idx)
					}
					if wantMp.Epoch.Dur != mp.Epoch.Dur {
						t.Errorf("Book signal #%d (action %v), has Epoch Dur %d, expected %d",
							i, s.action, mp.Epoch.Dur, wantMp.Epoch.Dur)
					}
					if len(wantMp.Preimages) != len(mp.Preimages) {
						t.Errorf("Book signal #%d (action %v), has %d Preimages, expected %d",
							i, s.action, len(mp.Preimages), len(wantMp.Preimages))
						continue
					}
					for ii := range wantMp.Preimages {
						if wantMp.Preimages[ii] != mp.Preimages[ii] {
							t.Errorf("Book signal #%d (action %v), has #%d Preimage %x, expected %x",
								i, s.action, ii, mp.Preimages[ii], wantMp.Preimages[ii])
						}
					}
					if len(wantMp.Misses) != len(mp.Misses) {
						t.Errorf("Book signal #%d (action %v), has %d Misses, expected %d",
							i, s.action, len(mp.Misses), len(wantMp.Misses))
						continue
					}
					for ii := range wantMp.Misses {
						if wantMp.Misses[ii].ID() != mp.Misses[ii].ID() {
							t.Errorf("Book signal #%d (action %v), has #%d missed Order %v, expected %v",
								i, s.action, ii, mp.Misses[ii].ID(), wantMp.Misses[ii].ID())
						}
					}

				case sigDataBookedOrder:
					wantOrd := exp.data.(sigDataBookedOrder).order
					if wantOrd.ID() != sigData.order.ID() {
						t.Errorf("Book signal #%d (action %v) has order %v, expected %v",
							i, s.action, sigData.order.ID(), wantOrd.ID())
					}

				case sigDataUnbookedOrder:
					wantOrd := exp.data.(sigDataUnbookedOrder).order
					if wantOrd.ID() != sigData.order.ID() {
						t.Errorf("Unbook signal #%d (action %v) has order %v, expected %v",
							i, s.action, sigData.order.ID(), wantOrd.ID())
					}

				case sigDataNewEpoch:
					wantIdx := exp.data.(sigDataNewEpoch).idx
					if wantIdx != sigData.idx {
						t.Errorf("new epoch signal #%d (action %v) has epoch index %d, expected %d",
							i, s.action, sigData.idx, wantIdx)
					}

				case sigDataEpochReport:
					expSig := exp.data.(sigDataEpochReport)
					if expSig.epochIdx != sigData.epochIdx {
						t.Errorf("epoch report signal #%d (action %v) has epoch index %d, expected %d",
							i, s.action, sigData.epochIdx, expSig.epochIdx)
					}
					if expSig.epochDur != sigData.epochDur {
						t.Errorf("epoch report signal #%d (action %v) has epoch duration %d, expected %d",
							i, s.action, sigData.epochDur, expSig.epochDur)
					}
				}

			}
		})
	}

	cancel()
}

func TestMarket_Cancelable(t *testing.T) {
	// Create the market.
	mkt, storage, auth, cleanup, err := newTestMarket()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
		return
	}
	defer cleanup()
	// This test wants to know when epoch order matching booking is done.
	storage.epochInserted = make(chan struct{}, 1)
	// and when handlePreimage is done.
	auth.handlePreimageDone = make(chan struct{}, 1)

	epochDurationMSec := int64(mkt.EpochDuration())
	startEpochIdx := 1 + time.Now().UnixMilli()/epochDurationMSec
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mkt.Start(ctx, startEpochIdx)
	}()

	// Make an order for the first epoch.
	clientTimeMSec := startEpochIdx*epochDurationMSec + 10 // 10 ms after epoch start
	lots := dex.PerTierBaseParcelLimit
	qty := uint64(dcrLotSize * lots)
	rate := uint64(1000) * dcrRateStep
	aid := test.NextAccount()
	pi := test.RandomPreimage()
	commit := pi.Commit()
	limitMsg := &msgjson.LimitOrder{
		Prefix: msgjson.Prefix{
			AccountID:  aid[:],
			Base:       dcrID,
			Quote:      btcID,
			OrderType:  msgjson.LimitOrderNum,
			ClientTime: uint64(clientTimeMSec),
			Commit:     commit[:],
		},
		Trade: msgjson.Trade{
			Side:     msgjson.SellOrderNum,
			Quantity: qty,
			Coins:    []*msgjson.Coin{},
			Address:  btcAddr,
		},
		Rate: rate,
		TiF:  msgjson.StandingOrderNum,
	}

	newLimit := func() *order.LimitOrder {
		return &order.LimitOrder{
			P: order.Prefix{
				AccountID:  aid,
				BaseAsset:  limitMsg.Base,
				QuoteAsset: limitMsg.Quote,
				OrderType:  order.LimitOrderType,
				ClientTime: time.UnixMilli(clientTimeMSec),
				Commit:     commit,
			},
			T: order.Trade{
				Coins:    []order.CoinID{},
				Sell:     true,
				Quantity: limitMsg.Quantity,
				Address:  limitMsg.Address,
			},
			Rate:  limitMsg.Rate,
			Force: order.StandingTiF,
		}
	}
	lo := newLimit()

	oRecord := orderRecord{
		msgID: 1,
		req:   limitMsg,
		order: lo,
	}

	auth.piMtx.Lock()
	auth.preimagesByMsgID[oRecord.msgID] = pi
	auth.piMtx.Unlock()

	// Wait for the start of the epoch to submit the order.
	mkt.waitForEpochOpen()

	if mkt.Cancelable(order.OrderID{}) {
		t.Errorf("Cancelable reported bogus order as is cancelable, " +
			"but it wasn't even submitted.")
	}

	// Submit the standing limit order into the current epoch.
	err = mkt.SubmitOrder(&oRecord)
	if err != nil {
		t.Fatal(err)
	}

	if !mkt.Cancelable(lo.ID()) {
		t.Errorf("Cancelable failed to report order %v as cancelable, "+
			"but it was in the epoch queue", lo)
	}

	// Let the epoch cycle and the fake client respond with its preimage
	// (handlePreimageResp done)..
	<-auth.handlePreimageDone
	// and for matching to complete (in processReadyEpoch).
	<-storage.epochInserted

	if !mkt.Cancelable(lo.ID()) {
		t.Errorf("Cancelable failed to report order %v as cancelable, "+
			"but it should have been booked.", lo)
	}

	mkt.bookMtx.Lock()
	_, ok := mkt.book.Remove(lo.ID())
	mkt.bookMtx.Unlock()
	if !ok {
		t.Errorf("Failed to remove order %v from the book.", lo)
	}

	if mkt.Cancelable(lo.ID()) {
		t.Errorf("Cancelable reported order %v as is cancelable, "+
			"but it was removed from the Book.", lo)
	}

	cancel()
	wg.Wait()
}

func TestMarket_handlePreimageResp(t *testing.T) {
	randomCommit := func() (com order.Commitment) {
		rnd.Read(com[:])
		return
	}

	newOrder := func() (*order.LimitOrder, order.Preimage) {
		qty := uint64(dcrLotSize * 10)
		rate := uint64(1000) * dcrRateStep
		return makeLORevealed(seller3, rate, qty, order.StandingTiF)
	}

	authMgr := &TAuth{}
	mkt := &Market{
		auth:    authMgr,
		storage: &TArchivist{},
	}

	piMsg := &msgjson.PreimageResponse{
		Preimage: msgjson.Bytes{},
	}
	msg, _ := msgjson.NewResponse(5, piMsg, nil)

	runAndReceive := func(msg *msgjson.Message, dat *piData) *order.Preimage {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			mkt.handlePreimageResp(msg, dat)
			wg.Done()
		}()
		piRes := <-dat.preimage
		wg.Wait()
		return piRes
	}

	// 1. bad Message.Type: RPCParseError
	msg.Type = msgjson.Request // should be Response
	lo, pi := newOrder()
	dat := &piData{lo, make(chan *order.Preimage)}
	piRes := runAndReceive(msg, dat)
	if piRes != nil {
		t.Errorf("Expected <nil> preimage, got %v", piRes)
	}

	// Inspect the servers rpc error response message.
	respMsg := authMgr.getSend()
	if respMsg == nil {
		t.Fatalf("no error response")
	}
	resp, _ := respMsg.Response()
	msgErr := resp.Error
	// Code 1, Message about parsing response and invalid type (1 is not response)
	if msgErr.Code != msgjson.RPCParseError {
		t.Errorf("Expected error code %d, got %d", msgjson.RPCParseError, msgErr.Code)
	}
	wantMsgPrefix := "error parsing preimage notification response"
	if !strings.Contains(msgErr.Message, wantMsgPrefix) {
		t.Errorf("Expected error message %q, got %q", wantMsgPrefix, msgErr.Message)
	}

	// 2. empty preimage from client: InvalidPreimage
	msg, _ = msgjson.NewResponse(5, piMsg, nil)
	//lo, pi := newOrder()
	dat = &piData{lo, make(chan *order.Preimage)}
	piRes = runAndReceive(msg, dat)
	if piRes != nil {
		t.Errorf("Expected <nil> preimage, got %v", piRes)
	}

	respMsg = authMgr.getSend()
	if respMsg == nil {
		t.Fatalf("no error response")
	}
	resp, _ = respMsg.Response()
	msgErr = resp.Error
	// 30 invalid preimage length (0 byes)
	if msgErr.Code != msgjson.InvalidPreimage {
		t.Errorf("Expected error code %d, got %d", msgjson.InvalidPreimage, msgErr.Code)
	}
	if !strings.Contains(msgErr.Message, "invalid preimage length") {
		t.Errorf("Expected error message %q, got %q",
			"invalid preimage length (0 bytes)",
			msgErr.Message)
	}

	// 3. correct preimage length, commitment mismatch
	//lo, pi := newOrder()
	lo.Commit = randomCommit() // break the commitment
	dat = &piData{
		ord:      lo,
		preimage: make(chan *order.Preimage),
	}
	piMsg = &msgjson.PreimageResponse{
		Preimage: pi[:],
	}

	msg, _ = msgjson.NewResponse(5, piMsg, nil)
	piRes = runAndReceive(msg, dat)
	if piRes != nil {
		t.Errorf("Expected <nil> preimage, got %v", piRes)
	}

	respMsg = authMgr.getSend()
	if respMsg == nil {
		t.Fatalf("no error response")
	}
	resp, _ = respMsg.Response()
	msgErr = resp.Error
	// 30 invalid preimage length (0 byes)
	if msgErr.Code != msgjson.PreimageCommitmentMismatch {
		t.Errorf("Expected error code %d, got %d",
			msgjson.PreimageCommitmentMismatch, msgErr.Code)
	}
	if !strings.Contains(msgErr.Message, "does not match order commitment") {
		t.Errorf("Expected error message of the form %q, got %q",
			"preimage hash {hash} does not match order commitment {commit}",
			msgErr.Message)
	}

	// 4. correct preimage and commit
	lo.Commit = pi.Commit() // fix the commitment
	dat = &piData{
		ord:      lo,
		preimage: make(chan *order.Preimage),
	}
	piMsg = &msgjson.PreimageResponse{
		Preimage: pi[:],
	}

	piRes = runAndReceive(msg, dat)
	if piRes == nil {
		t.Errorf("Expected preimage %x, got <nil>", pi)
	} else if *piRes != pi {
		t.Errorf("Expected preimage %x, got %x", pi, *piRes)
	}

	// no response this time (no error)
	respMsg = authMgr.getSend()
	if respMsg != nil {
		t.Fatalf("got error response: %d %q", respMsg.Type, string(respMsg.Payload))
	}

	// 5. client classified server request as invalid: InvalidRequestError
	msg, _ = msgjson.NewResponse(5, nil, msgjson.NewError(msgjson.InvalidRequestError, "invalid request"))
	lo, pi = newOrder()
	dat = &piData{lo, make(chan *order.Preimage)}
	piRes = runAndReceive(msg, dat)
	if piRes != nil {
		t.Errorf("Expected <nil> preimage, got %v", piRes)
	}

	// Inspect the servers rpc error response message.
	respMsg = authMgr.getSend()
	if respMsg != nil {
		t.Fatalf("server is not expected to respond with anything")
	}

	// 6. payload is not msgjson.PreimageResponse, unmarshal still succeeds, but PI is nil
	notaPiMsg := new(msgjson.OrderBookSubscription)
	msg, _ = msgjson.NewResponse(5, notaPiMsg, nil)
	dat = &piData{lo, make(chan *order.Preimage)}
	piRes = runAndReceive(msg, dat)
	if piRes != nil {
		t.Errorf("Expected <nil> preimage, got %v", piRes)
	}

	respMsg = authMgr.getSend()
	if respMsg == nil {
		t.Fatalf("no error response")
	}
	resp, _ = respMsg.Response()
	msgErr = resp.Error
	// 30 invalid preimage length (0 byes)
	if msgErr.Code != msgjson.InvalidPreimage {
		t.Errorf("Expected error code %d, got %d", msgjson.InvalidPreimage, msgErr.Code)
	}
	if !strings.Contains(msgErr.Message, "invalid preimage length") {
		t.Errorf("Expected error message %q, got %q",
			"invalid preimage length (0 bytes)",
			msgErr.Message)
	}

	// 7. payload unmarshal error
	msg, _ = msgjson.NewResponse(5, piMsg, nil)
	msg.Payload = json.RawMessage(`{"result":1}`) // ResponsePayload with invalid Result
	dat = &piData{lo, make(chan *order.Preimage)}
	piRes = runAndReceive(msg, dat)
	if piRes != nil {
		t.Errorf("Expected <nil> preimage, got %v", piRes)
	}

	respMsg = authMgr.getSend()
	if respMsg == nil {
		t.Fatalf("no error response")
	}
	resp, _ = respMsg.Response()
	msgErr = resp.Error
	// Code 1, Message about parsing response payload and invalid type (1 is not response)
	if msgErr.Code != msgjson.RPCParseError {
		t.Errorf("Expected error code %d, got %d", msgjson.RPCParseError, msgErr.Code)
	}
	// wrapped json.UnmarshalFieldError
	wantMsgPrefix = "error parsing preimage response payload result"
	if !strings.Contains(msgErr.Message, wantMsgPrefix) {
		t.Errorf("Expected error message %q, got %q", wantMsgPrefix, msgErr.Message)
	}
}

func TestMarket_CancelWhileSuspended(t *testing.T) {
	mkt, storage, auth, cleanup, err := newTestMarket()
	defer cleanup()
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
		return
	}

	auth.handleMatchDone = make(chan *msgjson.Message, 1)
	storage.archivedCancels = make([]*order.CancelOrder, 0, 1)
	storage.canceledOrders = make([]*order.LimitOrder, 0, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Insert a limit order into the book before the market has started
	lo := makeLO(buyer3, mkRate3(1.0, 1.2), 1, order.StandingTiF)
	if !mkt.book.Insert(lo) {
		t.Fatalf("Failed to Insert order into book.")
	}

	// Start the market
	epochDurationMSec := int64(mkt.EpochDuration())
	startEpochIdx := 2 + time.Now().UnixMilli()/epochDurationMSec
	startEpochTime := time.UnixMilli(startEpochIdx * epochDurationMSec)
	go mkt.Start(ctx, startEpochIdx)
	<-time.After(time.Until(startEpochTime.Add(50 * time.Millisecond)))
	if !mkt.Running() {
		t.Fatal("market should be running")
	}

	// Suspend the market, persisting the existing orders
	_, finalTime := mkt.Suspend(time.Now(), true)
	<-time.After(time.Until(finalTime.Add(50 * time.Millisecond)))
	if mkt.Running() {
		t.Fatal("market should not be running")
	}

	if mkt.book.BuyCount() != 1 {
		t.Fatalf("There should be an order in the book.")
	}

	// Submit a valid cancel order.
	loID := lo.ID()
	piCo := test.RandomPreimage()
	commit := piCo.Commit()
	cancelTime := time.Now().UnixMilli()
	aid := buyer3.Acct
	cancelMsg := &msgjson.CancelOrder{
		Prefix: msgjson.Prefix{
			AccountID:  aid[:],
			Base:       dcrID,
			Quote:      btcID,
			OrderType:  msgjson.CancelOrderNum,
			ClientTime: uint64(cancelTime),
			Commit:     commit[:],
		},
		TargetID: loID[:],
	}
	newCancel := func() *order.CancelOrder {
		return &order.CancelOrder{
			P: order.Prefix{
				AccountID:  aid,
				BaseAsset:  lo.Base(),
				QuoteAsset: lo.Quote(),
				OrderType:  order.CancelOrderType,
				ClientTime: time.UnixMilli(cancelTime),
				Commit:     commit,
			},
			TargetOrderID: loID,
		}
	}
	co := newCancel()
	coRecord := orderRecord{
		msgID: 1,
		req:   cancelMsg,
		order: co,
	}
	err = mkt.SubmitOrder(&coRecord)
	if err != nil {
		t.Fatalf("Error submitting cancel order: %v", err)
	}

	if mkt.book.BuyCount() != 0 {
		t.Fatalf("Did not remove order from book.")
	}

	// Make sure that the cancel order was archived, and the limit order was
	// canceled.
	if len(storage.archivedCancels) != 1 {
		t.Fatalf("1 cancel order should be archived but there are %v", len(storage.archivedCancels))
	}
	if !bytes.Equal(storage.archivedCancels[0].ID().Bytes(), co.ID().Bytes()) {
		t.Fatalf("Archived cancel order's ID does not match expected")
	}
	if len(storage.canceledOrders) != 1 {
		t.Fatalf("1 cancel order should be archived but there are %v", len(storage.archivedCancels))
	}
	if !bytes.Equal(storage.canceledOrders[0].ID().Bytes(), lo.ID().Bytes()) {
		t.Fatalf("Cacneled limit order's ID does not match expected")
	}

	// Make sure that we responded to the order request
	if len(auth.sends) != 1 {
		t.Fatalf("There should be 1 send, a response to the order request.")
	}
	msg := auth.sends[0]
	response := new(msgjson.OrderResult)
	msg.UnmarshalResult(response)
	if !bytes.Equal(response.OrderID, co.ID().Bytes()) {
		t.Fatalf("order response sent for the incorrect order ID")
	}

	// Make sure that we sent the match request to the client.
	msg = <-auth.handleMatchDone
	var matches []*msgjson.Match
	err = json.Unmarshal(msg.Payload, &matches)
	if err != nil {
		t.Fatalf("failed to unmarshal match messages")
	}
	if len(matches) != 2 {
		t.Fatalf("There should be 2 payloads, one for maker and taker match each: %v", len(matches))
	}
	var taker, maker bool
	if matches[0].Side == uint8(order.Maker) || matches[1].Side == uint8(order.Maker) {
		maker = true
	}
	if matches[0].Side == uint8(order.Taker) || matches[1].Side == uint8(order.Taker) {
		taker = true
	}
	if !taker || !maker {
		t.Fatalf("There should be 2 payloads, one for maker and taker match each")
	}
}

func TestMarket_NewMarket_AccountBased(t *testing.T) {
	testAccountAssets(t, true, false)
	testAccountAssets(t, false, true)
	testAccountAssets(t, true, true)
}

func testAccountAssets(t *testing.T, base, quote bool) {
	storage := &TArchivist{}
	balancer := newTBalancer()
	const numPerSide = 10
	ords := make([]*order.LimitOrder, 0, numPerSide*2)

	baseAsset, quoteAsset := assetDCR, assetBTC
	if base {
		baseAsset = assetETH
	}
	if quote {
		quoteAsset = assetMATIC
	}

	for i := 0; i < numPerSide*2; i++ {
		writer := test.RandomWriter()
		writer.Market = &test.Market{
			Base:    baseAsset.ID,
			Quote:   quoteAsset.ID,
			LotSize: dcrLotSize,
		}
		writer.Sell = i%2 == 0
		ord := makeLO(writer, mkRate3(0.8, 1.0), randLots(10), order.StandingTiF)
		if (ord.Sell && base) || (!ord.Sell && quote) { // eth-funded order needs a account address coin.
			ord.Coins = []order.CoinID{[]byte(test.RandomAddress())}
		}
		ords = append(ords, ord)
		storage.BookOrder(ord)
	}

	_, _, _, cleanup, err := newTestMarket(storage, balancer, [2]*asset.BackedAsset{baseAsset, quoteAsset})
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
	}
	defer cleanup()

	for _, lo := range ords {
		if base && balancer.reqs[lo.BaseAccount()] == 0 {
			t.Fatalf("base balance not requested for order")
		}
		if quote && balancer.reqs[lo.QuoteAccount()] == 0 {
			t.Fatalf("quote balance not requested for order")
		}
	}
}

func TestMarket_AccountPending(t *testing.T) {
	storage := &TArchivist{}
	writer := test.RandomWriter()
	writer.Market = &test.Market{
		Base:    assetETH.ID,
		Quote:   assetMATIC.ID,
		LotSize: dcrLotSize,
	}

	const rate = btcRateStep * 100
	const sellLots = 10
	const buyLots = 20
	ethAddr := test.RandomAddress()
	maticAddr := test.RandomAddress()

	writer.Sell = true
	lo := makeLO(writer, rate, sellLots, order.StandingTiF)
	lo.Coins = []order.CoinID{[]byte(ethAddr)}
	lo.Address = maticAddr
	storage.BookOrder(lo)

	writer.Sell = false
	lo = makeLO(writer, rate, buyLots, order.StandingTiF)
	lo.Coins = []order.CoinID{[]byte(maticAddr)}
	lo.Address = ethAddr
	storage.BookOrder(lo)

	mkt, _, _, cleanup, err := newTestMarket(storage, newTBalancer(), [2]*asset.BackedAsset{assetETH, assetMATIC})
	if err != nil {
		t.Fatalf("newTestMarket failure: %v", err)
	}
	defer cleanup()

	checkPending := func(tag string, addr string, assetID uint32, expQty, expLots uint64, expRedeems int) {
		t.Helper()
		qty, lots, redeems := mkt.AccountPending(addr, assetID)
		if qty != expQty {
			t.Fatalf("%s: wrong quantity: wanted %d, got %d", tag, expQty, qty)
		}
		if lots != expLots {
			t.Fatalf("%s: wrong lots: wanted %d, got %d", tag, expLots, lots)
		}
		if redeems != expRedeems {
			t.Fatalf("%s: wrong redeems: wanted %d, got %d", tag, expRedeems, redeems)
		}
	}

	checkPending("booked-only-eth", ethAddr, assetETH.ID, sellLots*dcrLotSize, sellLots, buyLots)

	quoteQty := calc.BaseToQuote(rate, buyLots*dcrLotSize)
	checkPending("booked-only-matic", maticAddr, assetMATIC.ID, quoteQty, buyLots, sellLots)

	const epochSellLots = 5
	writer.Sell = true
	lo = makeLO(writer, rate, epochSellLots, order.StandingTiF)
	lo.Coins = []order.CoinID{[]byte(ethAddr)}
	lo.Address = maticAddr
	mkt.epochOrders[lo.ID()] = lo
	const totalSellLots = sellLots + epochSellLots
	checkPending("with-epoch-sell-eth", ethAddr, assetETH.ID, totalSellLots*dcrLotSize, totalSellLots, buyLots)
	checkPending("with-epoch-sell-matic", maticAddr, assetMATIC.ID, quoteQty, buyLots, totalSellLots)

	// Market buy order.
	midGap := mkt.MidGap()
	mktBuyQty := quoteQty + calc.BaseToQuote(midGap, dcrLotSize/2)
	writer.Sell = false
	mo := makeMO(writer, 0)
	mo.Quantity = mktBuyQty
	mo.Coins = []order.CoinID{[]byte(maticAddr)}
	mo.Address = ethAddr
	mkt.epochOrders[mo.ID()] = mo
	redeems := int(totalSellLots)
	totalBuyLots := buyLots + calc.QuoteToBase(midGap, mktBuyQty)/dcrLotSize
	totalQty := quoteQty + mktBuyQty
	checkPending("with-epoch-market-buy-matic", maticAddr, assetMATIC.ID, totalQty, totalBuyLots, redeems)
	checkPending("with-epoch-market-buy-eth", ethAddr, assetETH.ID, totalSellLots*dcrLotSize, totalSellLots, int(totalBuyLots))
}
