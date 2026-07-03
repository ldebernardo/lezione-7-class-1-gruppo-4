# Warehouse Article Work Plan

Piano di lavoro per sviluppare in Go la prima capability del Warehouse BC:
l'aggregate `Article` con il suo inventario, le invarianti minime e i domain
event registrati ma non emessi.

Questo documento e' solo un piano: non contiene codice.

## Obiettivo

Costruire il primo nucleo tattico del Warehouse BC attorno a `Article`.

L'aggregate iniziale deve:

- rappresentare un articolo stockable con `SKU`, `name`, `description`, `price`
  e zero-o-piu' livelli di inventario;
- possedere l'inventario come parte dell'aggregate, non come oggetto caricato o
  salvato separatamente;
- far rispettare almeno queste invarianti:
  - `price > 0`;
  - `name` non vuoto;
  - `price.currency` non cambia dopo la creazione;
  - quantita e reserved non negative;
  - `reserved <= quantity`;
- registrare domain event dentro l'aggregate;
- non pubblicare eventi dal dominio.

## Regole non negoziabili per l'aggregate

Queste regole sono il gate della prima tranche di lavoro. Se una scelta tecnica
le viola, va respinta anche se rende il codice piu' comodo.

| Regola | Decisione pratica |
|---|---|
| 1. `Article` guards its invariants | Solo `Article` puo' creare o mutare stato che coinvolge price, name, currency o inventory. `price > 0`, `name` non vuoto e currency immutabile dopo creazione sono controlli del dominio, non di HTTP o database. |
| 2. Value objects have no identity | `SKU` e `Money` non hanno ID. Si confrontano per valore e nascono da factory/funzioni di costruzione che rifiutano input non validi. |
| 3. Stock stays sane | `Reserved` non supera mai `Quantity`; `Quantity` non diventa mai negativa; l'inventario e' raggiungibile solo attraverso `Article`. |
| 4. Domain layer only | La prima implementazione contiene solo tipi business e regole. Niente database, niente HTTP, niente ACL, niente repository concreti. |
| 5. Record events, do not publish | L'aggregate appende eventi pending. Un layer futuro li drenara'. Nessun broker, publisher o messaging nel dominio. |

## Fonti e decisioni gia' prese

| Fonte | Decisione usata dal piano |
|---|---|
| `phase-01-monolith/architecture.md` | Oggi MIC e' una SPA + monolite PHP + repository generico + MySQL. |
| `phase-01-monolith/coupling-map.md` | `Article` e' nodo condiviso da ordini, listini, magazzino e dashboard. |
| `phase-01-monolith/worst-antipatterns.md` | Evitare database generico, logica sparsa e relazioni cross-domain implicite. |
| `event-storming-canvas.md` | Warehouse nasce dal Cluster C; per l'esercizio si usa un solo aggregate `Article`. |
| `ubiquitous-language.md` | I nomi canonici sono `Article`, `SKU`, `Money`, `Price`, `Inventory Level`, `Quantity`, `Warehouse Location`, `Stock Reservation`. |
| `context-mapping.md` | Orders -> Warehouse e' Customer/Supplier; Go Warehouse <-> schema legacy passa da ACL. |
| `tactical-ddd-warehouse.md` | `Article` e' aggregate root; `InventoryLevel` e' entity interna; `SKU` e `Money` sono value object; gli eventi sono registrati, non pubblicati. |
| `DEPENDENCY-MAP.md` | Warehouse e' safe-to-extract first: 3 dipendenze in ingresso, 0 in uscita. |
| ADR-001 | Estrarre con Strangler Fig, non con riscrittura big-bang. |
| ADR-010 | Event Storming in ordine: enumerate -> cluster -> name. |
| ADR-011 | Il codice Go deve usare Aggregate, Entity, Value Object e Domain Event come vocabolario strutturale. |

## Part 1 - Strategico

### Step 1 - Fissare il confine della capability Warehouse

Descrivere la capability come:

> Warehouse gestisce articoli stockable, livelli di inventario per magazzino e
> riserva logica di stock. Non possiede ordini, fatture, clienti, IVA, categorie
> o listini.

Decisioni operative:

- `Article` nel Warehouse non e' il Catalog Article completo.
- Category, VAT rate, listino e customer-specific pricing restano fuori.
- `StockMovement` ledger resta fuori dal primo Go BC.
- `StockReservation` resta semplificata come contatore `Reserved` dentro
  `InventoryLevel`.

Output atteso:

- nota nel piano tecnico o issue epic che dichiara cosa e' dentro/fuori
  Warehouse.

### Step 2 - Confermare la Ubiquitous Language

