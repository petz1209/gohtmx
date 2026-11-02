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
