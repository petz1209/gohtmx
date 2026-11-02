# Cheatsheet: `html/template` für Go-Webapps

---

## Immer `html/template`, nie `text/template`

`html/template` **escaped automatisch** je nach Kontext (HTML, Attribut, URL, JS).
Das schützt vor XSS. Die API ist identisch mit `text/template`.

```go
import "html/template"
```

---

## Parsen

```go
t, err := template.ParseFS(fsys, "templates/*.html")    // aus embed.FS / fs.FS
t, err := template.ParseFiles("a.html", "b.html")       // von der Platte
t := template.Must(template.ParseFS(...))               // panic bei Fehler, gut beim Start

t, err = t.ParseFS(fsys, "x.html")                      // in ein BESTEHENDES Set dazuparsen
```

> `template.ParseFS(...)` (Paketfunktion) erzeugt ein **neues** Set.
> `t.ParseFS(...)` (Methode) parst in **dieses** Set dazu.

### Template-Namen

Ein Set enthält **mehrere benannte Templates**:

- Jede Datei wird unter ihrem **Dateinamen** abgelegt (`home.html`, ohne Ordner).
- Jedes `{{ define "name" }}` wird ein eigenes Template.

```go
t.ExecuteTemplate(w, "layout", data)   // Name eines Templates IM Set
```

> Der Name in `ExecuteTemplate` ist **kein** Map-Key und **kein** Pfad.
> Fehler: `html/template: "home" is undefined`. Prüfen kannst du das mit `t.DefinedTemplates()`.

---

## Syntax-Grundlagen

| Syntax | Bedeutung |
|---|---|
| `{{ . }}` | der aktuelle Wert („dot“) |
| `{{ .Title }}` | Feld eines Structs / Key einer Map |
| `{{ .User.Name }}` | verschachtelt |
| `{{ .Method }}` | Methode ohne Argumente |
| `{{ $ }}` | die **Root-Daten**, auch innerhalb von `range`/`with` |
| `{{ $x := .Title }}` | Variable anlegen |
| `{{ $x = "neu" }}` | Variable neu zuweisen |
| `{{/* Kommentar */}}` | Kommentar |
| `{{- . -}}` | Whitespace links/rechts entfernen |
| `{{ headline }}` | ❌ ist ein **Funktionsaufruf**, kein Feld! |

### Leer / „falsy“

`false`, `0`, `nil`, `""`, leeres Slice, leere Map → gelten in `if`/`with`/`range` als **leer**.

---

## Kontrollstrukturen

```html
{{ if .LoggedIn }}
  Hallo {{ .User.Name }}
{{ else if .Guest }}
  Hallo Gast
{{ else }}
  Bitte einloggen
{{ end }}
```

```html
{{ with .User }}            {{/* setzt . = .User, nur wenn nicht leer */}}
  {{ .Name }}
{{ else }}
  Kein User
{{ end }}

{{ with .A }}…{{ else with .B }}…{{ else }}…{{ end }}   {{/* Kette, seit Go 1.23 */}}
```

```html
<ul>
{{ range .Riders }}         {{/* . = aktuelles Element */}}
  <li>{{ .Name }} – {{ $.Title }}</li>     {{/* $ = Root */}}
{{ else }}
  <li>Keine Fahrer</li>     {{/* wenn leer */}}
{{ end }}
</ul>

{{ range $i, $r := .Riders }}{{ $i }}: {{ $r.Name }}{{ end }}
{{ range .Items }}{{ if .Hidden }}{{ continue }}{{ end }}…{{ end }}   {{/* break/continue */}}
```

### Default-Wert, wenn Daten fehlen können

```html
{{ with .Data }}
  {{ with .Headline }}<h1>{{ . }}</h1>{{ else }}<h1>Default</h1>{{ end }}
{{ else }}
  <h1>Default</h1>
{{ end }}
```

> `{{ or .Data.Headline "Default" }}` **knallt**, wenn `.Data` `nil` ist:
> `nil pointer evaluating interface {}.Headline`

---

## Eingebaute Funktionen

| Funktion | Beispiel |
|---|---|
| `eq ne lt le gt ge` | `{{ if eq .Status "done" }}` |
| `eq` mit mehreren Werten | `{{ if eq .Role "admin" "owner" }}` (= oder) |
| `and or not` | `{{ if and .A (not .B) }}` |
| `len` | `{{ len .Items }}` |
| `index` | `{{ index .Map "key" }}`, `{{ index .Slice 0 }}` |
| `printf` | `{{ printf "%.2f €" .Price }}` |
| `slice` | `{{ slice .Name 0 3 }}` |
| `urlquery` | `{{ urlquery .Search }}` |
| Pipes | `{{ .Price \| printf "%.2f" }}` |

### Eigene Funktionen

```go
t := template.New("").Funcs(template.FuncMap{
    "upper": strings.ToUpper,
    "date":  func(t time.Time) string { return t.Format("02.01.2006") },
})
t = template.Must(t.ParseFS(fsys, "templates/*.html"))   // Funcs VOR dem Parsen!
```

```html
{{ .Name | upper }}  {{ date .CreatedAt }}
```

---

