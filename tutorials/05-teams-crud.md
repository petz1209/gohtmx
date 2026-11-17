# 05 – Teams verwalten: Inline-CRUD

**Branch:** `t05/teams-crud` → `t05/complete`
**Neu:** POST/PUT/DELETE, Click-to-Edit, Validierung mit Fehleranzeige, Statuscodes und htmx-Swap-Verhalten, gepufferter Renderer, Render-Tests

## Worum es geht

Teams anlegen, bearbeiten, löschen, alles **in der Tabelle** ohne eigene Formularseiten:

- Oben eine Zeile „Neues Team“ mit Eingabefeldern.
- Klick auf „Bearbeiten“ verwandelt eine Zeile in Eingabefelder. „Speichern“ / „Abbrechen“
  verwandelt sie zurück.
- „Löschen“ fragt nach und entfernt die Zeile.

## Abnahme

- [ ] Neues Team anlegen: Die Zeile erscheint in der Tabelle, das Formular ist danach leer.
- [ ] Bearbeiten → Speichern ersetzt **nur diese eine Zeile**. Netzwerk-Tab: Antwort ist ein `<tr>`.
- [ ] Abbrechen holt die ursprüngliche Zeile vom Server (kein JS-Gedächtnis).
- [ ] Validierung: Name Pflicht und eindeutig, Land genau 2 Buchstaben (groß), Gründungsjahr 1970–2100.
- [ ] Fehler werden **am Feld** angezeigt, die eingegebenen Werte bleiben stehen.
- [ ] Doppelter Name („YT Mob“) → Fehlermeldung am Feld, kein 500.
- [ ] Löschen fragt: „Team ‚YT Mob‘ löschen? 1 Fahrer verliert sein Team.“ (Anzahl aus der DB.)
- [ ] Nach dem Löschen eines Teams steht Vali Höll ohne Team da (`ON DELETE SET NULL`).
- [ ] Ein Template-Fehler liefert eine saubere 500-Antwort, kein halbes HTML.
- [ ] Es gibt Tests, die jedes Fragment einmal rendern.

## Zusammenhänge

### REST-artige Routen

```
GET    /teams              Seite (Tabelle + Neu-Zeile)
POST   /teams              anlegen        → neue Zeile oder Formular mit Fehlern
GET    /teams/{id}/row     Anzeige-Zeile  (für "Abbrechen")
GET    /teams/{id}/edit    Edit-Zeile
PUT    /teams/{id}         speichern      → Anzeige-Zeile oder Edit-Zeile mit Fehlern
DELETE /teams/{id}         löschen        → leere Antwort
```

`/teams/{id}/row` und `/teams/{id}/edit` liefern **beide** ein `<tr>` mit derselben `id="team-{id}"`.
Mit `hx-target="closest tr" hx-swap="outerHTML"` tauschen sie sich gegenseitig aus.

### Statuscodes und was htmx damit macht

Das ist der wichtigste Abschnitt dieses Projekts.

| Status | htmx 2 Default |
| ------ | -------------- |
| 200 | Antwort wird eingesetzt |
| 204 No Content | **nichts** wird eingesetzt |
| 4xx / 5xx | **nichts** wird eingesetzt, Event `htmx:responseError` |

Folgen:

- **Löschen:** Antwortest du mit 204, bleibt die Zeile stehen. Antworte mit **200 und leerem
  Body** und `hx-swap="outerHTML"`: Die Zeile wird durch „nichts“ ersetzt. (Oder `hx-swap="delete"`.)
- **Validierungsfehler:** Antwortest du mit 422 und dem Formular mit Fehlern, sieht der Benutzer
  **nichts**. Zwei Wege:
  1. Mit **200** antworten. Funktioniert, aber der Status lügt (Logs, Tests, Monitoring).
  2. htmx sagen, dass 422 eingesetzt werden soll, über `responseHandling` in der Config:
     ```html
     <meta name="htmx-config" content='{"responseHandling":[
       {"code":"204","swap":false},
       {"code":"[23]..","swap":true},
       {"code":"422","swap":true},
       {"code":"[45]..","swap":false,"error":true}]}'>
     ```
  Nimm Weg 2 und versteh, warum die Reihenfolge der Einträge zählt.

### Validierung: wo und wie

