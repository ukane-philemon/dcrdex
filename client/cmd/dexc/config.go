package main

import "decred.org/dcrdex/client/app"

func configure() (*app.Config, error) {
	// Pre-parse the command line options to see if an alternative config file
	// or the version flag was specified. Override any environment variables
	// with parsed command line flags.
	iniCfg := app.DefaultConfig
	preCfg := iniCfg
	if err := app.ParseCLIConfig(&preCfg); err != nil {
		return nil, err
	}
	appData, configPath := app.ResolveCLIConfigPaths(&preCfg)

	// Load additional config from file.
	if err := app.ParseFileConfig(configPath, &iniCfg); err != nil {
		return nil, err
	}

	// Set the global *Config.
	cfg := &iniCfg
	return cfg, app.ResolveConfig(appData, cfg)
}
