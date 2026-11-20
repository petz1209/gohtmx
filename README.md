# gohtmx
learning the go-htmx stack

Given that AI has become so good at creating frontend applications, Everybody can simply fire up a session and spit out this type of stack.  
However, personally I have always been interested to understand how things work under the hood.  
I have a number of go services and systems, so I would say I am fairly familiar with go as as backend language. But I have never written an htmx project by hand.  
Today this changes. I am creating this project mostly to track my own progress with htmx and it's go integration.


# Framework-Dokumentation

Dieser Abschnitt beschreibt, wie die Web-UI aktuell aufgebaut ist und wie die Teile
zusammenkommen. Zielgruppe: Entwickler und KI-Agenten, die ein Feature hinzufügen oder
das Framework verstehen wollen.

Ziel des Frameworks: **Ein neues Feature (eine neue Seite) hinzuzufügen soll einfach sein,
ohne das Wiring im Hintergrund verstehen zu müssen.** Gleichzeitig soll alles, was ein
Feature ins System einhängt, im Feature selbst sichtbar sein.

Stack: Go (nur Standardbibliothek: `net/http`, `html/template`, `embed`), htmx, SQLite.

## Überblick

```
main.go
  ├─ ui.NewApp()                       globale Templates (Layout, Partials, Components)
  ├─ mux.Handle(ui.StaticPrefix, ...)  globale Static Files unter /static/
  └─ home.Register(mux, app)           pro Feature eine Zeile
        │
        ├─ app.NewFeature("home", ...) Template-Sets pro Seite + Static-Handler
        ├─ view.Page("home")           Seite beim Start holen
        ├─ mux.Handle(/assets/home/)   Feature-Static mounten
        └─ mux.HandleFunc(route, h)    Routen → Handler → service
                                              │
                                              └─ page.Render(w, r, ui.PageData{...})
```

Die Verantwortung ist so verteilt:

| Teil | Zuständig für | Nicht zuständig für |
|---|---|---|
| **Router** (heute `http.ServeMux`, später restify/`web.API`) | Routing, Middleware, Server | Templates |
| **`internal/ui`** | Templates parsen, Seiten rendern, Static-Handler liefern | Routing (besitzt **keinen** Mux) |
| **Feature** (`internal/features/<name>`) | eigene Seiten, Partials, Static Files, Routen, Handler, Service | globale Dinge |

## Verzeichnisstruktur

```
src/
  cmd/app/main.go                 Einstiegspunkt: DB, ui.NewApp, Features registrieren, Server
  internal/
    database/                     SQLite-Zugriff
    ui/                           das Framework
      assets.go                   //go:embed static templates
      app.go                      App, NewApp, StaticHandler, StaticPrefix
      feature.go                  Feature, NewFeature, Page(), StaticHandler()
      page.go                     Page, Render
      model.go                    PageData (Template-Daten)
      doc.go                      Kurzfassung des Vertrags (go doc)
      static/                     globale CSS/JS (app.css, htmx.min.js, ...) → /static/
      templates/
        pages/layout.html         das globale Layout ("layout")
        pages/error.html          Fehlerseite ("error", noch nicht verdrahtet)
        partials/navbar.html      globale Partials ("navbar")
        components/link.html      globale Components ("component.link")
    features/
      home/                       Beispiel-Feature, dient als Kopiervorlage
        assets.go                 //go:embed templates static
        router.go                 Register(): Feature bauen, Seiten holen, Routen mounten
        handler.go                HTTP-Handler
        service.go                Geschäftslogik
        templates/
          pages/home.html         eine Datei = eine Seite
        static/home.css, home.js  → /assets/home/...
```

## Das `ui`-Package

### API

```go
// global
const ui.StaticPrefix = "/static/"
func  ui.StaticHandler() http.Handler                 // globale Static Files

func  ui.NewApp() (*ui.App, error)                    // parst die globalen Templates

// pro Feature
func (a *App)     NewFeature(name, defaultTitle string, fsys fs.FS) (*Feature, error)
func (f *Feature) Page(name string) (*Page, error)    // templates/pages/<name>.html
func (f *Feature) StaticPrefix() string               // "/assets/<name>/"
func (f *Feature) StaticHandler() http.Handler        // serviert <feature>/static/

// pro Seite
func (p *Page) Render(w http.ResponseWriter, r *http.Request, data ui.PageData) error

type PageData struct {
    Title  string   // leer → defaultTitle des Features
    Header Header
    Footer Footer
    Data   any      // feature-spezifische Daten
}
```

