// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"

	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/encode"
	"decred.org/dcrdex/server/admin"
	_ "decred.org/dcrdex/server/asset/importall"
	dexsrv "decred.org/dcrdex/server/dex"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func mainCore(ctx context.Context) error {
	// Parse the configuration file, and setup logger.
	cfg, opts, err := loadConfig()
	if err != nil {
		fmt.Printf("Failed to load dcrdex config: %s\n", err.Error())
		return err
	}
	defer func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}()

	if cfg.ValidateMarkets {
		return dexsrv.ValidateConfigFile(cfg.MarketsConfPath, cfg.Network, log.SubLogger("V"))
	}

	// Request admin server password if admin server is enabled and
	// server password is not set in config.
	var adminSrvAuthSHA [32]byte
	if cfg.AdminSrvOn {
		if len(cfg.AdminSrvPW) == 0 {
			adminSrvAuthSHA, err = admin.PasswordHashPrompt(ctx, "Admin interface password: ")
			if err != nil {
				return fmt.Errorf("cannot use password: %v", err)
			}
		} else {
			adminSrvAuthSHA = sha256.Sum256(cfg.AdminSrvPW)
			encode.ClearBytes(cfg.AdminSrvPW)
		}
	}

	if opts.CPUProfile != "" {
		var f *os.File
		f, err = os.Create(opts.CPUProfile)
		if err != nil {
			return err
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}

	// HTTP profiler
	if opts.HTTPProfile {
		log.Warnf("Starting the HTTP profiler on path /debug/pprof/.")
		// http pprof uses http.DefaultServeMux
		http.Handle("/", http.RedirectHandler("/debug/pprof/", http.StatusSeeOther))
		go func() {
			if err := http.ListenAndServe(":9232", nil); err != nil {
				log.Errorf("ListenAndServe failed for http/pprof: %v", err)
			}
		}()
	}

	// Display app version.
	log.Infof("%s version %v (Go version %s)", appName, Version, runtime.Version())
	log.Infof("dcrdex starting for network: %s", cfg.Network)
	log.Infof("swap locktimes config: maker %s, taker %s",
		dex.LockTimeMaker(cfg.Network), dex.LockTimeTaker(cfg.Network))

	// Load the market and asset configurations for the given network.
	markets, assets, err := dexsrv.LoadConfig(cfg.Network, cfg.MarketsConfPath)
	if err != nil {
		return fmt.Errorf("failed to load market and asset config %q: %v",
			cfg.MarketsConfPath, err)
	}
	log.Infof("Found %d assets, loaded %d markets, for network %s",
		len(assets), len(markets), strings.ToUpper(cfg.Network.String()))
	// NOTE: If MaxUserCancelsPerEpoch is ultimately a setting we want to keep,
	// bake it into the markets.json file and load it per-market in settings.go.
	// For now, patch it into each dex.MarketInfo.
	for _, mkt := range markets {
		mkt.MaxUserCancelsPerEpoch = cfg.MaxUserCancels
	}

	// Load, or create and save, the DEX signing key.
	var privKey *secp256k1.PrivateKey
	if len(cfg.SigningKeyPW) == 0 {
		cfg.SigningKeyPW, err = admin.PasswordPrompt(ctx, "Signing key password: ")
		if err != nil {
			return fmt.Errorf("cannot use password: %v", err)
		}
	}
	privKey, err = dexKey(cfg.DEXPrivKeyPath, cfg.SigningKeyPW)
	encode.ClearBytes(cfg.SigningKeyPW)
	if err != nil {
		return err
	}

	// Create the DEX manager.
	dexConf := &dexsrv.DexConf{
		DataDir:    cfg.DataDir,
		LogBackend: cfg.LogMaker,
		Markets:    markets,
		Assets:     assets,
		Network:    cfg.Network,
		DBConf: &dexsrv.DBConf{
			DBName:       cfg.DBName,
			Host:         cfg.DBHost,
			User:         cfg.DBUser,
			Port:         cfg.DBPort,
			Pass:         cfg.DBPass,
			ShowPGConfig: cfg.ShowPGConfig,
		},
		BroadcastTimeout: cfg.BroadcastTimeout,
		TxWaitExpiration: cfg.TxWaitExpiration,
		CancelThreshold:  cfg.CancelThreshold,
		FreeCancels:      cfg.FreeCancels,
		PenaltyThreshold: cfg.PenaltyThreshold,
		DEXPrivKey:       privKey,
		CommsCfg: &dexsrv.RPCConfig{
			RPCCert:           cfg.RPCCert,
			NoTLS:             cfg.NoTLS,
			RPCKey:            cfg.RPCKey,
			ListenAddrs:       cfg.RPCListen,
			AltDNSNames:       cfg.AltDNSNames,
			DisableDataAPI:    cfg.DisableDataAPI,
			HiddenServiceAddr: cfg.HiddenService,
		},
		NoResumeSwaps: cfg.NoResumeSwaps,
		NodeRelayAddr: cfg.NodeRelayAddr,
	}
	dexMan, err := dexsrv.NewDEX(ctx, dexConf) // ctx cancel just aborts setup; Stop does normal shutdown
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	if cfg.AdminSrvOn {
		srvCFG := &admin.SrvConfig{
			Core:    dexMan,
			Addr:    cfg.AdminSrvAddr,
			AuthSHA: adminSrvAuthSHA,
			Cert:    cfg.RPCCert,
			Key:     cfg.RPCKey,
			NoTLS:   cfg.AdminSrvNoTLS,
		}
		adminServer, err := admin.NewServer(srvCFG)
		if err != nil {
			return fmt.Errorf("cannot set up admin server: %v", err)
		}
		wg.Add(1)
		go func() {
			adminServer.Run(ctx)
			wg.Done()
		}()
	}

	log.Info("The DEX is running. Hit CTRL+C to quit...")
	<-ctx.Done()
	// Wait for the admin server to finish.
	wg.Wait()

	log.Info("Stopping DEX...")
	dexMan.Stop()
	log.Info("Bye!")

	return nil
}

func main() {
	// Create a context that is canceled when a shutdown request is received.
	ctx := withShutdownCancel(context.Background())
	// Listen for both interrupt signals (e.g. CTRL+C) and shutdown requests
	// (requestShutdown calls).
	go shutdownListener()

	err := mainCore(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
