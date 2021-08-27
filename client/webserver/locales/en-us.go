package locales

var EnUS = map[string]string{
	"Markets":                        "Markets",
	"Wallets":                        "Wallets",
	"Notifications":                  "Notifications",
	"Recent Activity":                "Recent Activity",
	"Sign Out":                       "Sign Out",
	"Order History":                  "Order History",
	"load from file":                 "load from file",
	"loaded from file":               "loaded from file",
	"defaults":                       "defaults",
	"Wallet Password":                "Wallet Password",
	"w_password_helper":              "This is the password you have configured with your wallet software.",
	"w_password_tooltip":             "Leave the password empty if there is no password required for the wallet.",
	"App Password":                   "App Password",
	"app_password_helper":            "Your app password is always required when performing sensitive wallet operations.",
	"Add":                            "Add",
	"Unlock":                         "Unlock",
	"Wallet":                         "Wallet",
	"app_password_reminder":          "Your app password is always required when performing sensitive wallet operations.",
	"DEX Address":                    "DEX Address",
	"TLS Certificate":                "TLS Certificate",
	"remove":                         "remove",
	"add a file":                     "add a file",
	"Submit":                         "Submit",
	"Confirm Registration":           "Confirm Registration",
	"app_pw_reg":                     "Enter your app password to confirm DEX registration.",
	"reg_confirm_submit":             `When you submit this form, <span id="feeDisplay"></span> DCR will be spent from your Decred wallet to pay registration fees.`,
	"provied_markets":                "This DEX provides the following markets:",
	"base_header":                    "Base",
	"quote_header":                   "Quote",
	"lot_size_header":                "Lot Size",
	"lot_size_headsup":               "All trades are in multiples of the lot size.",
	"Password":                       "Password",
	"Register":                       "Register",
	"Authorize Export":               "Authorize Export",
	"export_app_pw_msg":              "Enter your app password to confirm Account export for",
	"Disable Account":                "Disable Account",
	"disable_app_pw_msg":             "Enter your app password to disable account",
	"disable_irreversible":           `<span class="red">Note:</span> This action is irreversible - once an account is disabled it can't be re-enabled.`,
	"Authorize Import":               "Authorize Import",
	"app_pw_import_msg":              "Enter your app password to confirm Account import",
	"Account File":                   "Account File",
	"Change Application Password":    "Change Application Password",
	"Current Password":               "Current Password",
	"New Password":                   "New Password",
	"Confirm New Password":           "Confirm New Password",
	"Cancel Order":                   "Cancel Order",
	"cancel_pw":                      "Enter your password to submit a cancel order for the remaining",
	"cancel_no_pw":                   "Submit a cancel order for the remaining",
	"cancel_remain":                  "The remaining amount may change before the cancel order is matched.",
	"Log In":                         "Log In",
	"epoch":                          "epoch",
	"price":                          "price",
	"volume":                         "volume",
	"buys":                           "buys",
	"Buy Orders":                     "Buy Orders",
	"Quantity":                       "Quantity",
	"Rate":                           "Rate",
	"Epoch":                          "Epoch",
	"Limit Order":                    "Limit Order",
	"Market Order":                   "Market Order",
	"reg_status_msg":                 `In order to trade at <span id="regStatusDex"></span>, the registration fee payment needs <span id="confReq"></span> confirmations.`,
	"Buy":                            "Buy",
	"Sell":                           "Sell",
	"Lot Size":                       "Lot Size",
	"Rate Step":                      "Rate Step",
	"Max":                            "Max",
	"lot":                            "lot",
	"Price":                          "Price",
	"Lots":                           "Lots",
	"min trade is about":             "min trade is about",
	"immediate_explanation":          "If the order doesn't fully match during the next match cycle, any unmatched quantity will not be booked or matched again. Taker-only order.",
	"Immediate or cancel":            "Immediate or cancel",
	"Balances":                       "Balances",
	"outdated_tooltip":               "Balance may be outdated. Connect to the wallet to refresh.",
	"available":                      "available",
	"connect_refresh_tooltip":        "Click to connect and refresh",
	"add_a_base_wallet":              `Add a<br><span data-unit="base"></span><br>wallet`,
	"add_a_quote_wallet":             `Add a<br><span data-unit="quote"></span><br>wallet`,
	"locked":                         "locked",
	"immature":                       "immature",
	"Sell Orders":                    "Sell Orders",
	"Your Orders":                    "Your Orders",
	"Type":                           "Type",
	"Side":                           "Side",
	"Age":                            "Age",
	"Filled":                         "Filled",
	"Settled":                        "Settled",
	"Status":                         "Status",
	"view order history":             "view order history",
	"cancel order":                   "cancel order",
	"order details":                  "order details",
	"verify_order":                   `Verify <span id="vSideHeader"></span>  Order`,
	"You are submitting an order to": "You are submitting an order to",
	"at a rate of":                   "at a rate of",
	"for a total of":                 "for a total of",
	"verify_market":                  "This is a market order and will match the best available order(s) on the book. Based on the current market mid-gap rate, you might receive about",
	"auth_order_app_pw":              "Authorize this order with your app password.",
	"lots":                           "lots",
	"order_disclaimer": `<span class="red">IMPORTANT</span>: Trades take time to settle, and you cannot turn off the DEX client software,
		or the <span data-unit="quote"></span> or <span data-unit="base"></span> blockchain and/or wallet software, until
		settlement is complete. Settlement can complete as quickly as a few minutes or take as long as a few hours.`,
	"Order":                     "Order",
	"see all orders":            "see all orders",
	"Exchange":                  "Exchange",
	"Market":                    "Market",
	"Offering":                  "Offering",
	"Asking":                    "Asking",
	"Fees":                      "Fees",
	"order_fees_tooltip":        "On-chain transaction fees, typically collected by the miner. Decred DEX collects no trading fees.",
	"Matches":                   "Matches",
	"Match ID":                  "Match ID",
	"Time":                      "Time",
	"ago":                       "ago",
	"Cancellation":              "Cancellation",
	"Order Portion":             "Order Portion",
	"Swap":                      "", // Label for the first on-chain transaction cast by each party
	"you":                       "you",
	"them":                      "them",
	"Redemption":                "Redemption",
	"Refund":                    "Refund",
	"Funding Coins":             "Funding Coins",
	"Exchanges":                 "Exchanges",
	"apply":                     "apply",
	"Assets":                    "Assets",
	"Trade":                     "Trade",
	"Set App Password":          "Set App Password",
	"reg_set_app_pw_msg":        "Set your app password. This password will protect your DEX account keys and connected wallets.",
	"Password Again":            "Password Again",
	"reg_dcr_required":          "Your Decred wallet is required to pay registration fees.",
	"reg_dcr_unlocked":          "Unlock your Decred wallet to pay registration fees.",
	"Add a DEX":                 "Add a DEX",
	"reg_ssl_needed":            "Looks like we don't have an SSL certificate for this DEX. Add the server's certificate to continue.",
	"Dark Mode":                 "Dark Mode",
	"Show pop-up notifications": "Show pop-up notifications",
	"Account ID":                "Account ID",
	"Export Account":            "Export Account",
	"simultaneous_servers_msg":  "The Decred DEX Client supports simultaneous use of any number of DEX servers.",
	"Change App Password":       "Change App Password",
	"Build ID":                  "Build ID",
	"Connect":                   "Connect",
	"Withdraw":                  "Withdraw",
	"Deposit":                   "Deposit",
	"Lock":                      "Lock",
	"Create a":                  "Create a",
	"New Deposit Address":       "New Deposit Address",
	"Address":                   "Address",
	"Amount":                    "Amount",
	"Authorize the withdraw with your app password.": "Authorize the withdraw with your app password.",
	"Reconfigure":            "Reconfigure",
	"pw_change_instructions": "Changing the password below does not change the password for your wallet software. Use this form to update the DEX client after you have changed your password with the wallet software directly.",
	"New Wallet Password":    "New Wallet Password",
	"pw_change_warn":         "Note: Changing to a different wallet while having active trades might cause funds to be lost.",
	"Show more options":      "Show more options",
	"seed_implore_msg":       "You should carefully write down your application seed and save a copy. Should you lose access to this machine or the critical application files, the seed can be used to restore your DEX accounts and native wallets. Some older accounts cannot be restored from seed, and whether old or new, it's good practice to backup the account keys separately from the seed.",
	"View Application Seed":  "View Application Seed",
	"Remember my password":   "Remember my password",
	"pw_for_seed":            "Enter your app password to show your seed. Make sure nobody else can see your screen.",
	"Asset":                  "Asset",
	"Balance":                "Balance",
	"Actions":                "Actions",
	"Restoration Seed":       "Restoration Seed",
	"Restore from seed":      "Restore from seed",
	"Import Account":         "Import Account",
}