### Template-Sets: eines pro Seite

Go-Templates haben kein `extends` wie Jinja. Alle `{{define}}`s eines Sets teilen sich
einen Namensraum. Würden alle Seiten in einem Set liegen, würden sich ihre
`{{define "content"}}` gegenseitig überschreiben.

Deshalb bekommt **jede Seite ein eigenes Set**:

```
Set einer Seite = globale Templates (ui/templates/**)
                + Feature-Templates außerhalb von pages/ (Partials des Features)
                + genau eine Datei aus templates/pages/
```

Technisch: Die globalen Templates werden einmal geparst. Pro Feature wird das Set geklont
(`Clone()`) und um die Feature-Partials ergänzt. Pro Seite wird das noch einmal geklont und
um die Seitendatei ergänzt. Alles passiert beim Start.

### Template-Vertrag

Das globale `layout` ruft diese Blöcke auf:

| Block | Pflicht | Inhalt |
|---|---|---|
| `"content"` | ja | der Seiteninhalt (im `<body>`) |
| `"head"` | nein | zusätzliche Tags im `<head>` (CSS, JS, Meta) |

Die Template-Daten sind immer ein `ui.PageData`. Eigene Daten kommen in `PageData.Data`.

Verfügbare Template-Funktion:

| Funktion | Ergebnis |
|---|---|
| `{{asset "home.css"}}` | `/assets/<feature>/home.css` |

Beispiel (`features/home/templates/pages/home.html`):

```html
{{define "head"}}
<link rel="stylesheet" href="{{asset "home.css"}}">
<script src="{{asset "home.js"}}" defer></script>
{{end}}

{{define "content"}}
<header>{{template "navbar" .}}</header>
<main>
    <h2>Hello World</h2>
</main>
{{end}}
```

Alle Templates werden mit `html/template` geparst. Die Ausgabe wird also automatisch
kontextabhängig escaped (HTML, Attribute, URLs, JS). Kein `text/template` verwenden.

### Static Files

| URL | Quelle | gemountet von |
|---|---|---|
| `/static/...` | `internal/ui/static/` | `main.go` |
| `/assets/<feature>/...` | `internal/features/<feature>/static/` | `router.go` des Features |

Jedes Feature hat seinen eigenen Präfix. Zwei Features können also gleichnamige Dateien
haben, ohne zu kollidieren. URLs zu Feature-Dateien immer mit `{{asset ...}}` erzeugen,
nie den Präfix von Hand schreiben.

### Fehlerverhalten

| Fehler | Wann er auffällt |
|---|---|
| Syntaxfehler in einem Template | beim Start (`NewApp` / `NewFeature`) |
| Seite ohne `{{define "content"}}` | beim Start (`NewFeature`) |
| Seitenname falsch geschrieben | beim Start (`Page(...)` im Router) |
| Fehler bei der Ausführung (z.B. falsches Feld in `.Data`) | beim Request: `Render` gibt einen `error` zurück |

`Render` puffert die Ausgabe. Bei einem Fehler wurde noch **nichts** geschrieben, und der
Handler entscheidet, wie er antwortet. Interne Fehlermeldungen gehen ins Log, nicht in den
Browser.

## Ein neues Feature hinzufügen

Am einfachsten `features/home` kopieren und umbenennen. Schritt für Schritt:

**1. `assets.go`** – ein einziges Embed für alles:

```go
package riders

import "embed"

//go:embed templates static
var files embed.FS
```

(Gibt es kein `static/`, nur `//go:embed templates` schreiben. Ein Embed auf einen
fehlenden Ordner ist ein Compile-Fehler.)

**2. Templates** – eine Datei pro Seite in `templates/pages/`, jede definiert `"content"`.
Partials, die mehrere Seiten des Features nutzen, kommen in einen anderen Ordner unter
`templates/` (z.B. `templates/partials/`).

