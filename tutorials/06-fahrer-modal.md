# 06 – Fahrer im Modal anlegen & bearbeiten

**Branch:** `t06/fahrer-modal` → `t06/complete`
**Neu:** Modal-Container im Layout, Formulare im Modal, `HX-Trigger`-Response-Header, Events mit `from:body`, Entkopplung von Auslöser und Empfänger

## Worum es geht

Die Fahrertabelle (mit Suche aus 03, gern vereinfacht) bekommt:

- einen Button „Neuer Fahrer“, der ein **Modal** mit Formular öffnet,
- pro Zeile „Bearbeiten“, das dasselbe Modal mit vorbelegten Werten öffnet.

Nach dem Speichern schließt sich das Modal, und die Tabelle aktualisiert sich. Das Modal
**weiß aber nicht**, dass es eine Tabelle gibt. Das ist der Kern dieses Projekts.

## Abnahme

- [ ] „Neuer Fahrer“ öffnet ein Modal mit Formular: Vorname, Nachname, Spitzname, Nation,
      Geburtsdatum, Disziplin (Select), Team (Select aus DB, „kein Team“ möglich).
- [ ] Escape, Klick auf den Hintergrund und „Abbrechen“ schließen das Modal.
- [ ] Validierungsfehler werden **im Modal** angezeigt, das Modal bleibt offen.
- [ ] Erfolg: Modal schließt sich, Tabelle lädt neu, die aktive Suche bleibt erhalten.
- [ ] „Bearbeiten“ öffnet dasselbe Formular, vorbelegt. Titel „Fahrer bearbeiten“.
- [ ] Das Modal-Template enthält **keinen** Hinweis auf die Tabelle (kein `hx-target="#rider-table"`).
- [ ] Ein Fahrer mit Akzent im Namen („Loïc“) erscheint nach dem Speichern korrekt im Toast
      „Loïc Bruni gespeichert“.
- [ ] Die Seite `/riders/new` direkt aufgerufen zeigt das Formular als normale Seite (Bonus).

## Zusammenhänge

### Der Modal-Container

Im Layout gibt es **einen** leeren Platz:

```html
<div id="modal"></div>
```

„Neuer Fahrer“ lädt das Modal-HTML dort hinein:

```html
<button hx-get="/riders/new" hx-target="#modal">Neuer Fahrer</button>
```

Schließen heißt: `#modal` leeren. Das kann der Server (leere Antwort mit Target `#modal`)
oder ein kleines Stück JS auf ein Event hin.

Für das Modal selbst bietet sich `<dialog>` an: `showModal()` bringt Backdrop, Escape und
Fokus-Falle mit. Es muss nach dem Einsetzen per JS geöffnet werden, zum Beispiel
`hx-on::after-swap` am Container oder ein `htmx.onLoad`-Handler.

### Das Problem: Wer aktualisiert die Tabelle?

Nach dem Speichern müssen **zwei** Dinge passieren: Modal zu, Tabelle neu.
Die Antwort auf `POST /riders` geht aber nur in **ein** Target.

Naive Lösung: Das Formular hat `hx-target="#rider-table"`, und der Server antwortet mit der
Tabelle. Dann ist das Modal fest an diese Tabelle gekoppelt. Öffnest du es von einer
anderen Seite (zum Beispiel Team-Detail), bricht es.

**Lösung mit Events:**

```
POST /riders  (aus dem Modal)
  Antwort:  200, Body leer
            HX-Trigger: {"rider-saved": {"id": 7, "label": "Loïc Bruni"}, "modal-close": true}

htmx feuert die Events am auslösenden Element. Sie blubbern bis <body>.

Tabelle:   <div id="rider-table" hx-get="/riders/results" hx-include="#filter"
                hx-trigger="rider-saved from:body">
Modal:     JS-Listener auf "modal-close" → #modal leeren / dialog.close()
Toast:     JS-Listener auf "rider-saved" → Text aus event.detail.label anzeigen
```

Der Sender sagt nur, **was passiert ist**. Wer darauf reagiert, entscheiden die Empfänger.
Gibt es auf einer Seite keine Fahrertabelle, hört einfach niemand zu.

