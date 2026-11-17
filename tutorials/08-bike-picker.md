# 08 – Bikes zuordnen: Picker & gestapelte Modals

**Branch:** `t08/bike-picker` → `t08/complete`
**Neu:** Autocomplete-Komponente (Template + ViewModel + JS + Endpoint), n:m-Beziehungen, zweites Modal über dem ersten, Events zwischen Modals, Komponenten-Vertrag

## Worum es geht

Das Fahrer-Modal aus 06 bekommt einen Abschnitt **Bikes**:

- Liste der zugeordneten Bikes mit Schalter „aktiv“ und „entfernen“.
- Ein **Picker** „Bike hinzufügen“: Tippen → Vorschläge (Marke Modell Jahr), Auswahl per Klick
  oder Pfeiltasten + Enter.
- Gibt es das Bike noch nicht: Button **„Neu“** im Picker öffnet ein **zweites Modal** über dem
  ersten, um ein Bike anzulegen. Nach dem Speichern schließt sich nur das obere Modal, und das neue Bike
  steht ausgewählt im Picker darunter.

Zweite Verwendung desselben Pickers: Team-Auswahl im Fahrer-Formular (statt Select).
Erst dann weißt du, ob die Komponente wirklich wiederverwendbar ist.

## Abnahme

- [ ] Tippen im Picker zeigt nach 250 ms bis zu 10 Vorschläge, Tippen von „spec“ findet alle Specialized-Bikes.
- [ ] Pfeil hoch/runter bewegt die Markierung, Enter wählt, Escape schließt die Vorschläge (nicht das Modal!).
- [ ] Nach Auswahl: Hidden-Input enthält die Bike-ID, sichtbares Feld zeigt „Specialized Demo 2025“.
- [ ] Bereits zugeordnete Bikes erscheinen nicht in den Vorschlägen.
- [ ] „Neu“ öffnet ein zweites Modal über dem ersten. Das erste bleibt mit allen Eingaben erhalten.
- [ ] Bike anlegen → oberes Modal schließt, Picker darunter enthält das neue Bike.
- [ ] Es wird dabei **keine** Tabelle im Hintergrund neu geladen.
- [ ] Derselbe Picker funktioniert für die Team-Auswahl, nur mit anderem ViewModel.
- [ ] Aktiv-Schalter und Entfernen aktualisieren nur die Bike-Liste im Modal.

## Zusammenhänge

### Was eine Komponente ausmacht

Eine htmx-Komponente besteht aus **vier** Teilen, die zusammenpassen müssen:

| Teil | Bei dir |
| ---- | ------- |
| Template | `{{define "picker"}}` im Base-Set |
| ViewModel | `ui.Picker{Name, Label, Value, SearchURL, NewURL, Placeholder}` |
| Endpoint-Vertrag | `GET <SearchURL>?q=…` liefert `<li data-id=… data-label=…>` |
| Verhalten (JS) | Tastatur, Auswahl übernehmen, Liste schließen |

Der Vertrag ist das Wichtigste. Jedes Feature, das den Picker nutzt, liefert seinen eigenen
Such-Endpoint, aber **im festen Format**. Schreib den Vertrag als Kommentar ans `Picker`-Struct.

### Der Modal-Stack

```html
<div id="modal"></div>        <!-- Ebene 1: Fahrer -->
<div id="modal-stack"></div>  <!-- Ebene 2: neues Bike -->
```

„Neu“ im Picker lädt nach `#modal-stack`. Das Fahrer-Formular in `#modal` bleibt **unverändert im DOM**.
Alle Eingaben bleiben erhalten, weil nichts davon neu gerendert wird.

Wichtig: `#modal-stack` liegt **außerhalb** von `#modal`. HTML erlaubt keine Formulare in Formularen,
und das obere Modal hat sein eigenes `<form>`.

### Wie kommt das neue Bike in den Picker?