**3. `router.go`** – Feature bauen, Seiten holen, mounten:

```go
func Register(mux *http.ServeMux, app *ui.App) error {
    view, err := app.NewFeature("riders", "Fahrer", files)
    if err != nil {
        return err
    }
    listPage, err := view.Page("list")
    if err != nil {
        return err
    }

    h := newHandler(listPage, newService())

    mux.Handle("GET "+view.StaticPrefix(), view.StaticHandler())
    mux.HandleFunc("GET /riders", h.List)
    return nil
}
```

Abhängigkeiten (DB, Konfiguration, ...) werden als Parameter an `Register` übergeben,
nicht über globale Variablen.

**4. `handler.go`** – rendern:

```go
func (h *handler) List(w http.ResponseWriter, r *http.Request) {
    data := ui.PageData{Data: riders}
    if err := h.listPage.Render(w, r, data); err != nil {
        slog.ErrorContext(r.Context(), "riders list", "err", err)
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    }
}
```

**5. `main.go`** – eine Zeile:

```go
if err := riders.Register(mux, app); err != nil {
    log.Fatal(err)
}
```

## Konventionen und Regeln

- **`ui` macht kein Routing.** Es liefert `http.Handler` und rendert. Mounten tut immer der
  Aufrufer. So bleibt `ui` unabhängig vom Router (Std-Mux heute, restify später).
- **Explizites Wiring im Feature.** Alles, was ein Feature ins System einhängt (Routen,
  Static Files, Abhängigkeiten), steht in seinem `router.go`. Keine `init()`-Magie, kein
  globaler Zustand.
- **Fehler beim Start statt beim Request.** Seiten werden im Router per `Page(...)` geholt,
  nicht im Handler über einen String-Namen.
- **Ein Feature ist in sich geschlossen.** Templates, Static Files und Logik liegen im
  Feature-Ordner. Was mehrere Features brauchen, gehört nach `ui/templates` (global) bzw.
  später in eigene Components.
- **Template-Namen:** Seiten definieren `content` und optional `head`. Eigene Partials
  bekommen sprechende, eindeutige Namen (z.B. `rider-list`, `rider-detail-panel`).
- **Keine `**` in `//go:embed`-Mustern.** `**` matcht nur **eine** Verzeichnisebene. Ganze
  Ordner einbetten (`//go:embed templates static`), die sind automatisch rekursiv.
- **Konstruktoren unexportiert**, wenn sie unexportierte Typen liefern (`newHandler`,
  `newService`). Exportiert ist pro Feature nur `Register`.

## Bewusste Entscheidungen

| Entscheidung | Warum |
|---|---|
| Template-Set pro Seite | jede Seite darf `content`/`head` definieren, ohne Dispatcher (`if eq .ContentTemplate ...`) |
| `ui` ohne Mux | dasselbe `ui` funktioniert mit `http.ServeMux` und mit restify (Middleware, Logging bleiben beim Router) |
| `Render` gibt `error` zurück | passt direkt zu `web.HandlerFunc` von restify (`return page.Render(...)`) |
| Static-Präfix pro Feature | keine Namenskollisionen zwischen Features, echte Pfade statt Lookup nach Dateiname |
| `html/template` | automatisches Escaping, Schutz vor XSS |
| explizites Wiring | alles im Feature sichtbar, weniger Seiteneffekte, gut für parallele Arbeit mehrerer Entwickler |

## Einordnung: vergleichbar mit Flask

| Flask | hier |
|---|---|
| `Flask(...)` + Routing | `http.ServeMux` (später restify / `web.API`) |
| Blueprint mit `template_folder` / `static_folder` | Feature mit `templates/` / `static/` |
| `url_for("home.static", filename=...)` | `{{asset "..."}}` |
| `render_template(...)` | `page.Render(w, r, data)` |
| `{% extends "layout.html" %}` | eigenes Template-Set pro Seite + globales `layout` |
| `@login_required` / `before_request` | Middleware am (Sub-)Router |

Unterschied zu Flask: keine globalen Kontext-Objekte (`request`, `g`, `current_app`). Alles
wird explizit übergeben, und Template-Fehler fallen beim Start auf statt beim Aufruf.

