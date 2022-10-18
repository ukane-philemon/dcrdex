package locales

var PlPL = map[string]string{
	"Language":                       "pl-PL",
	"Markets":                        "Rynki",
	"Wallets":                        "Portfele",
	"Notifications":                  "Powiadomienia",
	"Recent Activity":                "Ostatnia aktywność",
	"Sign Out":                       "Wyloguj się",
	"Order History":                  "Historia zleceń",
	"load from file":                 "wczytaj z pliku",
	"loaded from file":               "wczytano z pliku",
	"defaults":                       "domyślne",
	"Wallet Password":                "Hasło portfela",
	"w_password_helper":              "To hasło, które zostało skonfigurowane w oprogramowaniu Twojego portfela.",
	"w_password_tooltip":             "Zostaw puste pole, jeśli portfel nie wymaga użycia hasła.",
	"App Password":                   "Hasło aplikacji",
	"Add":                            "Dodaj",
	"Unlock":                         "Odblokuj",
	"Wallet":                         "Portfel",
	"app_password_reminder":          "Hasło aplikacji jest wymagane zawsze wtedy, gdy wykonuje się wrażliwe operacje z użyciem portfela.",
	"DEX Address":                    "Adres DEX",
	"TLS Certificate":                "Certyfikat TLS",
	"remove":                         "usuń",
	"add a file":                     "dodaj plik",
	"Submit":                         "Wyślij",
	"Confirm Registration":           "Potwierdź rejestrację",
	"app_pw_reg":                     "Podaj hasło aplikacji, aby potwierdzić rejestrację na DEX.",
	"reg_confirm_submit":             `Po wysłaniu tego formularza środki z Twojego portfela zostaną wykorzystane do zapłacenia opłaty rejestracyjnej.`,
	"provided_markets":               "Ten DEX obsługuje następujące rynki:",
	"accepted_fee_assets":            "Ten DEX przyjmuje następujące opłaty:",
	"base_header":                    "Waluta bazowa",
	"quote_header":                   "Waluta kwotowana",
	"lot_size_headsup":               "Wszystkie wymiany przeprowadzane są w zakresie wielokrotności rozmiaru lotu.",
	"Password":                       "Hasło",
	"Register":                       "Zarejestruj",
	"Authorize Export":               "Zatwierdź eksport",
	"export_app_pw_msg":              "Podaj hasło aplikacji, aby potwierdzić eksport konta dla",
	"Disable Account":                "Zablokuj konto",
	"disable_app_pw_msg":             "Podaj hasło aplikacji, aby zablokowac konto",
	"disable_dex_server":             "Ten serwer DEX może zostać ponownie włączony (nie będziesz musiał płacić opłaty) w dowolnym momencie w przyszłości, dodając go ponownie.",
	"Authorize Import":               "Zatwierdź import",
	"app_pw_import_msg":              "Podaj hasło aplikacji, aby potwierdzić import konta",
	"Account File":                   "Plik konta",
	"Change Application Password":    "Zmień hasło aplikacji",
	"Current Password":               "Obecne hasło",
	"New Password":                   "Nowe hasło",
	"Confirm New Password":           "Potwierdź nowe hasło",
	"cancel_pw":                      "Podaj hasło, aby wysłać zlecenie anulowania pozostałej sumy",
	"cancel_no_pw":                   "Wyślij zlecenie anulowania pozostałej sumy",
	"cancel_remain":                  "Pozostałą suma może ulec zmianie, zanim zlecenie anulowania zostanie wykonane.",
	"Log In":                         "Zaloguj się",
	"epoch":                          "epoka",
	"price":                          "cena",
	"volume":                         "wolumen",
	"buys":                           "kupno",
	"sells":                          "sprzedaż",
	"Buy Orders":                     "Zlecenia kupna",
	"Quantity":                       "Ilość",
	"Rate":                           "Kurs",
	"Limit Order":                    "Zlecenie oczekujące (limit)",
	"Market Order":                   "Zlecenie rynkowe (market)",
	"reg_status_msg":                 `Żeby rozpocząć handel na <span id="regStatusDex" class="text-break"></span>, opłata rejestracyjna potrzebuje <span id="confReq"></span> potwierdzeń.`,
	"Buy":                            "Kupno",
	"Sell":                           "Sprzedaż",
	"lot_size":                       "Rozmiar lotu",
	"Rate Step":                      "Różnica ceny zlecenia",
	"Max":                            "Max",
	"lot":                            "lot",
	"min trade is about":             "minimalna wielkość zamówienia wynosi około",
	"immediate_explanation":          "Jeśli zamówienie nie zostanie w pełni spasowane w następnym cyklu, pozostałe środki nie będą wystawiane ani pasowane ponownie. Zlecenie typu 'taker-only'.",
	"Immediate or cancel":            "Immediate or cancel",
	"Balances":                       "Salda",
	"outdated_tooltip":               "Saldo może być nieaktualne. Połącz się z portfelem, aby je odświeżyć.",
	"available":                      "dostępne",
	"connect_refresh_tooltip":        "Kliknij, aby połączyć i odświeżyć",
	"add_a_wallet":                   `Dodaj portfel <span data-tmpl="addWalletSymbol"></span> `,
	"locked":                         "zablokowane",
	"immature":                       "niedojrzałe",
	"Sell Orders":                    "Zlecenia sprzedaży",
	"Your Orders":                    "Twoje zlecenia",
	"Type":                           "Rodzaj",
	"Side":                           "Strona",
	"Age":                            "Wiek",
	"Filled":                         "Zrealizowane",
	"Settled":                        "Rozliczone",
	"Status":                         "Status",
	"view order history":             "wyświetl historię zleceń",
	"cancel_order":                   "anuluj zlecenie",
	"order details":                  "szczegóły zlecenia",
	"verify_order":                   `Zweryfikuj zlecenie <span id="vSideHeader"></span>`,
	"You are submitting an order to": "Wysyłasz zlecenie",
	"at a rate of":                   "po kursie",
	"for a total of":                 "na sumę",
	"verify_market":                  "Zlecenie rynkowe, które spasowywane jest z najlepszymi dostępnymi zleceniami w księdze zleceń. W oparciu o obecny kurs między stronami może otrzymać około",
	"auth_order_app_pw":              "Potwierdź to zlecenie swoim hasłem aplikacji.",
	"lots":                           "loty(ów)",
	"order_disclaimer": `<span class="red">UWAGA</span>: Wymiany potrzebują czasu, aby zostać rozliczone. NIE WYŁĄCZAJ swojego klienta DEX, ani oprogramowania, czy portfela blockchaina dla
		<span data-unit="quote"></span> lub <span data-unit="base"></span>, dopóki
		rozliczenie nie zostanie dokonane. Rozliczenie może trwać od kilku minut, aż do kilku godzin.`,
	"Order":                       "Zlecenie",
	"see all orders":              "wyświetl wszystkie zlecenia",
	"Exchange":                    "Giełda",
	"Market":                      "Rynek",
	"Offering":                    "Oferta",
	"Asking":                      "W zamian za",
	"Fees":                        "Opłaty",
	"order_fees_tooltip":          "Opłaty transakcyjne on-chain, zazwyczaj zbierane przez górników. Decred DEX nie pobiera żadnych opłat handlowych.",
	"Matches":                     "Spasowane zlecenia",
	"Match ID":                    "ID zlecenia",
	"Time":                        "Czas",
	"ago":                         "temu",
	"Cancellation":                "Anulowano",
	"Order Portion":               "Część zlecenia",
	"you":                         "Ty",
	"them":                        "oni",
	"Redemption":                  "Wykupienie środków",
	"Refund":                      "Zwrot środków",
	"Funding Coins":               "Środki fundujące zlecenie",
	"Exchanges":                   "Giełdy",
	"apply":                       "zastosuj",
	"Assets":                      "Aktywa",
	"Trade":                       "Wymiana",
	"Set App Password":            "Ustaw hasło aplikacji",
	"reg_set_app_pw_msg":          "Ustaw swoje hasło aplikacji. To hasło będzie chronić klucze Twojego konta DEX oraz połączone z nim portfele.",
	"Password Again":              "Wprowadź hasło ponownie",
	"Add a DEX":                   "Dodaj DEX",
	"reg_ssl_needed":              "Wygląda na to, że nie posiadamy certyfikatu SSL dla tego DEXa. Dodaj certyfikat serwera, aby przejść dalej.",
	"Dark Mode":                   "Tryb ciemny",
	"Show pop-up notifications":   "Pokazuj powiadomienia w okienkach",
	"Account ID":                  "ID konta",
	"Export Account":              "Eksportuj konto",
	"simultaneous_servers_msg":    "Klient Decred DEX wspiera jednoczesne korzystanie z wielu serwerów DEX.",
	"Change App Password":         "Zmień hasło aplikacji",
	"Build ID":                    "ID builda",
	"Connect":                     "Połącz",
	"Withdraw":                    "Wypłać",
	"Deposit":                     "Zdeponuj",
	"Lock":                        "Zablokuj",
	"New Deposit Address":         "Nowy adres do depozytów",
	"Address":                     "Adres",
	"Amount":                      "Ilość",
	"Reconfigure":                 "Skonfiguruj ponownie",
	"pw_change_instructions":      "Zmiana poniższego hasła nie powoduje zmiany hasła do Twojego oprogramowania portfela. Skorzystaj z tego formularza, aby zaktualizować klienta DEX po tym, jak zmienisz hasło do swojego oprogramowania portfela.",
	"New Wallet Password":         "Nowe hasło portfela",
	"pw_change_warn":              "Uwaga: Zmiana portfela podczas gdy trwają wymiany może spowodować utratę środków.",
	"Show more options":           "Wyświetl więcej opcji",
	"seed_implore_msg":            "Zapisz ziarno aplikacji dokładnie na kartce papieru i zachowaj jego kopię. Jeśli stracisz dostęp do tego urządzenia lub niezbędnych plików aplikacji, ziarno to umożliwi Ci przywrócenie kont DEX oraz wbudowanych portfeli. Niektóre starsze konta nie mogę być przywrócone ta metodą, i niezależenie od tego, czy konto jest nowe, czy stare, zachowanie kopii swoich kluczy w dodatku do ziarna jest zawsze dobrą praktyką.",
	"View Application Seed":       "Wyświetl ziarno aplikacji",
	"Remember my password":        "Zapamiętaj hasło",
	"pw_for_seed":                 "Enter your app password to show your seed. Make sure nobody else can see your screen.",
	"Asset":                       "Aktywo",
	"Balance":                     "Saldo",
	"Actions":                     "Czynności",
	"Restoration Seed":            "Ziarno do przywrócenia",
	"Restore from seed":           "Przywróć z ziarna",
	"Import Account":              "Importuj konto",
	"no_wallet":                   "brak portfela",
	"create_a_x_wallet":           "Utwórz portfel <span data-asset-name=1></span>",
	"dont_share":                  "Nie udostępniaj nikomu. Nie zgub go.",
	"Show Me":                     "Pokaż",
	"Wallet Settings":             "Ustawienia portfela",
	"add_a_x_wallet":              `Dodaj portfel <img data-tmpl="assetLogo" class="asset-logo mx-1"> <span data-tmpl="assetName"></span>`,
	"ready":                       "gotowy",
	"off":                         "wyłączony",
	"Export Trades":               "Eksportuj zlecenia wymiany",
	"change the wallet type":      "zmień typ portfela",
	"confirmations":               "potwierdzenia",
	"how_reg":                     "Jak chcesz uiścić opłatę rejestracyjną?",
	"All markets at":              "Wszystkie rynki na",
	"pick a different asset":      "wybierz inne aktywo",
	"Create":                      "Utwórz",
	"Register_loudly":             "Zarejestruj!",
	"1 Sync the Blockchain":       "1: Zsynchronizuj blockchain",
	"Progress":                    "Postęp",
	"remaining":                   "pozostało",
	"2 Fund the Registration Fee": "2: Wnieś opłatę rejestracyjną",
	"Registration fee":            "Opłata rejestracyjna",
	"Your Deposit Address":        "Twój adres do wpłaty",
	"Send enough for reg fee":     `Upewnij się, że wysyłasz wystarczająco dużo środków, aby pokryć również opłaty sieciowe.`,
	"add a different server":      "dodaj inny serwer",
	"Add a custom server":         "Dodaj niestandardowy serwer",
	"plus tx fees":                "+ opłaty transakcyjne",
	"Export Seed":                 "Eksportuj ziarno",
	"Total":                       "W sumie",
	"Trading":                     "Wymiana",
	"Receiving Approximately":     "Otrzymując około",
	"Fee Projection":              "Szacunkowa opłata",
	"details":                     "szczegóły",
	"to":                          "do",
	"Options":                     "Opcje",
	"fee_projection_tooltip":      "Jeśli warunki sieciowe nie ulegną zmianie zanim Twoje zlecenie zostanie spasowane, całkowite poniesione opłaty (jako procent wymienianej sumy) powinny zmieścić się w tym przedziale.",
	"unlock_for_details":          "Odblokuj portfele, aby wyciągnąć dane zleceń oraz dodatkowe opcje.",
	"estimate_unavailable":        "Szacunkowe dane zleceń i opcje są niedostępne",
	"Fee Details":                 "Szczegóły opłat",
	"estimate_market_conditions":  "Dane szacunkowe dla najlepszego i najgorszego scenariusza oparte są na obecnych warunkach sieciowych i mogą ulec zmianie do czasu wykonania zlecenia.",
	"Best Case Fees":              "Opłaty (najlepszy scenariusz)",
	"best_case_conditions":        "Najlepszy scenariusz dla opłat występuje wtedy, gdy zlecenie jest zrealizowane w całości jednym spasowaniem.",
	"Swap":                        "Zamiana",
	"Redeem":                      "Wykupienie",
	"Worst Case Fees":             "Opłaty (najgorszy scenariusz)",
	"worst_case_conditions":       "Najgorszy scenariusz dla opłat może wystąpić wtedy, gdy zlecenie jest spasowane po jednym locie na przestrzeni kilku epok.",
	"Maximum Possible Swap Fees":  "Maksymalne możliwe opłaty wymiany",
	"max_fee_conditions":          "To absolutne maksimum tego, co możesz zapłacić przy swojej wymianie. Opłaty są zazwyczaj wyceniane na ułamek tej kwoty. Maksimum nie ulega zmianie po złożeniu zlecenia.",
}