```
POST /bikes (aus #modal-stack)
  → 200, HX-Trigger: {"picker-created": {"field": "bike_id", "id": 8, "label": "Rose Scrub 2026"}}

app.js:  document.body.addEventListener("picker-created", e => {
           // in #modal das Picker-Feld mit name=e.detail.field suchen,
           // id + label setzen, #modal-stack leeren
         })
```

Bewusst **nicht** `crud-changed` oder `bike-created`: Wenn eine Bike-Tabelle auf der Seite
auf solche Events hört, würde sie neu laden oder (bei `crud-row-created`) versuchen, eine Zeile
mit einer **fremden** id einzufügen. Picker-Events sind eine eigene Event-Familie.

Und das „Neu“-Formular weiß nicht, ob es aus einem Picker oder von der Bike-Seite geöffnet wurde.
Wie sagst du dem Server, welches Event er schicken soll? (Tipp: Parameter `?picker=bike_id` an der
`NewURL`, der als hidden input im Formular landet.)

### n:m zuordnen

`rider_bikes` hat `UNIQUE (rider_id, bike_id)`. Doppelte Zuordnung = DB-Fehler → Service übersetzt
in eine verständliche Meldung. Die Bike-Liste im Modal ist ein **eigenes Fragment**
(`rider-bikes`), das bei Hinzufügen/Entfernen/Aktiv neu kommt. Nicht das ganze Modal neu rendern,
sonst ist der Rest der Eingaben weg.

## Bauplan

1. Fundament + Fahrer-Modal (vereinfacht aus 06).
2. `ui.Picker` + Template `picker` + `picker-results` (Vorschlagsliste) im Base-Set.
3. `picker.js`: Tastaturnavigation über Event-Delegation am `document`
   (Picker werden dynamisch eingesetzt, direkte Listener gehen verloren).
4. Bike-Feature: `GET /bikes/picker?q=&exclude_rider=` (Vorschläge), `GET /bikes/new`, `POST /bikes`.
5. Rider-Feature: `GET /riders/{id}/bikes` (Fragment), `POST /riders/{id}/bikes`,
   `PUT /riders/{id}/bikes/{bikeID}/active`, `DELETE /riders/{id}/bikes/{bikeID}`.
6. Modal-Stack + `picker-created`-Listener.
7. Team-Picker im Fahrer-Formular.

## Stolperfallen

- Escape im Picker schließt auch das `<dialog>`. `event.preventDefault()` + `stopPropagation()`
  im Picker-Handler, wenn die Vorschlagsliste offen ist.
- **IDs doppelt:** Wenn beide Modals ein Feld `id="name"` haben, findet `getElementById` das falsche.
  IDs in Komponenten mit Präfix oder ganz ohne IDs (über `closest`/`querySelector` im Container arbeiten).
- Blur auf dem Input schließt die Liste, **bevor** der Klick auf einen Vorschlag ankommt.
  Lösung: `mousedown` statt `click`, oder Schließen leicht verzögern.
- Vorschläge mit `hx-trigger="input changed delay:250ms"` + `hx-sync="this:replace"` (kennst du aus 03).
- Payload-Key **nicht** `target` (siehe 06).
- Picker-Label mit „–“ oder Umlauten im Header → ASCII-Escape (siehe 06).

## Selbsttest

1. Welche vier Teile hat eine Komponente, und welcher ist der Vertrag?
2. Warum bleibt das Fahrer-Formular erhalten, während das zweite Modal offen ist?
3. Warum `picker-created` und nicht `crud-row-created`?
4. Woher weiß der Server, dass `POST /bikes` aus einem Picker kam?
5. Warum Event-Delegation am `document` statt Listener direkt am Input?

## Wiedererkennen

Im Arbeitsprojekt: `ui.Picker` mit `NewURL`, Template `autocomplete`, `picker.js`, der
`#modal-stack` in `layout.html`, `PickerCreatedEvent` mit `field` (nicht `target`), und in
sloganmatching ein eigenes `SloganCreatedEvent`, weil die Picker dort auf der **Seite** liegen
und nicht im `#modal`. Der Listener in `layout.html` sucht nur in `#modal`. Genau die Frage
„wo sucht der Listener“ aus diesem Projekt.