## Noch nicht umgesetzt

- **htmx-Fragmente:** `Render` rendert immer die ganze Seite (`layout`). Rendering von
  Fragmenten (z.B. nur ein Panel bei `HX-Request`) kommt, sobald htmx im Projekt genutzt wird.
- **Fehlerseite:** `error.html` existiert, ist aber noch nicht verdrahtet. Handler antworten
  bei Render-Fehlern mit einem generischen 500.
- **Middleware / Auth:** gibt es im Toy-Projekt noch nicht. Im Zielsystem (restify) gilt:
  Auth pro Router-**Gruppe**, nicht pro Route.
- **Integration in restify / fgohtmxui:** siehe [`todo.md`](todo.md).


# 02 – Fahrer-Details nachladen

**Branch:** `t02/fahrer-details` → `t02/complete`
**Neu:** `hx-get`, `hx-target`, `hx-swap`, Seite vs. Fragment, `HX-Request`, `hx-push-url`, Lazy Loading

## Worum es geht

Master-Detail: links die Fahrerliste, rechts ein Detail-Panel. Klick auf einen Fahrer lädt
**nur das Panel** neu. Die URL ändert sich trotzdem, und ein Reload zeigt dieselbe Ansicht.

## Abnahme

- [ ] `/riders` zeigt links die Liste, rechts „Wähle einen Fahrer“.
- [ ] Klick auf einen Namen lädt rechts die Details, **ohne** Seitenreload (Netzwerk-Tab: nur ein kleiner Request).
- [ ] Die URL wechselt auf `/riders/{id}`.
- [ ] Reload auf `/riders/3` zeigt die **ganze Seite** mit Liste und Vali Höll rechts.
- [ ] Der ausgewählte Fahrer ist in der Liste markiert, auch nach Reload.
- [ ] Das Detail-Panel zeigt Team, die **Bikes** des Fahrers (aktive/inaktive unterscheiden)
      und einen Abschnitt „Ergebnisse“.
- [ ] Der Abschnitt „Ergebnisse“ wird **separat nachgeladen**, sobald das Panel da ist (`hx-trigger="load"`).
- [ ] Browser-Zurück funktioniert: Fahrer 3 → Fahrer 1 → Zurück zeigt wieder Fahrer 3.

## Zusammenhänge

### Eine URL, zwei Antworten

`/riders/3` wird auf zwei Arten angefragt:

| Wer fragt | Header | Erwartete Antwort |
| --------- | ------ | ----------------- |
| Browser (Reload, Link, neuer Tab) | kein `HX-Request` | ganze Seite: Layout + Liste + Panel |
| htmx (Klick in der Liste) | `HX-Request: true` | nur das Panel |

Der Handler lädt in beiden Fällen dieselben Daten. Nur der **Einstiegspunkt** ins Template-Set
ist ein anderer: `layout` oder `rider-detail-panel`.

Damit das funktioniert, muss das Panel ein **eigenes benanntes Template** sein, das die Seite
auch selbst einbindet:

```
{{define "content"}}
  <aside>{{template "rider-list" .}}</aside>
  <section id="detail">{{template "rider-detail-panel" .Detail}}</section>
{{end}}
```

So gibt es das Panel-HTML genau **einmal**. Die Seite und die htmx-Antwort können nicht
auseinanderlaufen.

### Wer entscheidet: Handler oder Renderer?

Zwei Varianten:

1. **Renderer entscheidet** anhand von `HX-Request` automatisch (dein früherer Renderer in
   `src/internal/htmx` hat das gemacht).
2. **Handler entscheidet** und ruft explizit `Render` oder `RenderPartial` auf.

Variante 1 ist bequem, wird aber falsch, sobald ein htmx-Request **etwas anderes** als
`content` braucht (die Liste, eine Zeile, ein Modal). Variante 2 ist eine Zeile mehr, aber
immer eindeutig. Probier beide aus und entscheide dich bewusst.

### Die drei wichtigsten Attribute

```html
<a href="/riders/3"
   hx-get="/riders/3"
   hx-target="#detail"
   hx-swap="innerHTML"
   hx-push-url="true">Vali Höll</a>
```

