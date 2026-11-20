# TODO

Offene Punkte für die Integration in das Referenzprojekt
(`dataentry_slogans`, fgohtmxui, focusauth). Stand: Review vom 2026-11-20.

---

## 1. Bug: fehlendes `return` in `RequireRoleAndCountryHtmx`

`pkg/focusauth/middleware.go:114`

```go
user, ok := UserFromContext(r.Context())
if !ok {
    http.Error(w, "internal server error", http.StatusInternalServerError)
    // <- return fehlt
}
```

Fehlt der User, läuft der Code weiter und schreibt ein zweites Mal in die Antwort
(`writeForbiddenHtmx`). Die anderen fünf `Require*`-Funktionen haben das `return`.

- [ ] `return` ergänzen (z.B. `return errors.New("focusauth: no user in context")`)

---

## 2. Sicherheitslücke: sloganmatching-Routen ohne Rollen- und Länderprüfung

`internal/usecases/promofocus/dataentry/sloganmatching/http.router.go`

Diese Routen sind **nicht** mit `RequireRoleAndCountryHtmx` gewrappt, und die
Handler prüfen es auch nicht selbst. Der `uiMiddleware` prüft nur den Login.

| Route | Handler |
|---|---|
| `GET /{prospektID}` | `MatchingIndex` (komplette Seite) |
| `GET /{prospektID}/edit` | `EditPageForm` |
| `GET /{prospektID}/aislogan` | `SloganAt` |
| `GET /{prospektID}/slogans/{group}/picker` | `SloganPicker` |
| `GET /themes/{group}/picker` | `ThemePicker` |
| `GET /types/picker` | `TypePicker` |
| `GET /forms/picker` | `FormPicker` |

Ein eingeloggter User ohne Rolle `dataentry` oder für ein fremdes Land kann
diese Routen aufrufen.

- [ ] durch Punkt 4 (Auth pro Gruppe) beheben
- [ ] danach die anderen Usecases (watcher/sloganmatching, sloganmm, admin, ...) auf dasselbe Muster prüfen

---

## 3. `focusauth`: Rollen als variadischer Parameter, Funktionen als Middleware

Heute gibt es sechs Funktionen, die jeweils eine Rolle und den Handler (`next`) nehmen:

```go
RequireRole(role string, next web.HandlerFunc) web.HandlerFunc
RequireRoleAndCountry(role, pathParam string, next web.HandlerFunc) web.HandlerFunc
// + RequireCountry, und je eine ...Htmx-Variante
```

Ziel:

```go
func RequireRole(roles ...string) func(web.HandlerFunc) web.HandlerFunc
func RequireRoleAndCountry(pathParam string, roles ...string) func(web.HandlerFunc) web.HandlerFunc
```

- [ ] **Rollen variadisch** (`roles ...string`): Der User braucht **mindestens eine** davon.
  Damit braucht es keine zweite Funktion wie `RequireAnyRole`.
  - Variadische Parameter müssen am Ende stehen, deshalb kommt `pathParam` nach vorne.
    Die Parameter-Reihenfolge ändert sich also für alle Aufrufer.
  - Keine Rollen übergeben (`RequireRole()`) ist ein Programmierfehler und darf
    **nicht** „alle erlaubt" bedeuten. Beim Erzeugen der Middleware abbrechen (panic beim Start).
- [ ] **Middleware-Form** `func(web.HandlerFunc) web.HandlerFunc`, damit die Funktion in
  `web.NewMiddlewareStack(...)` passt und an einen Subrouter gehängt werden kann
  (Voraussetzung für Punkt 4).
- [ ] **Htmx- und JSON-Variante zusammenführen?** Prüfen, ob die Antwort bei
  Ablehnung am Request erkannt werden kann (z.B. `HX-Request`-Header),
  statt je zwei Funktionen zu pflegen.
- [ ] `user.HasRole` gegebenenfalls um eine Variante für „eine von mehreren Rollen" ergänzen.

---

## 4. Auth pro Router-Gruppe statt pro Route

Regel: **Explizit überall dort, wo ein Fehler sofort auffällt. Strukturell
erzwungen dort, wo ein Fehler still bleibt.** Eine vergessene Auth-Prüfung
fällt nicht auf, deshalb gehört sie an den Subrouter und nicht an jede Route.

```go
func UIRouter(parent *web.API, conns platform.Conns) {
    root := parent.NewRouter("/sloganmatching", nil) // selbst ohne Routen!

    pages := root.NewRouter("/...", web.NewMiddlewareStack(
        focusauth.RequireRoleAndCountry("locale", "dataentry")))
    pages.Get("/{prospektID}", h.MatchingIndex)
    // ...

    // bewusst für mehrere Rollen freigegeben
    pickers := root.NewRouter("/pickers", web.NewMiddlewareStack(
        focusauth.RequireRoleAndCountry("locale", "dataentry", "watcher")))
    pickers.Get("/themes/{group}", h.ThemePicker)
}
```

- [ ] Der Root-Router eines Features bekommt **keine eigenen Routen**. Jede Route liegt in
  genau einer Gruppe mit klarer Rollenregel.
- [ ] Geteilte Partials/Picker sind eine eigene, sichtbare Gruppe (bewusste Ausnahme, im Review erkennbar).
- [ ] Achtung: `web.NewRouter("")` beendet das Programm per `log.Fatal` (leerer Präfix nicht erlaubt).
  Jede Gruppe braucht deshalb einen echten Präfix, oder `web` wird angepasst.
