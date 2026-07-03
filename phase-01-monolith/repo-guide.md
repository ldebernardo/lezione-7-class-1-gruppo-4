# Repo Guide - MIC Monolith

Questa guida dice dove guardare nella cartella `phase-01-monolith/`. Non e' un
inventario completo: e' una mappa per un collega che entra domani.

## Primo orientamento

| Percorso | A cosa serve | Quando aprirlo |
|---|---|---|
| `README-IT.md` | Istruzioni del checkpoint. | Per capire missione e deliverable. |
| `docker-compose.yml` | Stack locale. | Per porte, container e variabili DB. |
| `Dockerfile` | Runtime della app. | Per capire PHP-FPM, Nginx e dipendenze. |
| `nginx.conf` | Routing HTTP. | Per seguire static assets, SPA e API. |
| `openapi.yaml` | Contratto API legacy, parziale. | Per endpoint documentati, soprattutto Warehouse. |
| `database/` | Schema e seed. | Per capire come sono salvati davvero i dati. |
| `php-app/` | Codice PHP e frontend. | Per seguire GUI, API e query. |
| `page-map.md` | Mappa schermate. | Per capire cosa vede l'utente. |
| `architecture.md` | Mappa runtime. | Per capire browser -> app -> DB. |
| `coupling-map.md` | Accoppiamenti peggiori. | Per capire cosa ostacola l'estrazione. |
| `worst-antipatterns.md` | Top 3 antipattern. | Per capire i rischi architetturali principali. |

## Dove vivono le cose

### Database

`database/schema.sql` crea due tabelle importanti:

- `business_data`: tutte le entita di business;
- `business_relations`: collegamenti tra entita.

`database/seed.sql` carica dati dimostrativi. E' utile per vedere quali
`record_type` e `relation_type` esistono davvero.

### Backend PHP

Punti di ingresso:

- `php-app/index.php`: front controller e registrazione rotte API;
- `php-app/src/Router.php`: router minimale;
- `php-app/src/Database.php`: connessione PDO a MySQL;
- `php-app/src/Repository.php`: lettura/scrittura generica su database;
- `php-app/src/Controllers/`: controller per le aree funzionali.

I controller sembrano separare i domini, ma quasi tutti passano dal repository
generico. Per capire una regola, leggere il controller e subito dopo la query o
la relazione usata.

### Frontend SPA

Punti di ingresso:

- `php-app/public/index.html`: shell, sidebar e script caricati;
- `php-app/public/js/app.js`: router hash, fetch helper, modali, tabelle;
- `php-app/public/js/*.js`: una vista per area funzionale;
- `php-app/public/css/style.css`: stile della GUI.

Per capire una schermata, usare questa sequenza:

1. `page-map.md`;
2. file JS della vista;
3. controller PHP chiamato dalla vista;
4. `Repository.php` e schema dati.

## Rotte mentali utili

| Se cerchi... | Parti da... |
|---|---|
| Lista schermate | `page-map.md`, poi `php-app/public/index.html`. |
| Flusso ordine -> fattura | `orders.js`, `OrderController`, `InvoiceController`. |
| Magazzino e giacenze | `magazzino.js`, `MagazzinoController`, `MovimentoController`. |
| Significato reale dei dati | `database/schema.sql`, `database/seed.sql`. |
| Accoppiamenti da spezzare | `coupling-map.md`. |
| Perche' il design e' rischioso | `worst-antipatterns.md`. |

## Forma del codice

Il pattern dominante e':

```text
Vista JS -> /api/... -> index.php -> Router -> Controller -> Repository -> MySQL
```

La cosa da ricordare: il nome di una schermata o di un controller suggerisce un
dominio, ma il database non ha tabelle di dominio. Il dominio viene ricostruito
nel codice interpretando campi generici.

