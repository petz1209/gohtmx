# 03 – Suche & Filter

**Branch:** `t03/suche-filter` → `t03/complete`
**Neu:** Active Search, `hx-trigger` mit Modifiern, `hx-include`, `hx-sync`, `hx-indicator`, Zustand in der URL, dynamisches `WHERE`

## Worum es geht

Fahrer durchsuchen: Freitext (Name/Spitzname), Filter nach Disziplin und Team.
Die Ergebnisse aktualisieren sich **während des Tippens**. Die URL enthält immer den
aktuellen Filter, also kannst du die Ansicht als Link verschicken.

## Abnahme

- [ ] Suchfeld: Nach 300 ms Tipp-Pause wird gesucht, nicht bei jedem Tastendruck.
- [ ] Disziplin-Select und Team-Select lösen **sofort** eine Suche aus.
- [ ] Alle drei Filter wirken **zusammen** („downhill“ + Team „YT Mob“ + „hö“ → Vali Höll).
- [ ] Die URL zeigt den Filter: `/riders?q=hö&discipline=downhill&team=3`.
- [ ] Reload auf diese URL zeigt dasselbe Ergebnis **und** die Formularfelder sind vorbelegt.
- [ ] Während der Suche erscheint ein Lade-Indikator.
- [ ] Keine Treffer → freundlicher Leer-Zustand („Keine Fahrer gefunden“ + Link „Filter zurücksetzen“).
- [ ] Schnelles Tippen: Eine langsame alte Antwort überschreibt **nie** eine neuere.
- [ ] Ein ungültiger Wert (`?discipline=bmx`) wird ignoriert, kein 500.

## Zusammenhänge

### Der Zustand lebt in der URL

Der Filter ist **kein** Server-State und kein JS-State. Er steht in der Query-String-URL.
Daraus folgt:

- Ein einziges Go-Struct `RiderFilter` mit `ParseRiderFilter(url.Values) RiderFilter`.
  Hier wird validiert und normalisiert (trimmen, ungültige Disziplin → leer).
- Die **Seite** (`GET /riders`) parst den Filter, belegt das Formular vor und rendert die
  Ergebnisse direkt mit.
- Das **Fragment** (`GET /riders/results` oder derselbe Endpoint mit htmx) parst **denselben** Filter.

Ein Parser, eine Wahrheit. Wenn Seite und Fragment unterschiedlich parsen, sieht der Benutzer
nach einem Reload etwas anderes als vorher.

### Das Formular als Request-Quelle

```html
<form id="filter" hx-get="/riders/results" hx-trigger="change"
      hx-target="#results" hx-push-url="true" hx-sync="this:replace" hx-indicator="#spinner">
  <input type="search" name="q" value="{{.Filter.Q}}"
         hx-get="/riders/results" hx-trigger="input changed delay:300ms, search"
         hx-include="#filter" hx-sync="#filter:replace">
  <select name="discipline">…</select>
  <select name="team">…</select>
</form>
```

Die Teile:

- **Form** `hx-trigger="change"`: Die Selects feuern `change`, das Event **blubbert** zum Form
  hoch. Ein Request **vom Form** schickt automatisch **alle** Felder mit.
- **Input** hat einen eigenen Request:
  - `input changed delay:300ms`: bei Eingabe, nur wenn sich der Wert wirklich geändert hat,
    und erst 300 ms nach dem **letzten** Event (Debounce).
  - `search`: feuert, wenn man das Suchfeld mit dem ✕ leert.
  - Ein Request vom Input schickt nur **dessen** Wert. `hx-include="#filter"` nimmt die
    restlichen Felder mit.
- `hx-target`, `hx-push-url`, `hx-indicator` werden vom Form **geerbt**. `hx-get` nicht, der
  steht am Element, das den Request auslöst.
- `hx-sync="this:replace"` am Form und `hx-sync="#filter:replace"` am Input: Beide teilen sich
  dieselbe Warteschlange. Läuft schon ein Request, wird er abgebrochen und durch den neuen
  ersetzt. Das löst das Race aus der Abnahme.
