# Tutorials: Go + htmx von einfach bis komplex

Zehn Projekte, alle in der Mountainbike-Domain (Teams, Fahrer, Bikes, Rennen, Ergebnisse, Follows).
Jedes Projekt baust du **from scratch** auf einem eigenen Branch. Die Tutorials enthalten
**keine fertigen Lösungen**, sondern:

- was am Ende funktionieren muss (Abnahme-Checkliste),
- die Zusammenhänge dahinter,
- einen Bauplan mit Hinweisen und kleinen Syntax-Schnipseln,
- typische Stolperfallen,
- Selbsttest-Fragen: Wenn du sie ohne Nachschauen beantworten kannst, hast du es verstanden.
- „Wiedererkennen“: wo dir das Muster in großen Projekten (wie deinem Arbeitsprojekt) begegnet.

## Übersicht

| #  | Projekt | Neu dabei |
| -- | ------- | --------- |
| 01 | [Fahrerliste](01-fahrerliste.md) | Request-Zyklus, `embed`, Template-Sets, Layout/Content, Repository, ViewModel |
| 02 | [Fahrer-Details nachladen](02-fahrer-details.md) | erstes htmx: `hx-get`, `hx-target`, `hx-swap`, Seite vs. Fragment, `hx-push-url` |
| 03 | [Suche & Filter](03-suche-filter.md) | Active Search, `hx-trigger`, `hx-include`, `hx-sync`, Zustand in der URL |
| 04 | [Bike-Katalog: Sortierung & Pagination](04-bike-katalog.md) | wiederverwendbare Tabellen-Partials, Sort-Whitelist, Pagination, `FuncMap` |
| 05 | [Teams verwalten: Inline-CRUD](05-teams-crud.md) | PUT/DELETE, Click-to-Edit, Validierung, Statuscodes, Buffer, Render-Tests |
| 06 | [Fahrer im Modal](06-fahrer-modal.md) | Modals, `HX-Trigger`-Events, Entkopplung über `from:body` |
| 07 | [Rennergebnisse eintragen](07-rennergebnisse.md) | Zeilen-Updates, Event-Filter, Prepend, OOB-Swaps, Flash |
| 08 | [Bikes zuordnen: Picker & gestapelte Modals](08-bike-picker.md) | Autocomplete-Komponente, n:m, Modal-Stack, Event-Payloads |
| 09 | [Follows, Tastatur & eigenes JS](09-follows-js.md) | htmx-Lifecycle, `htmx.onLoad`, Toasts, Middleware, Request-Kontext |
| 10 | [Capstone: Renn-Erfassung](10-capstone-renn-erfassung.md) | alles zusammen + zentraler Renderer „fertig“ + Lesekarte für fremde Projekte |

Die Projekte bauen **inhaltlich** aufeinander auf. Was du in 02 lernst, wird in 05 vorausgesetzt.
Den **Code** schreibst du trotzdem jedes Mal neu.

## Branches

Schema wie im Haupt-README: Arbeits-Branch und Ergebnis-Branch pro Projekt.

| Projekt | WorkBranch | ResultBranch |
| ------- | ---------- | ------------ |
| 01 | `t01/fahrerliste` | `t01/complete` |
| 02 | `t02/fahrer-details` | `t02/complete` |
| …  | `tNN/<name>` | `tNN/complete` |

Startpunkt für jeden Branch ist ein Stand, der nur Schema und Daten enthält
(`ddl.sql`, `init.sql`, `datasets/`). Am einfachsten machst du dir dafür einmal einen Branch
`base/data` und zweigst jedes Tutorial davon ab.

## Das Fundament (baust du jedes Mal neu)

Jedes Projekt braucht dieselbe Grundlage. Am Anfang dauert das eine Stunde, ab Projekt 4
vielleicht noch 15 Minuten. **Genau das ist die Übung**: Danach kannst du das Fundament
jedes fremden Projekts auf einen Blick lesen.

