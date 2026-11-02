# 09 – Follows, Tastatur & eigenes JS

**Branch:** `t09/follows-js` → `t09/complete`
**Neu:** htmx-Lifecycle-Events, `htmx.onLoad`, JS-Struktur neben htmx, Toasts, Lade- und Sperr-Zustände, Middleware, Request-Kontext, Cookies

## Worum es geht

Ein kleines „soziales“ Feature und gleichzeitig das Projekt, in dem du lernst, **wo JS hingehört**.

- Oben im Header: „Angemeldet als: [Fahrer-Select]“. Kein echtes Login, die Auswahl landet in einem
  Cookie. Eine **Middleware** liest das Cookie und legt den aktuellen Fahrer in den Request-Kontext.
- Fahrerliste mit **Folgen/Entfolgen**-Button pro Fahrer und Follower-Zahl.
- Seite „Mein Feed“: die Fahrer, denen ich folge, mit ihren letzten Ergebnissen.
- **Tastaturbedienung** in der Fahrerliste: ↑/↓ bewegt die Markierung, Enter öffnet Details,
  `f` folgt/entfolgt.
- Globale **Toasts** für Erfolg und Fehler, Lade-Indikatoren, Buttons während Requests gesperrt.

## Abnahme

- [ ] Fahrer im Header wechseln → Seite lädt neu, Buttons zeigen den richtigen Follow-Zustand.
- [ ] Ohne Cookie: Follow-Buttons sind ausgeblendet, Feed-Seite zeigt „Wähle einen Fahrer“.
- [ ] Folgen-Klick: Button wird zu „Entfolgen“, Follower-Zahl in **derselben Zeile** und
      „Ich folge X Fahrern“ im Header aktualisieren sich.
- [ ] Sich selbst folgen geht nicht (Button fehlt; Server lehnt es trotzdem ab, `CHECK` im Schema).
- [ ] Doppelklick auf „Folgen“ erzeugt **einen** Request (Button während Request gesperrt).
- [ ] ↑/↓/Enter/`f` funktionieren, auch **nachdem** die Liste per Suche neu geladen wurde.
- [ ] Tastatur-Shortcuts sind aus, solange der Fokus in einem Eingabefeld steht.
- [ ] Jeder fehlgeschlagene Request (4xx/5xx/Netzwerk) zeigt einen Toast, ohne dass jeder Handler daran denkt.
- [ ] Kein `onclick=` und kein `<script>` in Fragment-Templates.

## Zusammenhänge

### Middleware und Request-Kontext

```go
func CurrentRider(repo *rider.Repository) func(http.Handler) http.Handler
```

- Liest Cookie `current_rider`, lädt den Fahrer, legt ihn mit `context.WithValue` in den Kontext.
- Ein **eigener Key-Typ** (`type ctxKey struct{}`) verhindert Kollisionen.
- Ein Getter `CurrentRiderFrom(ctx) (Rider, bool)` ist der einzige Weg, ihn auszulesen.
- Ungültiges Cookie → einfach kein Fahrer, kein Fehler.

So muss kein Handler das Cookie kennen. Später (Projekt 10) liegen dort Rolle und Sprache.

### Wohin gehört JS?

htmx ersetzt HTML laufend. Daraus folgen drei Regeln:

1. **Event-Delegation** für alles, was auf Klicks/Tasten reagiert: ein Listener am `document`,
   der mit `event.target.closest("[data-…]")` prüft, ob er zuständig ist. Überlebt jeden Swap.
2. **`htmx.onLoad(el => …)`** für alles, was neues HTML **initialisieren** muss (Fokus setzen,
   `<dialog>` öffnen, Markierung wiederherstellen). Wird für die erste Seite und für jedes eingesetzte Fragment aufgerufen.
3. **Konfiguration über `data-`-Attribute**, die der Server rendert: `data-shortcut="f"`,
   `data-row-url="/riders/3"`. JS kennt keine URLs und keine Geschäftsregeln.

