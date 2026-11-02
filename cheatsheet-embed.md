# Cheatsheet: `embed` für Go-Webapps

Dateien (Templates, CSS, JS, SQL …) werden **zur Compile-Zeit** ins Binary gepackt.
Ein Binary, keine Dateien neben der App ausliefern.

---

## Grundform

```go
import "embed"

//go:embed templates/*.html
var templates embed.FS
```

- Die Direktive steht **direkt über** einer Variable auf **Paketebene** (nicht in einer Funktion).
- Zwischen `//` und `go:embed` steht **kein Leerzeichen**.
- Erlaubte Typen:

| Typ | Wofür | Import |
|---|---|---|
| `embed.FS` | mehrere Dateien / Ordner (Templates, Static) | `"embed"` |
| `string` | eine einzelne Textdatei (z. B. SQL) | `_ "embed"` |
| `[]byte` | eine einzelne Binärdatei | `_ "embed"` |

```go
import _ "embed"

//go:embed sql/ddl.sql
var ddl string
```

---

## Pfade: die wichtigsten Regeln

| Regel | Beispiel |
|---|---|
| Pfade sind **relativ zum Ordner des Pakets** (der `.go`-Datei) | `templates/home.html` |
| **Kein `..`**, keine absoluten Pfade | `../ui/x.html` ❌ |
| Immer **`/`**, auch unter Windows | `templates\home.html` ❌ |
| Das Pattern muss **mindestens eine Datei** treffen, sonst Compile-Fehler | |
| `**` (rekursiver Glob) gibt es **nicht** | `templates/**/*.html` ❌ |
| Ein Ordnername bettet **rekursiv** alles ein | `//go:embed static` |
| Bei einem Ordnernamen werden Dateien mit `.` oder `_` am Anfang **ausgelassen** | `static/.DS_Store` |
| Mit dem Präfix `all:` werden sie mitgenommen | `//go:embed all:static` |
| Mehrere Patterns sind erlaubt | `//go:embed templates/*.html partials/*.html` |

### Der Pfad im FS behält den Ordner!

```go
//go:embed templates/*.html
var templates embed.FS

templates.Open("home.html")           // ❌ file does not exist
templates.Open("templates/home.html") // ✅
```

Die Variable ist die **Wurzel**, und `templates/` ist ein **Unterordner darin**.

Wenn du den Präfix loswerden willst, nimm `fs.Sub`:

```go
sub, err := fs.Sub(templates, "templates")
sub.Open("home.html") // ✅
```

---

## Webapp-Muster

### Templates pro Feature

`embed` kann nicht nach `..` greifen. Darum bettet **jedes Paket seine eigenen Dateien ein** und reicht sie als `fs.FS` weiter:

```
ui/assets.go                 //go:embed templates/*  → globales Layout
internal/feature/home/x.go   //go:embed templates/*.html → Seiten
```

```go
// ui/assets.go
//go:embed templates/*.html
var Templates embed.FS

// main.go
renderer := htmx.NewRenderer(ui.Templates)
```

Funktionen sollten **`fs.FS`** annehmen, nicht `embed.FS`:

```go
func NewRenderer(fsys fs.FS) *Renderer { ... }
```

Dann funktioniert dieselbe Funktion auch mit `os.DirFS(...)` (Live-Reload) oder mit `fstest.MapFS` (Tests).

### Statische Dateien ausliefern

```go
//go:embed static
var staticFS embed.FS

// URL /static/css/app.css → Datei static/css/app.css
// Das passt direkt, weil der Präfix "static/" im FS erhalten bleibt.
mux.Handle("GET /static/", http.FileServerFS(staticFS))
```

Wenn die URL anders heißen soll als der Ordner:

```go
sub, _ := fs.Sub(staticFS, "static")
mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(sub)))
```

`http.FileServerFS` gibt es seit Go 1.22. Davor schrieb man `http.FileServer(http.FS(staticFS))`.

### Dev vs. Prod: ohne Rebuild entwickeln

Eingebettete Dateien ändern sich **erst nach einem neuen Build**. Für die Entwicklung kannst du umschalten:

```go
var uiFS fs.FS = ui.Templates            // Prod: eingebettet
if *dev {
    uiFS = os.DirFS("ui")                // Dev: direkt von der Platte
}
```

Wichtig: Beide FS müssen dieselbe Struktur haben. Hier liegt `templates/…` in beiden Fällen unter der Wurzel.

---

## Debugging: Was ist eigentlich drin?

```go
fs.WalkDir(templates, ".", func(path string, d fs.DirEntry, err error) error {
    fmt.Println(path)
    return nil
})
```

Damit siehst du sofort, unter welchem Pfad eine Datei wirklich liegt.

---

## Typische Fehler

| Fehler | Ursache |
|---|---|
| `open home.html: file does not exist` | Ordnerpräfix vergessen → `templates/home.html` |
| `pattern ...: no matching files found` | Pattern trifft nichts (Tippfehler, falscher Ordner) |
| `invalid pattern syntax` / `..` | Pfad außerhalb des Paketordners |
| `go:embed only allowed in Go files that import "embed"` | Import fehlt (bei `string`/`[]byte`: `_ "embed"`) |
| Änderung am HTML ist nicht zu sehen | Kein Rebuild → Dev-Modus mit `os.DirFS` nutzen |
| `.htaccess` / `_partial.html` fehlt | Durch Ordner-Pattern ausgelassen → `all:` oder explizites `*` |