Prima di nominare package, struct o metodi Go, confermare i termini UL da usare:

- `Article`;
- `InventoryLevel`;
- `SKU`;
- `Money`;
- `Price`;
- `Quantity`;
- `WarehouseLocation`;
- `StockReservation`.

Regole:

- niente sinonimi come `Product`, `Item`, `StockItem`, `WarehouseItem`;
- niente termini di altri BC dentro il modello Warehouse, eccetto identificatori
  opachi quando servono, per esempio `orderId` in eventi di reservation futuri;
- se serve un nuovo concetto, aggiornare prima `ubiquitous-language.md`.

Output atteso:

- checklist UL allegata alla issue di implementazione;
- Prompt AI 1 eseguito su `event-storming-canvas.md` e confronto con
  `ubiquitous-language.md`.

### Step 3 - Validare il context mapping prima del design API

Usare queste relazioni come vincoli:

- Orders consuma Warehouse via Customer/Supplier.
- Go Warehouse protegge il proprio modello dal legacy schema con ACL.
- Pricing/Catalog restano fuori dal modello interno Warehouse.

Conseguenze per il piano:

- il dominio Go non deve contenere campi `amount_1`, `text_1`, `record_type`;
- il dominio Go non deve dipendere dal database MIC;
- eventuali adapter legacy traducono da/verso lo schema PHP, ma non contaminano
  aggregate e value object.

Output atteso:

- decisione esplicita: il package domain non importa adapter, HTTP o storage.

## Part 2 - Tattico + piano Go

### Step 4 - Disegnare lo scope del domain layer Go

La prima tranche deve fermarsi al dominio. Struttura orientativa, senza ancora
scrivere codice:

```text
warehouse/
  domain/
    article/
      Article aggregate root
      InventoryLevel entity
      SKU value object
      Money value object
      WarehouseLocation value object o LocationCode iniziale
      DomainEvent interface/concept
      ArticleCreated event
      ArticlePriceChanged event
      InventoryAdjusted event
```

Vincoli:

- non creare repository per `SKU`, `Money` o `InventoryLevel`;
- `InventoryLevel` e' mutabile solo attraverso `Article`.
- il domain layer non importa database, HTTP, adapter, framework o messaging;
- application layer, repository port, ACL legacy e HTTP handler sono esplicitamente
  fuori scope per questa tranche.

Output atteso:

- lista dei package e responsabilita;
- conferma che ogni nome e' classificabile come Aggregate, Entity, Value Object,
  Domain Event o fuori-BC.

### Step 5 - Modellare `Article` come aggregate root

Responsabilita di `Article`:

- creare un nuovo articolo valido;
- cambiare il nome solo se non vuoto;
- cambiare il prezzo solo se `currency` resta identica;
- aggiungere o aggiornare un `InventoryLevel`;
- aggiustare quantita e reserved rispettando le invarianti;
- registrare eventi di dominio in una lista interna pending.

Invarianti minime:

| Invariante | Dove va protetta |
|---|---|
| `name` non vuoto | factory/constructor e rename |
| `price > 0` | factory/constructor e change price |
| currency immutabile dopo creazione | change price |
| `quantity >= 0` | inventory adjustment |
| `reserved >= 0` | reservation/release o contatore iniziale |
| `reserved <= quantity` | ogni mutazione inventario |

Controlli da non accettare:

- non spostare queste regole in handler HTTP;
- non affidarle a vincoli database;
- non permettere mutazioni dirette su `InventoryLevel` dall'esterno
  dell'aggregate;
- non aggiungere setter pubblici che possano saltare le invarianti.

Output atteso:

- specifica testabile dei metodi dell'aggregate;
- nessuna regola affidata solo a handler HTTP o database.

### Step 6 - Definire i Value Object

Definire i value object come tipi validi per costruzione:

- `SKU`: codice alfanumerico globale, validato con la regex prevista;
- `Money`: amount in cents + currency ISO 4217;
- `WarehouseLocation` o, nella prima versione, `LocationCode` esplicito;
- `Quantity`, se il team decide di non lasciarla come intero primitivo.

Regole:

- un value object non ha identita;
- un value object non viene mutato, viene sostituito;
- la validazione fallisce subito;
- `Money` puo' rappresentare zero, ma `Article` non accetta price zero.
- `SKU` e `Money` si confrontano per valore: stessi dati, stesso value object;
- nessun repository, ID o lifecycle indipendente per i value object.

Output atteso:

- tabella Value Object -> validazioni -> errori di dominio attesi.

### Step 7 - Definire `InventoryLevel` come entity interna

`InventoryLevel` rappresenta la giacenza di un articolo in una location.

