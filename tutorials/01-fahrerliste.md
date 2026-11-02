# 01 – Fahrerliste

**Branch:** `t01/fahrerliste` → `t01/complete`
**Neu:** Request-Zyklus ohne htmx, `embed`, Template-Sets, Layout/Content/Title, Repository, ViewModel, Fehlerseiten

Noch **kein htmx**. Das klingt langweilig, ist aber die Basis: Alles, was htmx später macht, ist
„nur“ ein weiterer Request auf genau diese Mechanik.

## Worum es geht

Eine Übersicht aller Fahrer und eine Detailseite pro Fahrer, als ganz normale Webseiten mit Links.

## Abnahme

- [ ] `GET /` leitet auf `/riders` weiter (oder zeigt die Liste direkt).
- [ ] `GET /riders` zeigt eine Tabelle: Name, Spitzname (falls vorhanden), Nation, Disziplin, Team, Alter.
- [ ] Max Muster hat kein Team. Die Zelle zeigt „–“ statt eines Fehlers oder `<nil>`.
- [ ] Jeder Name ist ein Link auf `/riders/{id}`.
- [ ] `GET /riders/{id}` zeigt eine Detailseite: alle Felder plus Team-Name und Team-Website als Link.
- [ ] `GET /riders/999` liefert **404** mit einer eigenen Fehlerseite im Layout.
- [ ] `GET /riders/abc` liefert **400**.
- [ ] Der `<title>` jeder Seite ist sinnvoll („Fahrer“, „Loïc Bruni“).
- [ ] „Loïc“ und „Höll“ werden korrekt angezeigt (UTF-8).

## Zusammenhänge

### Der Request-Zyklus ohne htmx

```
Browser: GET /riders/2
  → ServeMux findet "GET /riders/{id}"
  → Handler liest r.PathValue("id"), wandelt in int um (Fehler → 400)
  → Repository: SELECT … WHERE r.id = ?  (keine Zeile → 404)
  → Handler baut ViewModel
  → Renderer: Template-Set "rider" ausführen, Startpunkt "layout"
  → Browser bekommt komplettes HTML
```

### Template-Sets und Namen

Ein `*template.Template` ist ein **Set** benannter Templates. `{{define "x"}}` legt ein Template
namens `x` an. `{{template "x" .}}` ruft es auf, **im selben Set**.

Das Layout ruft `{{template "content" .}}` auf. Jede Seite definiert ihr eigenes `content`.
Zwei Seiten im selben Set würden sich also gegenseitig überschreiben. Deshalb:

```
Base-Set (ui):   layout, title, error
   │ Clone()
   ├── Set "rider-list":   + content (Liste)
   └── Set "rider-detail": + content (Detail)
```

Überleg dir: Registrierst du pro **Feature** ein Set oder pro **Seite**? Hier hat das Feature
„rider“ zwei Seiten. Beide Wege gehen. Wichtig ist nur, dass jedes Set genau ein `content` hat.

### DB-Modell vs. ViewModel

Die Tabelle `riders` hat `nickname`, `team_id`, `birth_date`, die alle `NULL` sein können.
In Go landen sie in `sql.NullString` / `sql.NullInt64` (oder Pointern). Das Template soll sich
damit nicht herumschlagen.

- **Repository** liefert, was die DB hat (mit Null-Typen).
- **Handler** (oder eine kleine Funktion daneben) baut daraus ein **ViewModel** mit fertigen
  Strings: `TeamName` ist `""` statt `NullString`, `Age` ist bereits berechnet.
- **Template** zeigt nur noch an und entscheidet höchstens `{{with}}` / `{{else}}`.

Faustregel: Logik, die du testen willst, gehört nach Go, nicht ins Template.

## Bauplan

1. **Fundament** (siehe README): DB, `ui`-Layout, Renderer mit `Register` und `Render`.
2. **Repository** `internal/feature/rider/repository.go`:
   - `List(ctx) ([]Rider, error)` mit `LEFT JOIN teams`. Warum `LEFT` und nicht `INNER`?
     (Antwort: Max Muster.)
   - `Get(ctx, id) (Rider, error)`. Für „nicht gefunden“ einen eigenen Fehler zurückgeben
     (`ErrNotFound`), den der Handler mit `errors.Is` erkennt.
   - `rows.Close()` und `rows.Err()` nicht vergessen.
3. **ViewModel**: `RiderRow` (für die Liste) und `RiderDetail`. Alter aus `birth_date` berechnen
   (`time.Parse("2006-01-02", …)`).
4. **Templates**: `list.html` und `detail.html`, beide mit `{{define "content"}}`.
5. **Handler**: `List`, `Detail`. Fehlerfälle → `Render` mit Fehler-Template und richtigem Status.
6. **Routen** in `Connect(mux)`:
   ```go
   mux.HandleFunc("GET /riders", h.List)
   mux.HandleFunc("GET /riders/{id}", h.Detail)
   ```
7. **Fehlerseite** im Base-Set: `{{define "error"}}`. Der Renderer braucht eine Methode, die mit
   Status rendert, etwa `RenderError(w, status, msg)`.

## Stolperfallen

- `rows.Scan` braucht **einen Pointer pro Spalte**, in derselben Reihenfolge wie im `SELECT`.
  `rows.Scan(&rider)` funktioniert nicht.
- `SELECT name FROM riders`: Es gibt keine Spalte `name`, nur `first_name`/`last_name`.
- `{{with .X}}` setzt den Punkt auf `.X`. Innerhalb davon `{{.}}` statt `{{.X}}` (kennst du schon vom Title).
- `w.WriteHeader(404)` **vor** dem ersten `Write`, sonst ist es zu spät. Das ist ein Grund für
  den Buffer im Renderer (Projekt 05).
- `GET /` als Pattern matcht **alles**. Für „nur die Startseite“: `GET /{$}`.

## Selbsttest

1. Was passiert, wenn zwei Templates im selben Set beide `{{define "content"}}` haben?
2. Warum klonst du das Base-Set und parst nicht einfach alles in ein großes Set?
3. Wo wird aus `sql.NullString` ein anzeigbarer String, und warum genau dort?
4. Wann muss `Clone()` passieren, und warum nicht erst beim Request?
5. Was ist der Unterschied zwischen `template.ParseFS(...)` und `t.ParseFS(...)`?

## Wiedererkennen

Im Arbeitsprojekt heißt `Register` etwas anders: Jedes Usecase-Paket baut sich sein `pageTmpl`
aus `ui.Base()` plus Komponenten plus eigenen Templates. Dort gibt es **ein** Set pro Paket.
Deshalb gibt es dort den `content.html`-Dispatcher, wenn ein Paket zwei Seiten hat. Genau
das Problem aus der Frage 1.

## Erweiterung (optional)

- Nationalität als Flaggen-Emoji über eine eigene Template-Funktion (`FuncMap`). Achtung:
  `Funcs` muss **vor** dem Parsen aufgerufen werden.
