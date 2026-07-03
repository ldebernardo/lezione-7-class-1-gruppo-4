# Riepilogo di `phase-03-skeleton`

Questa fase produce lo **scheletro del microservizio Go** che sostituirà progressivamente il monolite PHP, seguendo lo **Strangler Fig Pattern**. Il progetto implementa il **Domain Layer** dell'articolo di magazzino secondo i principi della **Clean Architecture** e del **DDD**.

---

## Struttura generale

```
phase-03-skeleton/
├── main.go                  ← entry-point HTTP (porta 8081)
├── main_test.go             ← test HTTP per /health
├── go.mod                   ← modulo Go 1.22, path: warehouse.local/core
├── Makefile                 ← comandi test/race/stress/up/down/health
├── Dockerfile               ← build multi-stage (build + run tests + serve)
├── docker-compose.yml       ← stack: warehouse + test runner + MySQL 8
├── nginx.conf               ← reverse proxy (coesistenza con PHP)
├── openapi.yaml             ← contratto API (placeholder)
├── design-decisions.md      ← ADR inline sulle scelte di design
├── entities/                ← DOMAIN MODEL
│   ├── article.go           ← Aggregate Root
│   ├── inventory_level.go   ← Entità interna (privata)
│   ├── money.go             ← Value Object
│   ├── sku.go               ← Value Object
│   ├── warehouse_location.go← Value Object
│   ├── errors.go            ← Errori sentinella del dominio
│   ├── article_test.go      ← Test dell'aggregato
│   └── value_objects_test.go← Test dei value object
├── events/                  ← DOMAIN EVENTS
│   ├── domain_event.go      ← Interfaccia base
│   ├── article_events.go    ← ArticleCreated, ArticlePriceChanged
│   └── inventory_events.go  ← InventoryAdjusted, StockReserved, StockReservationReleased
├── interfaces/              ← PORTE (Ports & Adapters)
│   └── article_repository.go← Interfaccia repository (solo contratto)
├── database/
│   ├── schema.sql           ← Schema del monolite originale (anti-pattern documentato)
│   └── seed.sql             ← Dati di seed
└── php-app/                 ← Monolite PHP originale (invariato, coesistenza Strangler Fig)
```

---

## I file del dominio e cosa fanno

### `entities/article.go` — Aggregate Root

`Article` è l'**unica porta di accesso** a tutte le mutazioni del dominio. L'inventario (`InventoryLevel`) non è mai esposto come oggetto mutabile: i metodi pubblici sono `AdjustInventory`, `ReserveStock`, `ReleaseStock`. Il campo `inventory` è una mappa privata e il costruttore con `newInventoryLevel` (minuscolo) è package-private — nessun codice esterno può creare livelli di inventario direttamente.

Il mutex `sync.Mutex` per campo garantisce la **thread-safety** dell'aggregato. Ogni metodo che muta lo stato appende un domain event alla lista `pending` (campo privato).

### `entities/errors.go` — Errori sentinella

Tutti i vincoli di business sono codificati come `var Err... = errors.New(...)`, permettendo ai test di usare `errors.Is()` per verificare il tipo preciso di violazione senza dipendere da stringhe.

### `entities/money.go`, `sku.go`, `warehouse_location.go` — Value Objects

Ogni value object è **immutabile** (nessun setter), costruito tramite factory con validazione inline via regex:

- `SKU`: regex `^[A-Z0-9-]{3,32}$` — solo maiuscole, cifre, trattino, 3–32 caratteri
- `Money`: regex `^[A-Z]{3}$` per la valuta + `amountCents >= 0`
- `WarehouseLocation`: regex `^[A-Z]{2}-[A-Z0-9]{3,8}$` — formato tipo `IT-MIL1`

Espongono un metodo `Equal()` che formalizza il confronto per valore (non per riferimento).

### `events/` — Domain Events

Tre file, un pattern unico: tutti i tipi embeddano `baseEvent` (che porta `occurredAt`) e implementano l'interfaccia `DomainEvent`. I campi sono privati, esposti solo tramite getter. Gli eventi sono **immutabili** dopo la creazione (nessun setter pubblico).

