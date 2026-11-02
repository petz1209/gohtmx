# 07 – Rennergebnisse eintragen

**Branch:** `t07/rennergebnisse` → `t07/complete`
**Neu:** Updates einzelner Zeilen, gefilterte Events, Prepend neuer Zeilen, OOB-Swaps, serverseitiges Flash-Highlight, mehrere Fragmente in einer Antwort

## Worum es geht

Eine Seite pro Rennen (`competition_instance`), zum Beispiel „Leogang 2026“:

- oben ein **Podium** (Platz 1–3) und eine Statistik, z. B. „3 Starter, 2 gewertet, 1 DNF“,
- darunter die **Starterliste** mit Platzierung pro Fahrer,
- ein Bereich „Starter hinzufügen“ (Fahrer-Select, nur die, die noch nicht starten).

Jede Änderung aktualisiert **genau die betroffenen Teile**: die eine Zeile, das Podium und die Statistik.
Nicht die ganze Tabelle.

## Abnahme

- [ ] `/races/{id}` zeigt Titel (Wettbewerb, Ort, Datum, Status), Podium, Statistik und Starterliste.
- [ ] Platzierung einer Zeile ändern (Input + Enter oder Blur) → **nur diese Zeile** wird neu
      geladen und blinkt kurz grün.
- [ ] Gleichzeitig aktualisieren sich Podium und Statistik, **ohne** eigenen Request (OOB).
- [ ] Leeres Feld = DNF bzw. noch nicht gefahren (`NULL`).
- [ ] Platz doppelt vergeben → Fehler-Toast, Zeile zeigt den alten Wert. (Regel der App, nicht der DB.)
- [ ] Starter hinzufügen → neue Zeile erscheint **oben** in der Liste, blinkt blau, Statistik
      aktualisiert sich, der Fahrer verschwindet aus dem Select.
- [ ] Starter entfernen → Zeile verschwindet, Podium/Statistik aktualisieren sich.
- [ ] Rennen mit Status `planned` erlaubt keine Platzierungen, nur Starter. Rennen mit Status
      `cancelled` ist komplett schreibgeschützt.

> **Hinweis zu den Daten:** In `init.sql` sind fast alle Rennen `finished`, nur Leogang 2027 ist
> `planned`. Behandle `finished` in diesem Projekt deshalb noch wie `running` (Platzierungen
> erlaubt). Den Schreibschutz für `finished` und das Abschließen baust du in Projekt 10.

## Zusammenhänge

### Zwei Wege, eine Zeile zu aktualisieren

**A) Direkte Antwort:** Das Input sitzt **in** der Zeile. Request mit `hx-target="closest tr"
hx-swap="outerHTML"`. Die Antwort ist die neue Zeile. Einfach und direkt.

**B) Event + Zeilen-Endpoint:** Die Änderung passiert **woanders** (zum Beispiel in einem Modal).
Der Server antwortet mit `HX-Trigger: {"result-updated": {"id": 12}}`. Jede Zeile hört zu, aber
nur auf **ihre** id:

```html
<tr id="result-12"
    hx-get="/races/2/results/12/row?state=changed"
    hx-trigger="result-updated[detail.id==12] from:body"
    hx-swap="outerHTML">
```

Der Ausdruck in `[...]` ist ein **Event-Filter**: JavaScript, das gegen das Event ausgewertet wird.
`detail` ist der Payload aus dem Header.

Hier reicht A. Bau B trotzdem einmal für das Entfernen-Menü oder ein „Bearbeiten“-Modal ein,
damit du beide Wege kennst. Wann ist welcher richtig?

### Neue Zeile oben einfügen

```html
<tbody id="starters"
       hx-get="/races/2/results/row" hx-trigger="result-created from:body"
       hx-vals="js:{id: event.detail.id, state: 'new'}"
       hx-swap="afterbegin">
```

Oder direkt: Das „Hinzufügen“-Formular hat `hx-target="#starters" hx-swap="afterbegin"`, und der
Server liefert die neue Zeile. Denk wieder über A vs. B nach.

Die neue Zeile steht oben, **unabhängig von der Sortierung**. Das ist eine Produktentscheidung
(„ich will sehen, was ich gerade angelegt habe“), kein Zufall. Ein Reload sortiert sie richtig ein.

### Out-of-Band-Swaps

Die Antwort auf „Platz ändern“ enthält **mehr** als die Zeile:

```html
<tr id="result-12">…neue Zeile…</tr>

<div id="podium" hx-swap-oob="true">…neues Podium…</div>
<div id="race-stats" hx-swap-oob="true">…neue Statistik…</div>
```