```go
type TeamForm struct {
    Name, Country, FoundedYear, Website string // so wie eingegeben, als Strings
    Errors map[string]string                     // Feldname → Meldung
}
func (f *TeamForm) Validate() bool
```

- Das Formular hält die **rohen Strings**, damit eine Eingabe wie „20x5“ wieder angezeigt
  werden kann. Erst nach erfolgreicher Validierung entsteht daraus ein `Team`.
- Eindeutigkeit prüft letztlich die DB (`UNIQUE`). Den Fehler
  „UNIQUE constraint failed: teams.name“ übersetzt das Repository in `ErrDuplicateName`,
  der Handler in einen Feldfehler. SQL-Fehlertexte gehen nie an den Benutzer.

### Der Renderer wird erwachsen

Ab jetzt rendert der Renderer **immer in einen Buffer**:

1. Template in `bytes.Buffer` ausführen.
2. Fehler → loggen, Fehler-Template oder generische 500, **return**.
3. Sonst: `Content-Type`, `WriteHeader(status)`, Buffer schreiben.

Eine private Funktion `execute(w, set, name, status, data)`, die `Render`, `RenderPartial` und
`RenderError` alle benutzen. Der Status wird jetzt ein Parameter, weil du 422 brauchst.

### Render-Tests

Ein Test pro Fragment: `httptest.NewRecorder()`, Fragment mit Beispieldaten rendern, prüfen,
dass kein Fehler kommt und ein paar Schlüsselstellen im HTML stehen (`id="team-3"`, `hx-put="/teams/3"`).
Damit fallen Tippfehler in Template-Namen und Felder auf, bevor du klickst.

## Bauplan

1. Fundament, Renderer mit Buffer und Status.
2. Repository: `List`, `Get`, `Create`, `Update`, `Delete`, `RiderCount(teamID)`.
3. `TeamForm` + `Validate` + Tests.
4. Templates: `team-row`, `team-edit-row`, `team-new-row`, `team-table`.
5. Handler Schritt für Schritt: erst Anzeigen, dann Edit/Abbrechen, dann Speichern, dann Neu, zuletzt Löschen.
6. `htmx-config` im Layout.
7. Render-Tests.

## Stolperfallen

- `<form>` darf nicht direkt in `<tr>` stehen (ungültiges HTML, der Browser schiebt es weg).
  Lösung: `hx-include="closest tr"` am Speichern-Button, ganz ohne `<form>`.
  Oder das `form`-Attribut an den Inputs.
- `r.ParseForm()` liest den Body auch bei **PUT**. Bei **DELETE** schickt htmx 2 Werte in die
  **URL**, nicht in den Body.
- Neue Zeile oben einfügen: `hx-target="#team-tbody" hx-swap="afterbegin"`. Danach muss das
  Neu-Formular geleert werden: `hx-on::after-request="if(event.detail.successful) this.reset()"`.
  Ohne `successful` wird auch nach einem Validierungsfehler geleert.
- `hx-confirm` mit Text aus der DB: Der Text steht in einem Attribut, `html/template` escaped ihn.
  Teamnamen mit `'` testen.
- Zwei Zeilen gleichzeitig im Edit-Modus: Was passiert? Ist das ok? (Für dieses Projekt ja.)

## Selbsttest

1. Was macht htmx mit einer 204-Antwort, was mit einer 422-Antwort ohne Config?
2. Warum hält `TeamForm` Strings und nicht `int` für das Gründungsjahr?
3. Warum darf `Abbrechen` nicht einfach den alten Zustand aus dem DOM wiederherstellen?
4. Was passiert ohne Buffer, wenn ein Template in der Mitte fehlschlägt?
5. Warum übersetzt das Repository den SQLite-Fehler, und nicht der Handler?

## Wiedererkennen

Im Arbeitsprojekt: `ui.Render` mit Buffer, `ui.FieldRule` / `ui.FormInput` für Validierung
am Feld, `CrudFailedEvent` für Fehler-Toasts, und `render_test.go` pro Usecase, der jedes
Template mit Beispieldaten rendert. Dort wird bei Fehlern teils `200 + Toast`, teils
`4xx + Toast` geantwortet. Jetzt weißt du, warum das einen Unterschied macht.
