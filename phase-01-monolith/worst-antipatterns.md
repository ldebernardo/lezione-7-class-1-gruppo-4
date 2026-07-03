# Worst Antipatterns - MIC Monolith

Questa lista nomina le 3 scelte che rendono MIC piu' difficile da cambiare.
Sono le cose da tenere davanti agli occhi quando si pianifica lo Strangler Fig.

## 1. Database generico al posto del modello di dominio

### Cos'e'

Il database usa due tabelle generiche:

- `business_data` per tutte le entita;
- `business_relations` per tutti i legami.

Il tipo reale e' deciso da stringhe come `record_type='ordine'` o
`relation_type='cliente_di_ordine'`. I campi `amount_1`, `text_1`, `date_1` non
hanno un significato unico.

### Perche' fa male

Il dominio non e' espresso nello schema. Non ci sono tabelle, vincoli e relazioni
specifiche per clienti, ordini, fatture, movimenti o articoli. Il significato
dei dati vive nei controller e nella memoria del team.

Questo aumenta il rischio di bug e rende ogni modifica piu' lenta: prima di
cambiare un campo bisogna capire cosa significa per ogni `record_type`.

### Costo di conviverci

Molto alto. Qualunque estrazione dovra' prima tradurre righe generiche in modelli
di dominio veri. Non si puo' "spostare la tabella Warehouse", perche' la tabella
contiene anche Sales, Billing, Pricing e Anagrafiche.

## 2. Regole di business sparse tra controller e frontend

### Cos'e'

Le regole non hanno un livello applicativo o di dominio dedicato. Vivono nei
controller PHP e, in alcuni casi, nei file JS.

Esempi utili:

- l'ordine aggiunge righe e ricalcola totali nel controller;
- la fattura viene inviata allo SDI mock dal controller;
- la GUI Fatture registra un pagamento e poi aggiorna lo stato fattura;
- la dashboard fa query aggregate cross-domain.

### Perche' fa male

Non c'e' un posto ovvio dove proteggere invarianti e workflow. Un nuovo endpoint
potrebbe aggirare una regola; un cambio nella GUI potrebbe spostare pezzi di
processo fuori dal backend.

Il sistema funziona come demo, ma non offre confini chiari per testare o
riusare i casi d'uso.

### Costo di conviverci

Alto. Prima di estrarre un bounded context bisognera' separare i casi d'uso dal
trasporto HTTP e dal formato legacy del database. Altrimenti il nuovo servizio
eredita lo stesso disordine, solo in un altro runtime.

## 3. Accoppiamenti cross-domain nascosti in stringhe

### Cos'e'

I legami tra domini sono stringhe in `business_relations`:

- `articolo_in_riga_ordine`;
- `iva_di_riga`;
- `fattura_di_ordine`;
- `pagamento_di_fattura`;
- `movimento_di_articolo`;
- `voce_di_listino`.

Queste stringhe sono il vero grafo del dominio, ma non sono tipizzate e non
proteggono ownership o invarianti.

### Perche' fa male

I controller sembrano separati, ma attraversano lo stesso database condiviso.
Sales legge Catalogo e Fiscale; Pricing legge Catalogo; Warehouse legge
Articoli; Fatturazione legge Ordini e Clienti; Dashboard legge quasi tutto.

Quando si estrae un pezzo, emergono dipendenze che non erano visibili guardando
solo la sidebar o i nomi dei controller.

### Costo di conviverci

Alto per la modernizzazione. Ogni nuova relazione aggiunge un'altra stringa e
un'altra query da ricordare. Il team dovra' fare archeologia a ogni refactor.

Per Warehouse, il nodo critico e' `articolo`: serve a magazzino, ordini,
listini, IVA e dashboard.

## Sintesi

| Rank | Antipattern | Effetto pratico |
|---:|---|---|
| 1 | Database generico | Il dominio non ha forma nei dati. |
| 2 | Regole sparse | Workflow e invarianti sono difficili da proteggere. |
| 3 | Relazioni via stringhe | Gli accoppiamenti reali sono nascosti. |

## Prima mossa consigliata

Non partire da una riscrittura completa. Prima rendere espliciti i confini:

1. isolare i casi d'uso piu' critici: ordine, riga ordine, movimento, giacenza;
2. introdurre DTO/read model che nascondano `business_data`;
3. decidere ownership dell'articolo;
4. solo dopo iniziare a strangolare Warehouse.

