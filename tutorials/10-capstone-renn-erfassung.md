# 10 – Capstone: Renn-Erfassung

**Branch:** `t10/renn-erfassung` → `t10/complete`
**Neu:** alles zusammen in einem realistischen Data-Entry-Flow, eigene `<base href>` pro Seite, Filter-Gedächtnis per Cookie, Rollen, Workflow-Status, Redirect per htmx, der zentrale Renderer in seiner „fertigen“ Form, Lesekarte für fremde Projekte

## Worum es geht

Ein Erfassungs-Werkzeug für Rennergebnisse, so wie Data-Entry-Tools in echten Projekten aussehen.

**Seite 1: Übersicht der Rennen** (`/entry`)
- Tabelle aller `competition_instances`: Wettbewerb, Ort, Land, Datum, Status, Starter, gewertet.
- Filter: Suche, Disziplin, Status, „nur offene“ (nicht `finished`/`cancelled`), „letzte N Tage“.
- Sortierung + Pagination (aus 04).
- **Filter werden gemerkt**: Wer aus der Erfassung zurückkommt, sieht dieselbe Übersicht wieder.

**Seite 2: Erfassung eines Rennens** (`/entry/{raceID}`)
- Links: Starterliste mit Tastaturnavigation (aus 09). Die ausgewählte Zeile ist markiert.
- Rechts: Formular für den ausgewählten Starter: Platz, gefahrenes Bike (Picker über **seine** Bikes, aus 08),
  Notiz. Darunter eine Vorschau: Podium + „Platz 4 ist noch frei“-Hinweise.
- Starter hinzufügen über Fahrer-Picker mit „Neu“ (gestapeltes Modal, aus 08).
- Button **„Rennen abschließen“**: setzt Status `finished`, danach zurück zur Übersicht.
- Footer-Button **„Regeln“**: Panel mit den Erfassungsregeln (aus einer Markdown/HTML-Datei).

**Rollen:** `viewer` sieht alles read-only, `editor` darf erfassen. (Rolle aus dem Cookie wie in 09.)

**Vorarbeit an den Daten** (auch das ist Teil der Übung):
- `rider_competition_instances` hat noch keine Spalten für Bike und Notiz. Erweitere das Schema
  (`bike_id INTEGER REFERENCES bikes(id) ON DELETE SET NULL`, `note TEXT`). Überleg dir, wie eine
  bestehende `db.sql` die neuen Spalten bekommt: neu anlegen oder eine kleine Migration beim Start
  (`PRAGMA table_info` / `ALTER TABLE … ADD COLUMN`)?
- Fast alle Rennen in den Seed-Daten sind `finished`. Setz zwei davon auf `running` (Seed anpassen)
  und gib dem `editor` zusätzlich „Rennen wieder öffnen“. Sonst gibt es nichts zu erfassen.

## Abnahme

### Übersicht
- [ ] Filter, Sort, Seite stehen in der URL und werden zusätzlich in einem Cookie gemerkt.
- [ ] `/entry` ohne Parameter stellt den gemerkten Filter wieder her.
- [ ] Gemerkte Seite liegt hinter dem Ende (nach Filteränderung woanders) → letzte Seite.
- [ ] Klick auf eine Zeile oder Enter → Erfassungsseite.

### Erfassung
- [ ] Alle relativen URLs der Seite (`table`, `row`, `{id}/edit`, …) lösen sich unter
      `/entry/{raceID}/` auf, durch `<base href>` im Layout.
- [ ] Zeile auswählen → Formular rechts **und** Vorschau aktualisieren sich in **einem** Request (OOB).
- [ ] Speichern → nur die Zeile (Flash) + Vorschau aktualisieren sich, Fokus springt zur nächsten Zeile.
- [ ] Ungespeicherte Änderungen + andere Zeile wählen → Nachfrage.
- [ ] Bike-Picker zeigt nur Bikes **dieses** Fahrers. „Neu“ legt ein Bike an **und** ordnet es dem Fahrer zu.
- [ ] „Rennen abschließen“ nur, wenn alle Starter einen Platz oder bewusst DNF haben.
      Sonst Toast mit Begründung.
- [ ] Nach dem Abschließen: `HX-Redirect` zur Übersicht, die den gemerkten Filter zeigt.
- [ ] Rennen `finished`: Seite read-only, Hinweis oben.