Struktur: ein File pro Verhalten (`toast.js`, `table-keys.js`, `modal.js`), alle im Layout geladen.

### Die wichtigsten htmx-Events

| Event | Wofür |
| ----- | ----- |
| `htmx:configRequest` | Header/Parameter zum Request hinzufügen |
| `htmx:beforeRequest` | Request abbrechen (`preventDefault`), z. B. ungespeicherte Änderungen |
| `htmx:beforeSwap` | Swap-Verhalten pro Antwort ändern (`event.detail.shouldSwap`) |
| `htmx:afterSwap` / `htmx:afterSettle` | nach dem Einsetzen |
| `htmx:responseError` | Server hat 4xx/5xx geantwortet |
| `htmx:sendError` | Netzwerkfehler, Server nicht erreichbar |

Für die globalen Fehler-Toasts brauchst du die letzten beiden, am `document.body`.
Achtung: Für eine 422 mit Formular (aus 05) willst du **keinen** Toast. Wie unterscheidest du das?

### Sperren und Laden

- `hx-disabled-elt="this"`: Element ist während des Requests `disabled`. Löst den Doppelklick.
- `hx-indicator`: kennst du aus 03.
- Klasse `htmx-request` am auslösenden Element: für CSS-Zustände wie „Button dreht sich“.

### Follow-Button: Was kommt zurück?

Der Button selbst (`outerHTML`) **plus** OOB für die Follower-Zahl in der Zeile und den Zähler im
Header. Alternativ ein Event `follow-changed` mit id. Du kennst beide Wege aus 07. Begründe deine Wahl.

## Bauplan

1. Fundament + Fahrerliste mit Suche.
2. Middleware + Kontext + Header-Select (`POST /session/rider` setzt das Cookie und antwortet mit
   `HX-Refresh: true` → die ganze Seite lädt neu).
3. Follow-Repository: `Follow`, `Unfollow`, `IsFollowing`, `FollowerCount`, `FollowingCount`.
   Für die Liste: Follow-Status und -Zahl **in einer Query** (`LEFT JOIN` + `COUNT`), nicht pro Zeile.
4. Follow-Button + OOB.
5. Feed-Seite.
6. JS: `toast.js` (hört auf `crud-failed`, `htmx:responseError`, `htmx:sendError`, eigenes `toast`-Event),
   `table-keys.js` (Delegation + `htmx.onLoad` für die Markierung nach dem Neuladen).

## Stolperfallen

- `addEventListener` direkt an Zeilen → nach der ersten Suche tot. (Delegation!)
- `htmx.onLoad` läuft auch für **kleine** Fragmente (ein Button). Immer im übergebenen Element
  suchen, nicht im ganzen Dokument.
- Shortcuts in Eingabefeldern: `if (e.target.closest("input, textarea, select, [contenteditable]")) return`.
- `f` mit gedrückter Strg-Taste ist „Suchen“ im Browser. Modifier-Tasten prüfen.
- N+1-Queries: Follower-Zahl pro Zeile einzeln abzufragen fällt bei 6 Fahrern nicht auf, bei 6000 schon.
- Cookie-Wert ist Benutzereingabe. Immer gegen die DB prüfen.

## Selbsttest

1. Warum überleben delegierte Listener einen Swap, direkte nicht?
2. Wofür `htmx.onLoad`, wofür Delegation?
3. Warum ein eigener Typ als Context-Key?
4. Wie zeigst du bei jedem Fehler einen Toast, aber nicht bei einer 422 mit Formular?
5. Was macht `HX-Refresh`, und warum passt es zum Fahrerwechsel?

## Wiedererkennen

Im Arbeitsprojekt: `static/js/` mit `shortcuts.js`, `table.js`, `feedback.js`, `overlay.js`, …,
`ui.Table{Pointer: true}` für die Tastatur-Markierung, `focusauth.RequireRoleAndCountryHtmx`
als Middleware um Handler, `TranslationMw` und `ResolveLang`, das die Sprache aus Query und Cookie liest.
Gleiches Prinzip wie dein `current_rider`.
