package i18n

var enUS = map[string]*Translation{
	// CORE NOTIFICATIONS
	"Topic_Seed_Needs_Saving_Subject":          {Value: "Don't forget to back up your application seed"},
	"Topic_Seed_Needs_Saving_Template":         {Value: "A new application seed has been created. Make a back up now in the settings view."},
	"Topic_Upgraded_To_Seed_Subject":           {Value: "Back up your new application seed"},
	"Topic_Upgraded_To_Seed_Template":          {Value: "The client has been upgraded to use an application seed. Back up the seed now in the settings view"},
	"Topic_Fee_Payment_In_Progress_Subject":    {Value: "Fee payment in progress"},
	"Topic_Fee_Payment_In_Progress_Template":   {Value: "Waiting for %d confirmations before trading at %s", Docs: "[confs, host]"},
	"Topic_Fee_Payment_Error_Subject":          {Value: "Fee payment error"},
	"Topic_Fee_Payment_Error_Template":         {Value: "Error encountered while paying fees to %s: %v", Docs: "[host, error]"},
	"Topic_Fee_Coin_Error_Subject":             {Value: "Fee coin error"},
	"Topic_Fee_Coin_Error_Template":            {Value: "Empty fee coin for %s.", Docs: "[host]"},
	"Topic_Reg_Update_Subject":                 {Value: "Registration update"},
	"Topic_Reg_Update_Template":                {Value: "Fee payment confirmations %v/%v", Docs: "[confs, required confs]"},
	"Topic_Account_Registered_Subject":         {Value: "Account registered"},
	"Topic_Account_Registered_Template":        {Value: "You may now trade at %s", Docs: "[host]"},
	"Topic_Account_Unlock_Error_Subject":       {Value: "Account unlock error"},
	"Topic_Account_Unlock_Error_Template":      {Value: "Error unlocking account for %s: %v", Docs: "[host, error]"},
	"Topic_Wallet_Connection_Warning_Subject":  {Value: "Wallet connection warning"},
	"Topic_Wallet_Connection_Warning_Template": {Value: "Incomplete registration detected for %s, but failed to connect to the Decred wallet", Docs: "[host]"},
	"Topic_Wallet_Unlock_Error_Subject":        {Value: "Wallet unlock error"},
	"Topic_Wallet_Unlock_Error_Template":       {Value: "Connected to wallet to complete registration at %s, but failed to unlock: %v", Docs: "[host, error]"},
	"Topic_Wallet_Comms_Warning_Subject":       {Value: "Wallet connection issue"},
	"Topic_Wallet_Comms_Warning_Template":      {Value: "Unable to communicate with %v wallet! Reason: %v", Docs: "[asset name, error message]"},
	"Topic_Wallet_Peers_Restored_Subject":      {Value: "Wallet connectivity restored"},
	"Topic_Wallet_Peers_Restored_Template":     {Value: "%v wallet has reestablished connectivity.", Docs: "[asset name]"},

	// WALLET DEFINITIONS
	Native:             {Value: "Native"},
	NativeWalletDesc:   {Value: "Use the built-in SPV wallet"},
	ElectrumWallet:     {Value: "Electrum (external)"},
	ElectrumWalletDesc: {Value: "Use an External Electrum wallet"},
	External:           {Value: "External"},
	ConnectToBitcoind:  {Value: "Connect to bitcoind"},

	// WALLET CONFIG OPTIONS
	RPCConfigUserDisplayName:       {Value: "JSON-RPC Username"},
	RPCConfigUserDescTemplate:      {Value: "%s's 'rpcuser' setting", Docs: "[walletname]"},
	RPCConfigPasswordDisplayName:   {Value: "JSON-RPC Password"},
	RPCConfigPasswordDescTemplate:  {Value: "%s's 'rpcpassword' setting", Docs: "[walletname]"},
	RPCConfigPortDisplayName:       {Value: "JSON-RPC Port"},
	RPCConfigPortDescTemplate:      {Value: "%s's 'rpcport' (if not set with rpcbind)", Docs: "[walletname]"},
	RPCConfigAddressDisplayName:    {Value: "JSON-RPC Address"},
	RPCConfigAddressDescTemplate:   {Value: "%s's 'rpchost' <addr> or <addr>:<port> (default: %s)", Docs: "[walletname, default host]"},
	RPCConfigWalletFileDescription: {Value: "Full path to the wallet file (no default)"},

	// OTHER TRANSLATIONs
	WalletFile:              {Value: "Wallet File"},
	WalletName:              {Value: "Wallet Name"},
	WalletNameDesc:          {Value: "The Wallet Name"},
	"Markets":               {Value: "Markets"},
	"Wallets":               {Value: "Wallets"},
	"Notifications":         {Value: "Notifications"},
	"Recent Activity":       {Value: "Recent Activity"},
	"Sign Out":              {Value: "Sign Out"},
	"Order History":         {Value: "Order History"},
	"load from file":        {Value: "load from file"},
	"loaded from file":      {Value: "loaded from file"},
	"defaults":              {Value: "defaults"},
	"Wallet Password":       {Value: "Wallet Password"},
	"w_password_helper":     {Value: "This is the password you have configured with your wallet software."},
	"w_password_tooltip":    {Value: "Leave the password empty if there is no password required for the wallet."},
	"App Password":          {Value: "App Password"},
	"Add":                   {Value: "Add"},
	"Unlock":                {Value: "Unlock"},
	"Rescan":                {Value: "Rescan"},
	"Wallet":                {Value: "Wallet"},
	"app_password_reminder": {Value: "Your app password is always required when performing sensitive wallet operations."},
	"DEX Address":           {Value: "DEX Address"},
	"TLS Certificate":       {Value: "TLS Certificate"},
	"remove":                {Value: "remove"},
	"add a file":            {Value: "add a file"},
}
