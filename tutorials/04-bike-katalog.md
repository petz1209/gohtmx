# 04 – Bike-Katalog: Sortierung & Pagination

**Branch:** `t04/bike-katalog` → `t04/complete`
**Neu:** sortierbare Spaltenköpfe, Sort-Whitelist, Pagination-Struct, wiederverwendbare Tabellen-Partials, `FuncMap`, Zustand aus mehreren Quellen kombinieren

## Worum es geht

Ein Katalog aller Bikes als Tabelle: Marke, Modell, Jahr, Kategorie, Laufrad, Federweg vorne/hinten,
Gewicht. Sortierbar per Klick auf den Spaltenkopf, blätterbar, filterbar nach Kategorie.
Es gibt nur 7 Bikes, also nimm Seitengrößen von **3, 5 und 10**, damit du Pagination überhaupt siehst.

Das Besondere: Die Tabellen-Bausteine (Kopf, Pagination) baust du **generisch**, sodass eine zweite
Tabelle sie ohne Kopieren nutzen kann. Als zweite Tabelle dienen die Teams (nur Liste, ohne Filter).

## Abnahme

- [ ] Klick auf „Gewicht“ sortiert aufsteigend, erneuter Klick absteigend. Ein Pfeil zeigt die Richtung.
- [ ] Nur erlaubte Spalten sind sortierbar. `?sort=; DROP TABLE bikes` sortiert einfach nach Default.
- [ ] Pagination: „Seite 2 von 3“, Vor/Zurück, Seitengröße wählbar.
- [ ] Kategorie-Filter + Sortierung + Seite wirken **gemeinsam** und stehen alle in der URL.
- [ ] Filter ändern → zurück auf Seite 1.
- [ ] `?page=99` zeigt die letzte Seite, nicht eine leere Tabelle.
- [ ] Die Team-Tabelle benutzt **dieselben** Templates für Kopf und Pagination.
- [ ] Reload auf jede URL ergibt dieselbe Ansicht.

## Zusammenhänge

### Warum der Server den nächsten Sort-Zustand berechnet

Der Spaltenkopf „Gewicht“ muss wissen: Ist gerade nach Gewicht sortiert, und in welche Richtung?
Daraus folgt der Link für den **nächsten** Klick. Das gehört nach Go, nicht ins Template:

```go
type SortHeader struct {
    Label    string
    Key      string // "weight"
    Sortable bool
    Active   bool   // gerade danach sortiert?
    Desc     bool   // aktuelle Richtung
    NextSort string // "-weight" oder "weight"
}
```

Das Template zeigt nur noch an. Konvention zum Beispiel: `weight` = aufsteigend, `-weight` = absteigend.

### Die Whitelist

```go
var bikeSortColumns = map[string]string{
    "brand":  "brand",
    "year":   "model_year",
    "weight": "weight_kg",
    …
}
```

Der Key kommt aus der URL, der Wert geht in das SQL. `ORDER BY` kann **keinen** Platzhalter
haben. Deshalb ist diese Map die einzige Sicherung gegen SQL-Injection an dieser Stelle.
Unbekannter Key → Default-Sortierung. Hänge immer `, id` als zweiten Sortierschlüssel an, sonst
ist die Reihenfolge bei gleichen Werten zufällig, und Pagination springt.

### Pagination

```go
type Pagination struct {
    Page, PageSize, Total int
}
func (p Pagination) Offset() int
func (p Pagination) TotalPages() int
func (p Pagination) HasPrev() bool
func (p Pagination) HasNext() bool
```

Ablauf im Service:
1. `COUNT(*)` mit **demselben** `WHERE` wie die Datenabfrage.
2. Ist `Page > TotalPages`, auf die letzte Seite klemmen.
3. Datenabfrage mit `LIMIT ? OFFSET ?`.

Die `PageSize` kommt aus der URL. Auch hier Whitelist (`3, 5, 10`), sonst fragt jemand
`pageSize=1000000` an.

### Mehrere Zustände in einem Formular

Filter, Sortierung, Seite und Seitengröße müssen bei **jedem** Request zusammen geschickt werden.
Bewährtes Muster: Ein Formular enthält alles, Sort und Page als **hidden inputs**:

```html
<form id="bike-filter" hx-get="/bikes" hx-target="#bike-table" hx-push-url="true">
  <select name="category">…</select>
  <input type="hidden" name="sort" value="{{.Sort}}">
  <input type="hidden" name="page" value="1">
  <select name="pageSize">…</select>
</form>
```

Ein Klick auf den Spaltenkopf oder „Weiter“ schickt dann das Formular **plus** einen
überschriebenen Wert:

```html
<button hx-get="/bikes" hx-include="#bike-filter" hx-vals='{"sort": "{{.NextSort}}"}'>…</button>
```

`hx-vals` gewinnt gegen gleichnamige Felder aus `hx-include`. Prüf das im Netzwerk-Tab nach.

Frage dazu: Wenn die Tabelle per htmx neu geladen wird, das Formular aber nicht, steht im
hidden `sort` noch der **alte** Wert. Wie löst du das? (Tipp: Gehören die hidden inputs
vielleicht **in** das Fragment, das neu geladen wird? Oder ein OOB-Swap, siehe 07.)

### Generische Partials

Das Tabellen-Fragment ruft `{{template "sort-header" .}}` und `{{template "pagination" .}}`
auf. Diese Templates liegen im **Base-Set** (in `ui`), nicht im Feature. Jedes Feature kann sie
benutzen, weil jedes Feature-Set ein Klon von Base ist.

Damit ein generisches Template funktioniert, braucht es eine **feste Datenform**: Struct-Typen
wie `SortHeader` und `Pagination` leben im Renderer- oder ui-Paket, nicht im Feature.

### FuncMap

Für Kleinigkeiten wie „Gewicht mit einer Nachkommastelle“ oder „Seiten-Liste 1…n“ brauchst
du Funktionen. Sie müssen **vor** dem Parsen registriert sein, also am Base-Set, bevor
`ParseFS` aufgerufen wird. Klone erben sie.

Halte die FuncMap klein. Alles, was mit Geschäftslogik zu tun hat, gehört ins ViewModel.

## Bauplan

1. Fundament. Renderer bekommt eine `FuncMap` am Base-Set.
2. `ui`-Paket: `SortHeader`, `Pagination`, Templates `sort-header` und `pagination`.
3. `ParseSort(raw, whitelist)` und `ParsePage`, `ParsePageSize`, mit Tests.
4. Bike-Repository: `Count(ctx, filter)` und `List(ctx, filter, sort, pagination)`.
5. Bike-Feature: Seite + Tabellen-Fragment. Spaltenköpfe als `[]SortHeader` im Handler bauen.
6. Team-Feature: nur Liste mit denselben Partials. Hier merkst du, ob die Partials wirklich generisch sind.

## Stolperfallen

- `ORDER BY ?` mit Platzhalter sortiert **nach einer Konstante**, also gar nicht. Kein Fehler, nur falsch.
- `COUNT` mit anderem `WHERE` als die Daten → „Seite 3 von 2“.
- Bei `Total = 0`: `TotalPages` ist 0, aber die aktuelle Seite ist 1. Die Pagination nicht kaputt anzeigen.
- `hx-vals` ist JSON. Werte aus dem Template darin sind in einem JS-Kontext.
  `html/template` escaped das richtig, aber probier aus, was mit einem `"` passiert.
- `range` über Seitenzahlen: Templates haben keine Zahlen-Schleife. Liefere `[]int` aus Go.

## Selbsttest

1. Warum kann `ORDER BY` keinen Platzhalter haben, und was ist die Konsequenz?
2. Warum braucht es `, id` als zweiten Sortierschlüssel?
3. Wo liegt `SortHeader`, und warum nicht im Bike-Feature?
4. Was passiert mit Sortierung und Filter, wenn du auf „Seite 2“ klickst? Verfolge die Werte vom Klick bis zur SQL-Abfrage.
5. Warum muss die `FuncMap` vor dem Parsen registriert werden?

## Wiedererkennen

Im Arbeitsprojekt: `ui.SortHeader`, `ui.Pagination` mit `Beyond()`, `ui.ParsePageSize` mit
`AllowedPageSizes`, die `sortdbmap` im Repository und die Templates `table-sort-header` und
`overview-table-pagination`. Genau dieser Aufbau. Das „Zurückklemmen auf die letzte Seite“
macht dort `loadProspektSelectPage` mit `result.Beyond()`.