- `href` bleibt. Ohne JS (oder bei Strg+Klick) funktioniert der Link weiterhin.
- `hx-target`: wohin. CSS-Selektor. Ohne Angabe: das Element selbst.
- `hx-swap`: wie. `innerHTML` (Default) ersetzt den Inhalt, `outerHTML` das Element selbst.
- `hx-push-url="true"` schreibt die Request-URL in die Adresszeile und die History.

Tipp: Liegen die Attribute auf **jedem** Link, wiederholst du dich. htmx **vererbt**
`hx-target` und `hx-push-url` von Elternelementen. Setz sie einmal auf die `<ul>`.

### History und Caching

Mit `hx-push-url` gibt es zwei Fallen:

- **Zurück-Button:** htmx speichert Schnappschüsse der Seite. Fehlt einer, fragt htmx die URL
  neu an, mit dem Header `HX-History-Restore-Request: true`. Dann **muss** der Server die ganze Seite
  liefern, nicht das Fragment.
- **Browser-Cache:** Dieselbe URL liefert je nach Header unterschiedliches HTML. Setz
  `Vary: HX-Request`, sonst zeigt der Browser beim Zurück manchmal nur das nackte Fragment.

Beides gehört in den Renderer bzw. eine kleine Hilfsfunktion `isFragmentRequest(r)`.

### Lazy Loading

```html
<div hx-get="/riders/3/results" hx-trigger="load">Lade Ergebnisse…</div>
```

`load` feuert, sobald das Element im DOM ist, also auch, wenn es gerade per htmx eingesetzt
wurde. Ohne `hx-target` ersetzt die Antwort den Inhalt des Divs selbst.
Alternative: `revealed` (erst laden, wenn sichtbar).

## Bauplan

1. Fundament neu aufbauen. Renderer bekommt zusätzlich `RenderPartial(key, name, w, data)`.
2. Repository: `Bikes(ctx, riderID)` (JOIN `rider_bikes` + `bikes`) und
   `Results(ctx, riderID)` (JOIN `rider_competition_instances` + `competition_instances` + `competitions`).
3. Templates: `rider-list`, `rider-detail-panel`, `rider-results` als eigene `define`s.
4. Handler `Detail`: Daten laden, dann Seite **oder** Panel rendern.
5. Neuer Endpoint `GET /riders/{id}/results` → immer Fragment.
6. Markierung des aktiven Fahrers: Die Liste braucht die `SelectedID`. Nach einem htmx-Klick
   wird aber nur das Panel neu geladen. Wie kommt die Markierung trotzdem richtig?
   Mögliche Lösungen: eine CSS-Klasse per kleinem JS umsetzen, oder in Projekt 07 per OOB
   die Liste mit aktualisieren. Für jetzt: Wähle eine Lösung und schreib in einem Kommentar, warum.

## Stolperfallen

- htmx-Script mit `defer` laden und den Pfad prüfen (Netzwerk-Tab: 404 auf `htmx.min.js` → nichts passiert, kein Fehler).
- Im Layout steht `<script href=…>`: Ein Script braucht `src`, nicht `href`.
- `hx-target="detail"` statt `"#detail"`: Das ist ein Element-Selektor, kein id-Selektor.
- Antwortet der Handler auf htmx mit der ganzen Seite, steckt ein komplettes Layout im Panel
  („Seite in der Seite“). Das ist der Klassiker.
- `ParseInt` von `PathValue` prüfen, bevor du damit in die DB gehst.

## Selbsttest

1. Woran erkennt der Server einen htmx-Request?
2. Warum muss `/riders/3` ohne htmx die ganze Seite liefern?
3. Was macht `Vary: HX-Request`, und was passiert ohne?
4. Was ist der Unterschied zwischen `hx-swap="innerHTML"` und `outerHTML`? Wann brauchst du welches?
5. Warum gibt es das Panel-Template nur einmal, und wo wird es überall benutzt?

## Wiedererkennen

Im Arbeitsprojekt lädt die Übersicht ihre Tabelle mit `hx-trigger="load"` nach
(`overview-table-skeleton`). Die Index-Seite rendert nur das Gerüst. Gleiches Muster wie
deine Ergebnis-Sektion.