- Nebeneffekt: Verlässt du das Suchfeld, feuert es zusätzlich `change`, und das Form sucht noch
  einmal mit denselben Werten. Harmlos. Überleg dir, ob und wie du das verhindern willst.

### Die URL beim Push

`hx-push-url="true"` würde `/riders/results?q=…` in die Adresszeile schreiben, also die
Fragment-URL. Ein Reload darauf zeigt ein nacktes Fragment. Zwei Lösungen:

1. **Derselbe Endpoint** für Seite und Fragment (`GET /riders`, Entscheidung über `HX-Request`
   wie in 02). Dann ist `hx-push-url="true"` korrekt.
2. **Getrennte Endpoints**, und der Server setzt die Seiten-URL per Response-Header
   `HX-Push-Url: /riders?q=…`.

Variante 2 ist sauberer, wenn Fragment und Seite verschiedene Daten brauchen. Variante 1 ist
weniger Code. Entscheide dich und begründe es.

### Dynamisches WHERE, aber sicher

```
WHERE 1=1
  [AND (first_name LIKE ? OR last_name LIKE ? OR nickname LIKE ?)]
  [AND discipline = ?]
  [AND team_id = ?]
```

Bedingungen werden **als feste Strings** angehängt, Werte **immer** als Platzhalter in eine
`args []any`. Niemals einen Benutzerwert in den SQL-String kleben.

Hinweis: In SQLite ist `LIKE` nur für ASCII case-insensitive. „hö“ findet „Höll“, „HÖ“ aber nicht.
Gut zu wissen, für dieses Projekt nicht schlimm.

## Bauplan

1. Fundament + Fahrerliste (Liste reicht, kein Detail-Panel nötig).
2. `RiderFilter` + `ParseRiderFilter` + Tests dafür (Table-Driven Tests, ganz ohne HTTP).
3. Repository `Search(ctx, RiderFilter)`.
4. Template: Formular (vorbelegt), `rider-results` als eigenes `define`, Leer-Zustand darin.
5. Handler + Push-URL-Entscheidung.
6. Lade-Indikator: htmx setzt während des Requests die Klasse `htmx-request` auf das Element
   aus `hx-indicator`. CSS: `.htmx-indicator{opacity:0}` und `.htmx-request .htmx-indicator, .htmx-request.htmx-indicator{opacity:1}`.
7. Teams für das Select aus der DB laden (nicht hart codieren).

## Stolperfallen

- `delay` ohne `changed`: Pfeiltasten im Feld lösen Requests aus, obwohl sich nichts ändert.
- `<select>` vorbelegen: `{{if eq .Value $.Filter.Discipline}}selected{{end}}`. Innerhalb von
  `range` brauchst du `$`, um an den Filter zu kommen.
- Team-ID im Filter ist ein `int`, im Template ist es ein String aus der Query. Vergleich
  `eq` mit verschiedenen Typen scheitert. Normalisiere im ViewModel.
- Ein leeres `q` darf **nicht** `LIKE '%%'` erzeugen. Das funktioniert zwar, ist aber unnötig.
  Bedingung einfach weglassen.

## Selbsttest

1. Warum schickt ein Request vom `<form>` alle Felder mit, einer vom `<input>` aber nur dessen Wert?
   Welche Attribute erbt der Input vom Form, welche nicht?
2. Was genau macht `delay:300ms`, und was `throttle:300ms`?
3. Welches Problem löst `hx-sync="this:replace"`? Beschreib die Situation ohne.
4. Warum ist die URL der richtige Ort für den Filter, nicht ein Cookie und nicht JS?
5. Was passiert bei Reload auf eine gepushte Fragment-URL?

## Wiedererkennen

Im Arbeitsprojekt heißt die Suche `#search-form`. Seite und Tabelle teilen `Overview`
(TableURL, Sort, Pagination). Zusätzlich wird der Filter dort in einem **Cookie** gemerkt,
damit „Zurück“ von einer ganz anderen Seite den Filter wiederherstellt. Das kommt in Projekt 10.