- [ ] Routenänderungen (z.B. `/themes/...` → `/pickers/themes/...`) in den Templates nachziehen (`hx-get`).
- [ ] Liefern Picker länderbezogene Daten, brauchen sie weiterhin die **Länderprüfung**, nicht nur die Rolle.
- [ ] Picker, die wirklich von mehreren Features genutzt werden, nach `internal/components/`
  verschieben (wie `creativepreview`), mit eigenen Templates, Static Files, Routen und Rollenregel.

---

## 5. URL-Struktur: UI auf `/`, `/api` und `/auth` als Abzweig

Heute liegt die ganze UI unter `/ui/...`. Ziel: Die UI ist das eigentliche Produkt
und bekommt saubere URLs (`/home`, `/dataentry/...`). Sonderfälle bekommen einen Präfix.

```
srv.API()                       öffentlich
 ├─ /auth/login, /auth/logout   focusauth (BasePath "/auth")
 ├─ /static/, /assets/<feature>/
 ├─ /api/...                    eigene Middleware (API-Auth)
 └─ Group(authStack)            privat, OHNE Präfix
     ├─ /home
     └─ /dataentry/country/{locale}/...
```

**Voraussetzung:**

- [ ] `web`: Router-**Gruppe ohne Präfix** (z.B. `api.Group(stack)`), die nur Middleware bündelt.
  Heute beendet `web.NewRouter("")` das Programm per `log.Fatal`. Ohne diese Gruppe müsste
  Auth an jedes Feature einzeln, und „privat" wäre keine Struktur mehr.

**Umsetzung:**

- [ ] Öffentliches und Privates über **zwei Router-Bäume** trennen, **keine** Pfad-Ausnahmen
  in der Auth-Middleware (`if path == "/login"` veraltet und öffnet still Lücken).
- [ ] `focusauth.BasePath` auf `"/auth"` umstellen → `/auth/login`, `/auth/logout`, ...
  Prüfen, wie focusauth die Pfade zusammensetzt (kein `/auth/auth`).
- [ ] Logout bleibt `POST` (kein versehentliches Auslösen per Link oder Prefetch).
- [ ] `LandingPath` (heute `home.Href` = `/ui/home`) auf `/home` bzw. `/` anpassen.
- [ ] **Reservierte Pfade** dokumentieren: `/auth`, `/api`, `/static`, `/assets`, `/docs`,
  `/openapi.json`. Kein Feature darf so heißen bzw. dort Routen haben.
- [ ] Root-Route mit `GET /{$}` statt `GET /` (`/` ist im Go-Mux ein Catch-all), dadurch fällt die
  manuelle Prüfung `if r.URL.Path != "/"` in `main.go` weg.
- [ ] **Cookie-Pfade** bewusst entscheiden: Session-Cookie mit `Path=/` gilt auch für `/api`.
  Gewollt? Feature-Cookies (z.B. `sloganmatching_filter`) auf neue Pfade umstellen.
- [ ] Alle `/ui`-Vorkommen nachziehen: `Href`-Konstanten, `PageData.BaseHref` / `<base>`,
  `Config.LogoutURL`, `Config.CountrySwitchURL`, `hx-get`-URLs in Templates, Tests.
- [ ] Static Files bleiben öffentlich (Login-Seite braucht CSS). Daher nichts Vertrauliches in `static/`.
- [ ] Reihenfolge in restify beachten: Ein Kind-Router übernimmt die Middleware des Parents **bei
  der Erstellung**, `Use()` wirkt nur auf danach registrierte Routen. Middleware direkt in
  `NewRouter(prefix, stack)` / `Group(stack)` mitgeben.

**Zeitpunkt:** Für neue Apps von Anfang an so. In `dataentry_slogans` erst **nach** den Punkten 3, 4 und 6
(eigenes Refactoring, `/ui` steckt an vielen Stellen).

---

## 6. fgohtmxui: zentraler Renderer (Muster aus dem Toy-Projekt `gohtmx`)

Umsetzung im fgohtmxui-Repo, nicht in `vendor/`. Usecase für Usecase migrieren,
parallel zum bestehenden `ui.Render`. sloganmatching zuerst, `render_test.go` als Absicherung.

- [ ] `ui.New(cfg)` statt `ui.Configure`: kein globaler Zustand, keine Abhängigkeit von der Init-Reihenfolge
- [ ] Template-Set **pro Seite** (`templates/pages/*.html`), dadurch fällt der `content.html`-Dispatcher weg
- [ ] Seiten beim Start holen (`f.Page("index")`): Fehler beim Start statt `template.Must` in einer Package-Variable
- [ ] `Page.Render(w, r, data) error`: passt direkt zu `web.HandlerFunc`
- [ ] `PageData`-Felder, die aus dem Request kommen (Lang, Nav, ...), zentral im Renderer füllen
- [ ] `ui` macht kein Routing: Static-Handler zurückgeben, der Aufrufer mountet
- [ ] HTML-Fehler-Middleware (analog `web.AutoRespondErrors`) für den `/ui`-Router.
  Handler machen nur `return err`, statt `http.Error` / `web.JsonError` gemischt zu verwenden
- [ ] `web`: `FileServerFS(path, fs.FS)`, damit Static-Mounts nicht an `Mux()` vorbei laufen