Deve avere:

- identita propria;
- location;
- quantity;
- reserved.

Non deve avere:

- repository proprio;
- API pubblica indipendente;
- mutazioni dirette da application layer.

Tutte le modifiche passano da `Article`.

Regole di sanity:

- se un adjust porterebbe `quantity < 0`, il dominio rifiuta l'operazione;
- se una reserve porterebbe `reserved > quantity`, il dominio rifiuta
  l'operazione;
- se una release porterebbe `reserved < 0`, il dominio rifiuta l'operazione.

Output atteso:

- descrizione dei comportamenti `adjust`, `reserve`, `release` come operazioni
  esposte dall'aggregate root, non dall'entity verso l'esterno.

### Step 8 - Registrare domain events senza emetterli

Gli eventi iniziali da pianificare sono:

| Evento | Quando viene registrato | Nota |
|---|---|---|
| `ArticleCreated` | creazione valida di un Article | Include id, sku, name, price, currency, occurredAt. |
| `ArticlePriceChanged` | cambio prezzo con stessa currency | Include old/new price e currency. |
| `InventoryAdjusted` | variazione quantity di un InventoryLevel | Include articleId, location, delta, newQuantity, reason, occurredAt. |

Regola fondamentale:

- l'aggregate registra eventi in `pendingEvents`;
- l'aggregate non pubblica eventi;
- un layer futuro potra' leggerli/drainarli dopo persistenza riuscita;
- nessun event bus, broker o side effect nella prima implementazione.

Output atteso:

- contratto testabile: dopo un comando valido, l'aggregate contiene gli eventi
  attesi; dopo drain, la lista pending e' vuota.

### Step 9 - Esplicitare cosa resta fuori scope ora

Per rispettare la regola "Domain layer only", questi elementi non vanno
implementati nella prima tranche:

- handler HTTP;
- controller REST;
- database o migrazioni;
- repository concreti;
- adapter ACL verso `business_data` / `business_relations`;
- event publisher, broker, queue o messaging;
- use case applicativi che orchestrano salvataggio e pubblicazione.

Saranno pianificati dopo che il dominio sara' stabile. In una phase successiva,
gli use case potranno:

- costruire value object;
- caricare aggregate quando necessario;
- invocare metodi dell'aggregate;
- salvare aggregate tramite repository port;
- drenare eventi pending dopo save riuscito.

Output atteso:

- lista di esclusioni accettata in review;
- nessuna dipendenza infrastrutturale nel domain layer.

### Step 10 - Pianificare il repository port come lavoro successivo

Il repository e' fuori scope dalla prima tranche, ma va gia' vincolato per non
rompere l'aggregate quando verra' introdotto.

Quando sara' introdotto, dovra':

- salvare `Article` aggregate;
- trovare `Article` per id;
- trovare `Article` per SKU, se serve al controllo unicita;
- eventuale optimistic locking in fase successiva.

Regola:

- il repository salva l'aggregate intero;
- non salva `InventoryLevel` separatamente;
- non espone query orientate a `amount_1` o `record_type`.

Output atteso:

- nota di design futura sul repository port;
- nota che l'ACL legacy sara' un adapter, non parte del dominio.

### Step 11 - Pianificare i test

Prima dei dettagli infrastrutturali, pianificare test di dominio.

Test obbligatori:

- non crea `Article` con `name` vuoto;
- non crea `Article` con `price <= 0`;
- non crea `Article` con `Money` costruito da input invalido;
- non crea `Article` con `SKU` costruito da input invalido;
- confronta `SKU` e `Money` per valore, non per identita;
- non cambia price se cambia currency;
- cambia price se currency resta uguale e amount > 0;
- registra `ArticleCreated` alla creazione;
- registra `ArticlePriceChanged` al cambio prezzo;
- crea o aggiorna `InventoryLevel` solo attraverso `Article`;
- non permette quantity negativa;
- non permette reserved negativo;
- non permette `reserved > quantity`;
- registra `InventoryAdjusted` dopo aggiustamento valido;
- gli eventi restano pending finche' non vengono drenati.

Output atteso:

- lista test Given/When/Then;
- nessun test che richieda MySQL, Docker o PHP legacy per validare il dominio.
- nessun test che richieda HTTP handler o publisher.

### Step 12 - Preparare la migrazione Strangler Fig

Non migrare subito i consumatori. Preparare il percorso:

1. implementare dominio Go pulito;
2. solo dopo aggiungere application layer e repository port;
3. solo dopo aggiungere adapter ACL verso dati legacy;
4. eseguire dual-read/dual-write quando previsto dalle phase successive;
5. confrontare comportamento Go e PHP;
6. spostare gradualmente i consumatori;
7. mantenere fallback al monolite finche' la confidence non e' sufficiente.

