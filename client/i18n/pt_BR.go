package i18n

var ptBR = map[string]*Translation{
	Native:             {Value: "Nativo"},
	NativeWalletDesc:   {Value: "Use a carteira SPV integrada"},
	ElectrumWallet:     {Value: "Electro (externo)"},
	ElectrumWalletDesc: {Value: "Use uma carteira Electrum externa"},
	External:           {Value: "Externo"},
	ConnectToBitcoind:  {Value: "Conecte-se ao bitcoind"},

	// WALLET CONFIG OPTIONS
	RPCConfigUserDisplayName:       {Value: "Nome de usuário JSON-RPC"},
	RPCConfigUserDescTemplate:      {Value: "%s's configuração 'rpcuser'", Docs: "[walletname]"},
	RPCConfigPasswordDescTemplate:  {Value: "Senha JSON-RPC"},
	RPCConfigPasswordDisplayName:   {Value: "%s's configuração 'rpcpassword'", Docs: "[walletname]"},
	RPCConfigPortDisplayName:       {Value: "Porta JSON-RPC"},
	RPCConfigPortDescTemplate:      {Value: "%s's configurações 'rpcport' (se não forem definidas com rpcbind)", Docs: "[walletname]"},
	RPCConfigAddressDisplayName:    {Value: "Endereço JSON-RPC"},
	RPCConfigAddressDescTemplate:   {Value: "%s's 'rpchost' <addr> or <addr>:<port> (default: %s)", Docs: "[walletname, default host]"},
	RPCConfigWalletFileDescription: {Value: "Caminho completo para o arquivo da carteira (sem padrão)"},

	// OTHER TRANSLATIONs
	WalletFile:     {Value: "Arquivo de Carteira"},
	WalletName:     {Value: "Nome da Carteira"},
	WalletNameDesc: {Value: "O nome da carteira"},
}