```
src/
  cmd/app/main.go        DB öffnen, Renderer bauen, Features verbinden, Server starten
  cmd/app/routes.go      (optional) alle Features an den Mux hängen
  internal/database/     Open(): SQLite öffnen, Schema + Daten beim ersten Start laden
  internal/render/       der Renderer (wächst von Projekt zu Projekt)
  internal/feature/<x>/  handler.go, repository.go, model.go, templates/*.html
  ui/                    Layout-Templates + static (css, js, htmx.min.js)
```

Reihenfolge, in der du es aufbaust:

1. `go mod init`, SQLite-Treiber (`modernc.org/sqlite`).
2. `database.Open` mit eingebettetem Schema und Seed-Daten.
3. `ui`: `layout` mit `{{template "content" .}}`, `title`, htmx einbinden, statische Dateien.
4. Renderer: Base-Set aus `ui` parsen, `Register(key, fsys, patterns...)` klont Base pro Feature.
5. Erstes Feature: `New(...)` registriert seine Templates, `Connect(mux)` hängt seine Routen ein.
6. `main` verbindet alles.

## Der rote Faden: ein htmx-Request

Alle Tutorials drehen sich um diesen Kreislauf. Wenn du ihn verstanden hast, verstehst du jedes
htmx-Projekt.

```
 Browser                         Server (Go)
 ───────                         ───────────
 Event (click, input, load,      
 eigenes Event "crud-changed")   
        │                        
        ▼                        
 htmx baut Request ───────────▶  ServeMux: Methode + Pfad → Handler
  hx-get/post/put/delete          │
  + Header HX-Request: true       ▼
  + Werte (Form, hx-include,     Handler: Eingaben parsen + validieren
    hx-vals)                      │
                                  ▼
                                 Service/Repository: DB lesen/schreiben
                                  │
                                  ▼
                                 Renderer: Template-Set + Name + Daten → HTML
                                  │      (+ Response-Header: HX-Trigger, HX-Redirect…)
        ◀─────────────────────────┘
 htmx nimmt die Antwort:
  1. Status prüfen (4xx/5xx: standardmäßig KEIN Swap)
  2. HX-Trigger-Events feuern
  3. HTML in hx-target einsetzen (hx-swap)
  4. hx-swap-oob-Elemente an ihre ids setzen
  5. neues HTML wird "lebendig" (htmx.onLoad, neue hx-Attribute)
```

## Vokabular

| Begriff | Bedeutung |
| ------- | --------- |
| **Seite (Page)** | komplettes HTML mit Layout. Antwort auf normale Browser-Navigation. |
| **Fragment (Partial)** | ein Ausschnitt ohne Layout. Antwort auf htmx-Requests. |
| **Template-Set** | ein `*template.Template` mit allen benannten Templates, die zusammen ausgeführt werden können. |
| **Target** | das Element, in/an das htmx die Antwort setzt (`hx-target`). |
| **Swap** | wie eingesetzt wird: `innerHTML`, `outerHTML`, `beforeend`, `afterbegin`, `delete`, `none`. |
| **Trigger** | was den Request auslöst (`hx-trigger`). |
| **Event** | vom Server per `HX-Trigger`-Header ausgelöstes Browser-Event. Andere Elemente können darauf hören. |
| **OOB** | Out-of-Band: zusätzliche Elemente in der Antwort, die *nicht* ins Target gehen, sondern an ihre eigene `id`. |
| **ViewModel** | Go-Struct, das genau das enthält, was ein Template braucht, unabhängig vom DB-Modell. |

## Regeln, die durch alle Projekte gelten

- `html/template`, nie `text/template`.
- Templates werden **beim Start** geparst. Fehler beim Start sind gut, Fehler beim Request schlecht.
- Jede Seite muss **auch ohne htmx** per direkter URL funktionieren (Reload, Lesezeichen, neuer Tab).
- Der Server entscheidet, was angezeigt wird. JS ist nur Kleber.
- SQL immer mit Platzhaltern (`?`). Spalten für `ORDER BY` nur über eine Whitelist.