I 5 eventi definiti coprono esattamente le operazioni del dominio articolo:

- `ArticleCreated`, `ArticlePriceChanged`
- `InventoryAdjusted`, `StockReserved`, `StockReservationReleased`

### `interfaces/article_repository.go` — Porta Repository

Solo un'**interfaccia Go pura**: `Save`, `FindByID`, `FindBySKU`, `List`, `Delete`. Nessuna implementazione concreta in questa fase — rispetta il vincolo che il domain layer non importa driver di database o framework HTTP.

---

## I test e come dimostrano il rispetto dei constraint

### `entities/article_test.go`

| Test | Constraint verificato |
|---|---|
| `TestArticleFactoryGuardsNameAndPrice` | Rifiuta ID vuoto (`ErrInvalidArticleID`), nome blank (`ErrInvalidName`), prezzo zero (`ErrInvalidPrice`) |
| `TestArticleChangePriceKeepsCurrencyStable` | Rifiuta cambio di valuta (`ErrCurrencyChange`); verifica che l'evento `ArticlePriceChanged` venga registrato |
| `TestArticleRecordsEventsAndDrainsWithoutPublishing` | L'evento `ArticleCreated` è nell'elenco pending dopo la costruzione; `PullEvents()` svuota la lista; nessun publish esterno avviene (nessuna dependency su broker) |
| `TestInventoryStaysSaneAndOnlyArticleMutatesIt` | Quantità non può scendere sotto zero (`ErrInvalidQuantity`); non si può riservare più dello stock (`ErrReservedExceedsStock`); non si può rilasciare più del riservato (`ErrInvalidReserved`); `Available() = Quantity - Reserved` è sempre verificato |
| `TestInventoryStressKeepsStockSane` | Test di stress con goroutine concorrenti, verifica che il mutex garantisca la coerenza sotto carico (`make race` nel Makefile) |

### `entities/value_objects_test.go`

| Test | Constraint verificato |
|---|---|
| `TestSKUFactoryRejectsBadInputAndComparesByValue` | Rifiuta stringhe vuote, troppo corte, con minuscole, con underscore, troppo lunghe; `Equal()` funziona per valore |
| `TestMoneyFactoryRejectsBadInputAndComparesByValue` | Rifiuta centesimi negativi, valuta vuota, corta, lunga, minuscola; `0 EUR` è valido come value object (il vincolo `> 0` è sull'articolo, non sul tipo) |

### `main_test.go`

| Test | Constraint verificato |
|---|---|
| `TestHealthEndpoint` | GET `/health` → 200 con body `{"status":"ok"}` |
| `TestHealthEndpointRejectsUnsupportedMethod` | POST `/health` → 405 (solo GET consentito) |

---

## Constraint architetturali rispettati (da `design-decisions.md`)

1. **Article è l'unico Aggregate Root** — `InventoryLevel` è costruito con `newInventoryLevel` (privato) e mutato solo da metodi di `Article`. Nessun `InventoryRepository` indipendente.

2. **Domain packages senza import infrastrutturali** — `go.mod` non contiene alcun driver MySQL, framework HTTP o client broker. I package `entities`, `events`, `interfaces` importano solo la stdlib.

3. **Gli eventi sono registrati, non pubblicati** — Il `pending []events.DomainEvent` è drenato da `PullEvents()`, che restituisce gli eventi e azzera la lista. Un layer applicativo esterno (non ancora implementato) si occuperà della pubblicazione dopo la persistenza.

4. **Coesistenza con il monolite PHP** — La cartella `php-app/` e l'`nginx.conf` sono presenti: il reverse proxy può instradare le route `/articles/*` verso il nuovo servizio Go e tutto il resto verso PHP, abilitando lo Strangler Fig graduale.

5. **I test girano nel build Docker** — Il `Dockerfile` esegue `go test ./...` nello stage di build (`RUN go test ./...`), quindi un'immagine che non passa i test **non viene prodotta**.