htmx nimmt alle Elemente mit `hx-swap-oob` aus der Antwort heraus und ersetzt die Elemente mit
derselben `id` auf der Seite. Der Rest geht wie gewohnt ins Target.

Für den Renderer heißt das: Eine Antwort besteht aus **mehreren Templates**. Möglichkeiten:

1. Ein Sammel-Template `{{define "result-row-swap"}}` ruft Zeile, Podium (mit OOB-Flag) und Statistik auf.
2. Der Renderer bekommt `RenderPartials(w, set, names []string, data)` und hängt alles hintereinander.

Variante 1 ist explizit und testbar. Variante 2 ist flexibler. Das Arbeitsprojekt nutzt Variante 1
(`matching-form-swap`).

Das Podium-Template brauchst du einmal **mit** und einmal **ohne** `hx-swap-oob`
(Seite vs. Antwort). Wie vermeidest du doppeltes HTML? (Tipp: ein Wrapper-Template mit Flag,
oder getrennte Wrapper, die denselben Body aufrufen.)

### Flash vom Server

Das Highlight („grün blinken“) kommt **mit dem HTML** vom Server: Die Zeile wird mit
`State: "changed"` gerendert, das Template setzt `class="row-flash-success"`, eine CSS-Animation
erledigt den Rest. Kein JS, das nachträglich Klassen setzt. Das führt nur zu Race-Conditions
mit dem Swap.

## Bauplan

1. Fundament. Feature `race` mit Repository-Methoden für: Rennen + Wettbewerb, Starter mit Fahrernamen,
   Podium, Statistik, mögliche Starter (Fahrer, die noch nicht starten), `SetPosition`, `AddStarter`, `RemoveStarter`.
2. Service-Schicht `race.Service`: Regeln wie „Platz doppelt“ und „Status erlaubt Änderung“
   gehören hierher, nicht in den Handler und nicht ins Repository.
3. Seite mit allen Teilen, jedes Teil als eigenes `define`.
4. Platz ändern über Weg A, mit OOB für Podium und Statistik.
5. Starter hinzufügen (prepend) und entfernen. Das Select der möglichen Starter muss auch
   aktualisiert werden. Noch ein OOB-Kandidat.
6. Flash-CSS und `State` im Zeilen-ViewModel.
7. Fehler-Toast für doppelte Plätze: `HX-Trigger: {"crud-failed": "Platz 1 ist schon vergeben"}`
   + Listener in `app.js`. Und die Zeile? Antworte mit der **unveränderten** Zeile, damit das Input
   den alten Wert zeigt.

## Stolperfallen

- **OOB mit Tabellenzeilen:** Ein `<tr hx-swap-oob="true">` außerhalb einer Tabelle wirft der
  HTML-Parser weg. Für OOB-Zeilen in `<template>` einpacken. Podium und Statistik als `<div>` haben
  das Problem nicht.
- OOB-Elemente brauchen eine **id**, die es auf der Seite gibt, sonst passiert nichts (und htmx
  meldet `htmx:oobErrorNoTarget` in der Konsole).
- `hx-trigger="change"` am Input feuert bei Blur **und** bei Enter in manchen Browsern unterschiedlich.
  Teste beides. Enter in einem Input ohne Form tut gar nichts. `keyup[key=='Enter']` hilft.
- Ohne `[detail.id==12]` lädt **jede** Zeile neu, wenn irgendeine sich ändert. Funktioniert,
  ist aber genau das, was du vermeiden willst.
- `NULL` vs. `0` für die Platzierung: Das Schema erlaubt nur `> 0` oder `NULL`.

## Selbsttest

1. Wann reicht die direkte Antwort (Weg A), wann brauchst du Event + Zeilen-Endpoint (Weg B)?
2. Was macht `[detail.id==12]`, und woher kommt `detail`?
3. Wie findet htmx das Ziel eines OOB-Elements?
4. Warum kommt die Flash-Klasse vom Server und nicht aus JS?
5. Wo liegt die Regel „Platz doppelt“, und warum nicht im Handler?

## Wiedererkennen

Im Arbeitsprojekt: `CrudCreatedEvent` / `CrudRowUpdatedEvent` mit id, `.../row`-Endpoints mit
`State="isNew"/"isChanged"`, die Klassen `row-flash-new` / `row-flash-success`, und der lange
Kommentar in `events.go`, warum das Update **kein** `crud-changed` feuert. Dazu
`matching-form-swap` = Formular + `matching-preview-oob` in einer Antwort.