### Querschnitt
- [ ] `viewer` sieht keine Bearbeiten-Elemente **und** der Server lehnt Schreib-Requests mit 403 ab.
- [ ] Jede Route ist durch die Rollen-Middleware geschützt. Prüf das mit einem Test, der alle Routen durchgeht.
- [ ] Alle Templates werden beim Start geparst und geprüft. Ein Tippfehler im Template-Namen bricht den **Start** ab.
- [ ] Jedes Fragment hat einen Render-Test.

## Der zentrale Renderer in fertiger Form

Hier setzt du um, was du dir für die Lib vorgenommen hast:

| Aufgabe | Ergebnis |
| ------- | -------- |
| Pro **Seite** ein Set (`Clone` von Base + Komponenten + Seiten-Templates) | Kein `content`-Dispatcher, keine Namenskollisionen |
| Sets werden **beim Start** gebaut, `Clone` vor jeder Ausführung | Fehler beim Start statt beim Request |
| Pflicht-Templates prüfen (`layout`, `content`; deklarierte Fragmente existieren) | Tippfehler fallen sofort auf |
| Komponenten (Picker, Modal, Tabelle, Pagination) registrieren sich am Base-Set | Features nutzen sie ohne eigene Verschachtelung |
| Seitenrahmen (Titel, Navigation, Base-Href, aktueller Benutzer, Regeln-URL) wird **an einer Stelle** aus Request + Seite gefüllt | Kein Handler vergisst ein Feld |
| `RenderPage` / `RenderFragment` / `RenderError`, alle gepuffert, alle mit Status | Ein Fehlerpfad, keine halben Antworten |
| `Vary: HX-Request`, History-Restore-Request → ganze Seite | Zurück-Button funktioniert |

Überleg vorher: Wie **benennt** ein Handler sein Fragment, ohne Strings zu verstreuen?
Mögliche Antworten: Konstanten pro Feature, ein kleiner Typ `Fragment` mit Rückgabe aus der
Registrierung, oder eine Methode pro Fragment am Feature. Entscheide dich und schreib die Begründung
in den Paketkommentar des Renderers.

## Zusammenhänge, die neu sind

### `<base href>`

Die Erfassungsseite setzt `<base href="/entry/7/">`. Dann bedeutet `hx-get="table"` automatisch
`/entry/7/table`. Die Templates werden dadurch kürzer und kennen die Rennen-ID nicht.

Aber: **Alles**, was nicht unter dieser Basis liegt (Modals, Picker, Regeln, Statics), braucht eine
**absolute** URL. Sonst landet der Picker auf `/entry/7/bikes/picker`. Das ist die häufigste
Fehlerquelle in solchen Projekten. Halte die Regel im Code fest: Hilfsfunktionen `raceURL(id)` und
`entryURL()`, und Komponenten bekommen immer absolute URLs.

### Filter-Gedächtnis

```
GET /entry/table?…  (Tabellen-Request)  → Filter in Cookie speichern (nur erlaubte Keys!)
GET /entry          (Seite, ohne Query) → Cookie lesen → Formular vorbelegen → Tabelle lädt mit hx-trigger="load"
```

URL **und** Cookie: Die URL ist für Links und Reload, das Cookie für „zurück von woanders“.
Speichere nur eine Whitelist von Keys und codier sie neu. Das Cookie ist Benutzereingabe.
`Path` des Cookies auf `/entry` setzen.

### `HX-Redirect` vs. `HX-Location` vs. `HX-Refresh`

| Header | Wirkung |
| ------ | ------- |
| `HX-Redirect: /entry` | voller Seitenwechsel (wie `window.location`) |
| `HX-Location: /entry` | htmx-Navigation ohne Full Reload (Body austauschen, URL pushen) |
| `HX-Refresh: true` | aktuelle Seite neu laden |
| `302` + `Location` | folgt der **Browser** intern, htmx bekommt das Ziel-HTML in den Swap. Fast nie, was du willst |

Für „abschließen → Übersicht“ ist `HX-Redirect` richtig. Warum ist ein normales `302` hier falsch?

### Read-only aus einer Quelle

Rolle `viewer` und Status `finished` führen beide zu „nur lesen“. Berechne **ein** Flag `ReadOnly`
im ViewModel. Das Template fragt nur `{{if .ReadOnly}}`. Der Service prüft **unabhängig davon**
noch einmal: Das Template ist nur Anzeige, keine Sicherheit.

