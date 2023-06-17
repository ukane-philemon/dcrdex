package eth

import (
	"fmt"
	"path/filepath"

	"decred.org/dcrdex/dex"
	dexeth "decred.org/dcrdex/dex/networks/eth"
	dexpolygon "decred.org/dcrdex/dex/networks/polygon"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
)

// genesisConfig is a map of chain ID to genesis block. The simnet genesis block
// config is not present in this map because it can be read from the simnet test
// directory when it is needed.
var genesisConfig = map[int64]*core.Genesis{
	// Ethereum
	dexeth.TestnetChainID: core.DefaultGoerliGenesisBlock(),
	dexeth.MainnetChainID: core.DefaultGenesisBlock(),

	// Polygon
	dexpolygon.MainnetChainID: dexpolygon.DefaultBorMainnetGenesisBlock(),
	dexpolygon.TestnetChainID: dexpolygon.DefaultMumbaiGenesisBlock(),
}

// chainConfig returns chain config for Ethereum and Ethereum compatible chains
// (Polygon).
func chainConfig(chainID int64, network dex.Network) (c ethconfig.Config, err error) {
	if network == dex.Simnet {
		dataDir, err := simnetDataDir(chainID)
		if err != nil {
			return c, err
		}

		genesisFile := filepath.Join(dataDir, "genesis.json")
		genesis, err := dexeth.LoadGenesisFile(genesisFile)
		if err != nil {
			return c, fmt.Errorf("error reading genesis file: %v", err)
		}

		c.Genesis = genesis
		c.NetworkId = genesis.Config.ChainID.Uint64()
		return c, nil
	}

	genesisCfg, ok := genesisConfig[chainID]
	if !ok {
		return c, fmt.Errorf("unknown chain ID: %d", chainID)
	}

	ethCfg := ethconfig.Defaults
	ethCfg.Genesis = genesisCfg
	ethCfg.NetworkId = genesisCfg.Config.ChainID.Uint64()
	return ethCfg, nil
}