Output atteso:

- runbook di transizione;
- nota esplicita che Orders, Pricing e Dashboard non devono leggere storage
  interno Warehouse.

## Sequenza suggerita di lavoro

1. Aprire Event Storming e confermare che `Article` e inventory appartengono al
   Warehouse BC.
2. Eseguire Prompt AI 1 della UL e confrontare il risultato con il vocabolario
   corrente.
3. Confermare context mapping: Orders Customer/Supplier, legacy schema via ACL.
4. Scrivere una scheda tattica per `Article` aggregate, `InventoryLevel`, `SKU`,
   `Money` e domain events.
5. Definire gli invarianti e trasformarli in test Given/When/Then.
6. Implementare in Go i value object.
7. Implementare in Go `InventoryLevel` come entity interna.
8. Implementare in Go `Article` aggregate root.
9. Aggiungere registrazione pending events senza publisher.
10. Fermarsi: niente HTTP, database, repository concreto o ACL in questa tranche.
11. Solo dopo review del dominio, progettare application layer e repository port.
12. Solo dopo, progettare adapter ACL e integrazione legacy.

## Verifica rispetto al README-IT

Questa tabella verifica che il piano copra tutti i punti della sezione
`Verifica` del README Phase 02.

| Punto README | Copertura nel piano | Evidenza attesa |
|---|---|---|
| Sai condurre un Event Storming su una narrazione di dominio di 1 pagina senza guardare i passi | Step 1 e sequenza 1 richiamano enumerate -> cluster -> name e il Cluster C Warehouse. | Il team sa spiegare perche' Warehouse deriva dagli eventi, non dallo schema. |
| Sai scrivere una riga di UL per un nuovo termine di dominio | Step 2 impone aggiornamento UL prima di introdurre nuovi concetti. | Nuova riga nel formato canonico EN, IT, definizione. |
| Hai eseguito almeno Prompt AI 1 con `event-storming-canvas.md` | Step 2 e sequenza 2 lo rendono gate obbligatorio. | Output del prompt confrontato con `ubiquitous-language.md`. |
| Sai nominare la famiglia per ogni integrazione cross-BC | Step 3 usa Customer/Supplier e ACL, e riconosce Conformist indiretto. | Ogni integrazione ha famiglia Cooperation / Upstream-Downstream / Separate Ways. |
| Sai classificare un nuovo nome Warehouse come Aggregate / Entity / Value Object / Domain Event / fuori-BC | Step 4-8 classificano `Article`, `InventoryLevel`, `SKU`, `Money`, eventi e concetti fuori scope. | Ogni nuovo nome nella issue e' classificato prima del codice. |
| Hai risposto a tutte le 7 domande della participant checklist | La sequenza di lavoro richiede di usare i documenti nell'ordine del README prima del codice. | Risposte D1-D7 raccolte nel materiale di review del team. |
| Hai letto ADR-001, ADR-010 e ADR-011 | Le decisioni degli ADR sono integrate nella tabella fonti e negli step 1, 8, 12. | Review conferma: Strangler Fig, Event Storming e building block DDD sono rispettati. |

## Definition of Ready per iniziare il codice Go

Il codice Go puo' iniziare solo quando:

- la UL Warehouse e' confermata;
- `Article` e' accettato come aggregate root;
- `InventoryLevel` e' accettato come entity interna;
- `SKU` e `Money` sono accettati come value object;
- gli eventi iniziali sono nominati al passato;
- e' chiaro che gli eventi vengono registrati, non emessi;
- gli invarianti sono trasformati in test;
- la prima tranche e' limitata al domain layer;
- non esiste dipendenza da database, HTTP, ACL o messaging;
- l'ACL legacy e' pianificato come adapter esterno al dominio.

## Acceptance checklist delle cinque regole aggregate

- [ ] `Article` protegge direttamente `price > 0`, `name` non vuoto e currency
  immutabile dopo creazione.
- [ ] `SKU` e `Money` sono value object senza identita, confrontabili per valore
  e costruiti solo tramite factory/funzioni che rifiutano input invalido.
- [ ] Lo stock resta consistente: `Reserved <= Quantity`, `Quantity >= 0`,
  `Reserved >= 0`, e ogni mutazione passa da `Article`.
- [ ] Il lavoro iniziale e' solo domain layer: niente database, HTTP, adapter,
  repository concreti o framework.
- [ ] Gli eventi sono solo registrati pending dall'aggregate; nessuna
  pubblicazione o messaggistica nel dominio.