## Bauplan (Vorschlag)

1. Renderer in fertiger Form + Tests für den Renderer selbst (fehlendes Template → Startfehler).
2. Middleware: aktueller Benutzer + Rolle. `RequireRole("editor")` als Wrapper.
3. Übersicht: Filter-Struct, Sort, Pagination, Cookie-Gedächtnis.
4. Erfassungsseite: Gerüst mit `<base href>`, Starterliste, Auswahl → Formular + Vorschau (OOB).
5. Speichern mit Zeilen-Update, Fokus auf nächste Zeile (`HX-Trigger-After-Swap` + JS).
6. Picker für Bikes (gefiltert), „Neu“-Modal, Starter hinzufügen.
7. Abschließen mit Regelprüfung im Service + `HX-Redirect`.
8. Regeln-Panel, ungespeicherte-Änderungen-Schutz.
9. Route-Test: alle Schreib-Routen mit `viewer` → 403.

## Lesekarte: ein fremdes htmx-Projekt verstehen

Das ist das eigentliche Ziel aller zehn Projekte. Wenn du ein Projekt wie dieses (oder dein
Arbeitsprojekt) zum ersten Mal öffnest, geh in dieser Reihenfolge vor:

1. **Router lesen.** Jede Route einer von vier Arten zuordnen:
   - **Seite** (rendert `layout`)
   - **Fragment** (rendert ein benanntes Template)
   - **Aktion** (POST/PUT/DELETE: antwortet mit Fragment und/oder Events)
   - **JSON** (für JS)
2. **Pro Seite das Template-Set finden.** Wie wird es gebaut (Base + Komponenten + eigene)?
   Wo kommt `content` her?
3. **Die Seiten-Templates nach `hx-` durchsuchen.** Für jedes `hx-get/post/put/delete`: Welche
   Route, welches Target, welcher Swap, welcher Trigger?
4. **Events kartieren.** Suche nach `HX-Trigger` / `SetTrigger` im Go-Code (**Sender**) und nach
   `from:body` / `addEventListener` in Templates und JS (**Empfänger**). Schreib dir eine
   Tabelle: Event → wer sendet → wer hört → was passiert.
5. **OOB-Stellen suchen** (`hx-swap-oob`). Welche Antwort aktualisiert heimlich noch etwas anderes?
6. **Base-Href und absolute URLs prüfen.** Welche Seite setzt welche Basis?
7. **JS zuletzt.** Es sollte nur Kleber sein. Ist es mehr, ist das ein Hinweis auf eine Stelle,
   die man sich genauer anschauen muss.

Mach die Lesekarte einmal für dein **eigenes** Capstone-Projekt (schriftlich, als `MAP.md` im Branch)
und einmal für den sloganmatching-Usecase im Arbeitsprojekt. Wenn beide Karten aufgehen, bist du nicht mehr lost.

## Selbsttest

1. Was löst eine relative URL in einer Seite mit `<base href="/entry/7/">` auf, und welche URLs müssen absolut sein?
2. Warum URL **und** Cookie für den Filter?
3. Warum `HX-Redirect` statt `302`?
4. Wo wird `ReadOnly` berechnet, und warum reicht es nicht, Buttons auszublenden?
5. Welche Fehler fängt dein Renderer jetzt beim Start ab, die er in Projekt 01 erst beim Request bemerkt hat?
6. Zeichne den Ablauf „Zeile auswählen“ vom Tastendruck bis zum Einsetzen von Formular und Vorschau.

## Wiedererkennen

Dieses Projekt **ist** im Kern sloganmatching: Prospekt-Auswahl ↔ Übersicht, `filterCookie` +
`saveFilter`/`loadFilter` ↔ Filter-Gedächtnis, `matchingBaseHref` ↔ `<base href>`,
`usecasePath` für absolute URLs, `matching-form-swap` ↔ Formular + Vorschau per OOB,
`PutComplete` ↔ „abschließen“, `ui.RulesHandler` ↔ Regeln-Panel, `RequireRoleAndCountryHtmx`
↔ Rollen-Middleware. Der Unterschied: Bei dir gibt es einen zentralen Renderer und
pro Seite ein eigenes Template-Set. Genau das ist dann der Vorschlag für die Lib.
