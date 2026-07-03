# Page Map - MIC Monolith

Questa mappa serve a un nuovo collega per capire rapidamente quali schermate
esistono, a cosa servono e quali aree di dominio toccano. Non e' una descrizione
esaustiva di tutti i campi: per i dettagli aprire il file JS indicato.

## Forma della GUI

MIC e' una SPA con routing hash-based. La sidebar in `php-app/public/index.html`
espone le rotte `#/...`; ogni vista e' registrata dai file in
`php-app/public/js/`.

| Contesto | Schermate | Perche' conta |
|---|---|---|
| Overview | Dashboard | Mostra l'andamento del business e attraversa piu' domini. |
| Catalogo e pricing | Articoli, Listini, Sconti, Categorie, Aliquote IVA | Definisce cosa si vende, a che prezzo e con quale IVA. |
| Anagrafiche | Clienti, Fornitori, Agenti | Soggetti collegati a ordini, fatture e vendite. |
| Vendite e fatturazione | Ordini, Fatture | Flusso operativo principale: ordine -> righe -> fattura -> pagamento. |
| Magazzino | Magazzino | Movimenti e giacenze per articolo. |
| Sistema | Utenti | Account applicativi e riferimenti per audit. |

## Schermate

### Dashboard (`#/dashboard`)

Vista iniziale con KPI, grafici e ultime attivita'. Serve a capire lo stato
generale: fatturato mese, ordini in corso, pagamenti scaduti e proxy articoli
sotto scorta.

Da notare: e' una schermata trasversale. Legge fatture, ordini, clienti,
articoli e audit log.

File: `php-app/public/js/dashboard.js`. API: `GET /api/dashboard/kpi`.

### Articoli (`#/articles`)

Catalogo dei prodotti/servizi vendibili. Permette di cercare, creare,
modificare ed eliminare articoli.

Da notare: l'articolo e' usato anche da listini, righe ordine e movimenti di
magazzino. E' uno dei nodi piu' importanti per una futura estrazione Warehouse.

File: `php-app/public/js/articles.js`. API: `/api/articles`.

### Listini (`#/listini`)

Gestione dei listini e delle voci listino. Dal dettaglio listino si aggiunge un
articolo con un prezzo.

Da notare: ogni voce listino collega pricing e articolo. Il pricing non e'
isolato dal catalogo.

File: `php-app/public/js/listini.js`. API: `/api/listini`,
`/api/listini/:id/voci`.

### Sconti (`#/sconti`)

Gestione dei codici sconto. Serve al pricing commerciale, anche se nella GUI e'
una lista CRUD semplice.

File: `php-app/public/js/sconti.js`. API: `/api/sconti`.

### Categorie (`#/categories`)

Configurazione delle categorie articolo. Serve soprattutto alla schermata
Articoli.

File: `php-app/public/js/settings.js`. API: `/api/categories`.

### Aliquote IVA (`#/iva`)

Configurazione delle aliquote IVA. Serve ad articoli e righe ordine, quindi sta
tra catalogo, vendite e fiscale.

File: `php-app/public/js/settings.js`. API: `/api/iva-rates`.

### Clienti (`#/customers`)

Anagrafica clienti. Oltre al CRUD, il dettaglio cliente mostra tab interne per
anagrafica, ordini, fatture e riepilogo pagamenti.

Da notare: cliente e' letto da ordini, fatture e dashboard. Non e' solo
un'anagrafica isolata.

File: `php-app/public/js/customers.js`. API: `/api/customers`,
`/api/customers/:id/orders`, `/api/customers/:id/invoices`.

### Fornitori (`#/suppliers`)

Anagrafica fornitori. Nel sistema corrente e' una lista CRUD semplice e meno
collegata ai flussi principali.

File: `php-app/public/js/suppliers.js`. API: `/api/suppliers`.

### Agenti (`#/agenti`)

Anagrafica agenti commerciali. Gli agenti possono essere collegati agli ordini.

File: `php-app/public/js/agenti.js`. API: `/api/agenti`.

### Ordini (`#/orders`)

Schermata centrale del ciclo vendite. Permette di creare ordini, collegare
cliente/listino/agente, aggiungere righe e generare una fattura.

Da notare: ogni riga ordine legge articolo e IVA, calcola imponibile/IVA/totale
e poi ricalcola la testata ordine. Qui si vede bene l'accoppiamento tra Sales,
Catalogo, Pricing e Fiscale.

File: `php-app/public/js/orders.js`. API: `/api/orders`,
`/api/orders/:id/righe`, `/api/invoices`.

### Fatture (`#/invoices`)

Gestione fatture. Permette creazione/modifica, invio SDI mock, registrazione
pagamento ed eliminazione.

Da notare: fattura collega cliente e ordine; il pagamento viene creato da questa
schermata e poi la fattura viene marcata `pagato`.

File: `php-app/public/js/invoices.js`. API: `/api/invoices`,
`/api/invoices/:id/invia-sdi`, `/api/pagamenti`.

### Magazzino (`#/magazzino`)

Vista operativa per magazzini, movimenti recenti e giacenze. Permette di
registrare movimenti su articolo e magazzino, e di aprire una modale con le
giacenze.

Da notare: la giacenza non e' una tabella, ma una somma dei movimenti collegati
a un articolo in un magazzino.

File: `php-app/public/js/magazzino.js`. API: `/api/magazzini`,
`/api/magazzini/:id/giacenze`, `/api/movimenti`.

### Utenti (`#/users`)

Gestione utenti applicativi. Nel sistema corrente e' una lista CRUD semplice,
ma gli utenti sono usati come riferimento nei log di audit.

File: `php-app/public/js/settings.js`. API: `/api/users`.

## API senza schermata dedicata

Alcune funzionalita esistono nel backend ma non hanno una pagina autonoma nella
sidebar:

| Area | Come emerge nella GUI | Perche' conta |
|---|---|---|
| Pagamenti | Azione "Registra pagamento" in Fatture | Collega incassi e fatturazione. |
| Note credito | Controller/API presenti, non esposte in sidebar | Fiscale incompleto lato GUI. |
| Audit log | Ultime attivita in Dashboard | Read model di sistema, non gestione operativa. |