### Die drei Zeitpunkte

| Header | feuert |
| ------ | ------ |
| `HX-Trigger` | sofort nach Empfang, **vor** dem Swap |
| `HX-Trigger-After-Swap` | nach dem Einsetzen |
| `HX-Trigger-After-Settle` | nach dem Settle (Attribute, CSS-Übergänge) |

Wenn ein Empfänger das **neue** HTML braucht, nimm `After-Swap`. Für „Tabelle neu laden“ ist
`HX-Trigger` richtig.

### Event-Namen sind ein Vertrag

Behandle Event-Namen wie eine API: Leg sie als Konstanten in Go fest, mit Hilfsfunktionen,
zum Beispiel `events.RiderSaved(w, id, label)`. Im Template steht derselbe Name als String.
Schreib an die Konstante, wer das Event sendet und wer zuhört.

## Bauplan

1. Fundament + Fahrertabelle mit Filter (Kurzform von 03).
2. Modal-Container + `<dialog>`-Template als **Komponente im Base-Set**: `{{define "modal"}}`
   mit Titel und Inhalt. Wie gibst du einem Template zwei Werte mit (Titel + Formular-Daten)?
   Ein kleines Struct, nicht `dict`-Magie.
3. `RiderForm` + `Validate` (wie `TeamForm` in 05).
4. Routen: `GET /riders/new`, `POST /riders`, `GET /riders/{id}/edit`, `PUT /riders/{id}`.
5. Paket `events` mit `SetTrigger(w, map[string]any)`: JSON marshallen, **Nicht-ASCII escapen**
   (siehe Stolperfallen), Header setzen.
6. Tabelle hört auf `rider-saved from:body`.
7. `app.js`: Listener für `modal-close` und `rider-saved` (Toast).

## Stolperfallen

- **Umlaute/Akzente im Header:** HTTP-Header sind Bytes. Der Browser liest sie als Latin-1.
  Aus „Loïc“ wird „LoÃ¯c“. Lösung: Im JSON alle Nicht-ASCII-Zeichen als `\uXXXX` schreiben.
  `JSON.parse` macht daraus wieder „ï“.
- **Payload-Key `target`:** Ein Event-Payload mit dem Key `"target"` interpretiert htmx als
  CSS-Selektor für das Element, an dem das Event gefeuert wird. Nimm `"field"` oder `"id"`.
- **Header + leerer Body + 200:** Ohne `hx-swap="none"` am Formular setzt htmx den leeren Body
  in das Target ein. Je nach Target ist das gewollt (Modal leeren) oder nicht.
- `hx-trigger="rider-saved from:body"` ohne `from:body` hört nur auf Events am **eigenen**
  Element. Das Event kommt aber vom Formular im Modal.
- Der Lade-Request der Tabelle braucht die aktuelle Suche: `hx-include="#filter"`. Sonst
  verliert der Benutzer seinen Filter nach jedem Speichern.
- Team-Select: „kein Team“ als `value=""`, in Go zu `NULL`.

## Selbsttest

1. Warum ist `hx-target="#rider-table"` im Modal-Formular eine schlechte Idee?
2. Was bedeutet `from:body`, und warum funktioniert es (Stichwort Bubbling)?
3. Wann `HX-Trigger`, wann `HX-Trigger-After-Swap`?
4. Warum müssen Nicht-ASCII-Zeichen im Header-JSON escaped werden?
5. Wie würde eine zweite Seite (Team-Detail mit Fahrerliste) auf neue Fahrer reagieren, ohne dass du das Modal änderst?

## Wiedererkennen

Im Arbeitsprojekt: `events.go` mit `CrudSuccessEvent`, `CrudCreatedEvent`, `CrudRowUpdatedEvent`,
`CrudFailedEvent` und `SetTrigger` mit `asciiJSON`. Die Overview-Tabelle hört auf
`crud-changed from:body`. Der lange Kommentar dort zum Key `"target"` ist genau die
Stolperfalle von oben. Lies den Kommentar in `events.go` nach diesem Projekt noch einmal.
Du wirst jeden Satz verstehen.