## Layout-Muster für Webapps

### Bausteine

| Syntax | Bedeutung |
|---|---|
| `{{ define "name" }}…{{ end }}` | benanntes Template definieren |
| `{{ template "name" . }}` | Template einbinden (Name muss **konstant** sein!) |
| `{{ block "name" . }}Default{{ end }}` | = define + template: Platzhalter mit Default-Inhalt |

### Layout (global)

```html
{{ define "layout" }}
<!DOCTYPE html>
<html lang="de">
<head>
  <title>{{ template "title" .Title }}</title>
  <link rel="stylesheet" href="/static/css/app.css">
  <script src="/static/js/app.js" defer></script>
</head>
<body>
  <main>{{ template "content" . }}</main>
</body>
</html>
{{ end }}
```

### Seite (nur Inhalt)

```html
{{ define "content" }}
  <h1>{{ .Data.Headline }}</h1>
{{ end }}
```

> Die Seite „benutzt“ das Layout nicht. Das **Layout** ruft den Block der Seite auf.
> Go hat kein `extends`. `define` + `template` ist das Gegenstück dazu.

### Problem: Jede Seite definiert `content`

In **einem** Set darf `content` nur einmal vorkommen, sonst überschreibt die letzte Datei alle anderen.
Lösung: **ein Set pro Seite**, per `Clone()`:

```go
base := template.Must(template.ParseFS(uiFS, "templates/*.html")) // layout, title, …
pages := map[string]*template.Template{}

func register(key string, fsys fs.FS, patterns ...string) error {
    t, err := base.Clone()                 // eigene Kopie von layout + title
    if err != nil { return err }
    t, err = t.ParseFS(fsys, patterns...)  // Methode! Seite dazuparsen
    if err != nil { return err }
    pages[key] = t
    return nil
}

register("home",        homeFS,  "templates/home.html")
register("rider-list",  riderFS, "templates/list.html", "templates/partials/*.html")
```

```go
t := pages["home"]                        // 1. Set holen (Map-Key)
t.ExecuteTemplate(w, "layout", page)      // 2. Template im Set ausführen (Template-Name)
t.ExecuteTemplate(w, "content", page)     //    nur das Fragment (z. B. für HTMX)
```

> ⚠️ **Nach dem ersten `Execute` kann man ein Set nicht mehr klonen.**
> Das Base-Set also **nur klonen, nie direkt ausführen**.

### Immer dieselbe Datenform übergeben

Gib Layout **und** Fragment dieselbe Struktur (z. B. `Page{Title, Data}`).
Sonst braucht ein Block einmal `.Headline` und einmal `.Data.Headline`.

### Mehrere Werte an ein Sub-Template

`{{ template "x" }}` nimmt **genau ein** Argument. Mehrere Werte gehen nur über ein Struct oder eine Map
(bzw. eine selbst geschriebene `dict`-Funktion über `Funcs`).

---

## Rendern im Handler

Rendere zuerst **in einen Buffer**. So bekommt der Client bei einem Fehler keine halbe Seite plus Fehlertext:

```go
var buf bytes.Buffer
if err := t.ExecuteTemplate(&buf, "layout", page); err != nil {
    http.Error(w, "render error", http.StatusInternalServerError)
    log.Println(err)
    return
}
w.Header().Set("Content-Type", "text/html; charset=utf-8")
buf.WriteTo(w)
```

Templates **beim Start** parsen, nicht pro Request. Fertig geparste Sets sind für parallele Requests sicher.

---

## Escaping

| Kontext | Was passiert |
|---|---|
| `<p>{{ . }}</p>` | `<` → `&lt;` usw. |
| `<a href="{{ . }}">` | gefährliche URLs (`javascript:…`) werden zu `#ZgotmplZ` |
| `<script>var x = {{ . }}</script>` | wird als JS-Wert/JSON escaped |
| `template.HTML(s)` | **kein** Escaping. Nur für Inhalte, denen du vertraust! |

> Wenn `ZgotmplZ` im Output auftaucht, hat `html/template` einen unsicheren Wert blockiert.

---

## Typische Fehler

| Fehlermeldung | Ursache |
|---|---|
| `unexpected EOF` | `{{ end }}` fehlt (oft beim `define`) |
| `function "x" not defined` | `{{ x }}` statt `{{ .X }}`, oder `Funcs` erst nach dem Parsen registriert |
| `no such template "content"` | Block ist nicht im **ausgeführten** Set (anderes Set, nicht geparst) |
| `"home" is undefined` | Map-Key statt Template-Name an `ExecuteTemplate` übergeben |
| `can't evaluate field X in type Y` | Feld existiert nicht. Oft wurde `Page` statt `Page.Data` (oder umgekehrt) übergeben |
| `nil pointer evaluating interface {}.X` | Zugriff auf ein Feld von `nil`. Mit `with` absichern |
| `cannot Clone … after it has executed` | Base-Set wurde schon ausgeführt |
| Map-Key fehlt → leere Ausgabe | Gewollt. Strenger geht es mit `t.Option("missingkey=error")` |
| Titel zeigt `{Home home-page <nil>}` | ganzes Struct übergeben statt `.Title` |
