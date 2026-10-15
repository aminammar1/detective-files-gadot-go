# Art Bible — Detective Files (v1, awaiting approval)

Locked visual language for the whole game. Every screen in every phase must follow this file.
If a screen needs something outside this file, propose an amendment here first — do not
improvise per-node styles.

## 1. Palette

All values are sRGB hex. There is no other palette: no pure `#000`/`#FFF`, no gradients as
decoration, no per-screen accent colors.

### Backgrounds / surfaces

| Token         | Hex       | Use                                              |
|---------------|-----------|--------------------------------------------------|
| `midnight`    | `#12151D` | App background, full-screen backdrop             |
| `charcoal`    | `#1E2028` | Secondary background, header/footer bands        |
| `panel`       | `#292C38` | Cards, dialogs, dossier cards, panels            |
| `input_bg`    | `#171A22` | Text fields, recessed wells                      |
| `panel_border`| `#3A3E52` | 1px panel outlines, dividers (decorative only)   |

### Accents (sparingly — one accent per screen region, never both at full weight)

| Token         | Hex       | Use                                                        |
|---------------|-----------|------------------------------------------------------------|
| `purple`      | `#8B7BC7` | Primary accent: selected tab, active nav, links, focus glow |
| `purple_dim`  | `#6E5FA8` | Pressed accent, progress-track fill background             |
| `gold`        | `#D9A441` | Highlight: primary-action text/border, case seals, XP, focus ring |
| `gold_bright` | `#E8BE5A` | Gold hover state only                                      |

### Text

| Token       | Hex       | Use                                            |
|-------------|-----------|------------------------------------------------|
| `primary`   | `#E9E2D2` | Headings, body, input text (warm off-white)    |
| `secondary` | `#B0AEBE` | Captions, metadata, hints, timestamps          |
| `disabled`  | `#6E7180` | Disabled text only (exempt from contrast rule) |
| `error`     | `#E07A6B` | Errors, destructive confirmation               |
| `success`   | `#8FC07A` | Verdict correct, success confirmations         |

### Contrast (WCAG 2.1, computed, not estimated)

Body text target: **4.5:1 (AA)** on every background it sits on.

| Foreground → background      | Ratio   | Pass? |
|------------------------------|---------|-------|
| primary on midnight          | 14.14   | ✅ AA |
| primary on charcoal          | 12.59   | ✅ AA |
| primary on panel             | 10.77   | ✅ AA |
| primary on input_bg          | 13.48   | ✅ AA |
| secondary on midnight        | 8.38    | ✅ AA |
| secondary on charcoal        | 7.46    | ✅ AA |
| secondary on panel           | 6.38    | ✅ AA |
| secondary on input_bg        | 7.99    | ✅ AA |
| gold on midnight             | 8.11    | ✅ AA |
| gold on charcoal             | 7.22    | ✅ AA |
| gold on panel                | 6.18    | ✅ AA |
| error on midnight / charcoal | 6.23 / 5.54 | ✅ AA |
| success on midnight / charcoal | 8.68 / 7.73 | ✅ AA |
| purple on midnight           | 5.00    | ✅ AA (accent text allowed on midnight/charcoal only) |
| disabled on midnight         | 3.77    | ⚠️ exempt — never use for information-bearing text |

Rules derived from the table:
- Body and metadata text may sit on any background. Disabled text never carries information.
- `purple` as *text* is allowed only on `midnight`/`charcoal`. On `panel` it drops to ~3.8:1,
  so on panels use `gold` or `primary` instead; purple there is shape/fill only.
- `panel_border` (1.73:1 on midnight) is decorative separation, never a state signal.
  The focus indicator is a 2px `gold` outline (8.11:1 on midnight), which passes non-text 3:1.
- Error/success text always goes on dark backgrounds, never on gold/purple fills.

## 2. Typography

Exactly two fonts. No italics except in-universe documents (see roles).

| Role            | Font          | Size | Line height | Color     | Use                                     |
|-----------------|---------------|------|-------------|-----------|-----------------------------------------|
| Display         | Special Elite | 40   | 48          | primary   | Screen titles only (one per screen)     |
| Heading         | Special Elite | 28   | 34          | primary   | Section headers, dialog titles          |
| Subheading      | Special Elite | 20   | 26          | gold      | Card titles, suspect names              |
| Body            | IBM Plex Sans | 16   | 24          | primary   | Briefings, dialogue, descriptions       |
| Body small      | IBM Plex Sans | 14   | 20          | secondary | Metadata, timestamps, hints, journal    |
| Button label    | IBM Plex Sans | 16   | — (centered)| see §4    | All buttons, semi-bold weight           |
| Document italic | Special Elite | 16   | 24          | secondary | In-universe papers (letters, tickets) — the only italic |

Size scale: **14 / 16 / 20 / 28 / 40** — no other sizes. Line height is 1.5× for 14/16,
~1.3× for 20, ~1.2× for 28/40. Body text is never all-caps; stamps/seals ("CASE CLOSED")
may be all-caps Special Elite 20 in gold.

## 3. Spacing, corners, borders, shadows

- Spacing scale (px): **8 / 16 / 24 / 32 / 48 / 64**. Screen padding 32 (1280×720) or 48
  (1920×1080). Card internal padding 24. Gap between related controls 8, between groups 24,
  between sections 32.
- Corner radius: **6px** panels/cards/inputs, **6px** buttons (rectangular noir file-card look —
  no pill buttons), **0px** full-bleed bands.
- Border widths: 1px panel outline (`panel_border`), 2px focus ring (`gold`), 2px primary-button
  border (`gold`). Nothing thicker.
- Shadows: one rule only — dialogs and raised cards get `0 4px 0 rgba(0,0,0,0.35)` (hard offset,
  like a stacked file). No blur glows, no colored shadows. Flat everywhere else.
- 9-slice: when the Kenney UI pack lands, panels/buttons use `StyleBoxTexture` 9-slice with the
  same 6px radius and palette; until then `StyleBoxFlat` with the tokens above. The swap happens
  in the single global Theme resource (§4) — never per-node.

## 4. Component states

One global Theme resource (`game/themes/noir_theme.tres`, built in Phase 1) owns every state.
No per-node style overrides except a reason documented in code.

### Button — primary (gold outline on panel: "Begin Investigation", "Accuse")
| State    | Background | Border      | Text        |
|----------|------------|-------------|-------------|
| Normal   | panel      | 2px gold    | gold        |
| Hover    | #313445*   | 2px gold_bright | gold_bright |
| Pressed  | purple_dim | 2px gold    | primary     |
| Focused  | = normal + 2px outer gold ring (keyboard only, doesn't replace hover) |
| Disabled | panel      | 1px panel_border | disabled, label prefixed "Coming soon" per spec |

### Button — secondary (quiet: "Back", "Cancel", "Journal")
Same table with 1px `panel_border` border and `secondary` text; hover raises text to `primary`
and border to purple. Never gold.

### Text input
| State    | Background | Border              | Text      |
|----------|------------|---------------------|-----------|
| Normal   | input_bg   | 1px panel_border    | primary   |
| Hover    | input_bg   | 1px purple          | primary   |
| Focused  | input_bg   | 2px gold            | primary   |
| Disabled | charcoal   | 1px panel_border    | disabled  |
| Error    | input_bg   | 2px error + 14px error message below in error color |

Placeholder text uses `disabled` color and is a real example ("e.g. violet_hart@example.com"),
never "Placeholder" / "Label" / lorem ipsum.

### Panel / card
`panel` fill, 1px `panel_border`, 6px radius, 24px padding. Selected card (e.g. active suspect):
2px `purple` left spine + unchanged body. Pinned evidence: 2px `gold` left spine. Spines are the
only selection signal — card fill never changes.

### List item (evidence row, location row)
Normal: transparent bg, primary title + secondary 14px subtitle. Hover: `charcoal` bg.
Selected: purple left spine + `charcoal` bg. Disabled/locked: disabled-color title + lock icon.

Every state must be visibly different from every other state; hover and focus are never the
same treatment (hover = fill shift, focus = gold ring).

\* `#313445` = panel lightened one step for hover fill; recorded here so hover isn't improvised.

## 5. Icons

One set only: Kenney UI/game icons (single-line, 2px stroke at 24px, `secondary` color;
active state `gold`, disabled `disabled`). No emoji as icons, no mixed sets, no per-screen
one-offs. Needed list (fixed — request approval before adding):

`back, close, continue_arrow, lock, unlock, star_xp, check, error_dot, suspect, location,
evidence_pin, journal, board, interrogate, present, accuse, logout, settings, profile,
trophy, sound_on, sound_off, show_password, hide_password, loading_spinner`

Sizes: 24px in buttons/lists, 32px in headers, 16px inline with 14px metadata. Always paired
with a text label — icons never stand alone (except window close).

## 6. Layout rules

- Minimum supported window **1280×720**; verify at 1280×720 and 1920×1080.
- Everything with containers (`VBox`/`HBox`/`Grid`/`Margin`/`Center`) and anchors.
  **No hand-set pixel offsets** — no `position = Vector2(...)` for UI layout, no magic margins.
- One screen = one root `Control` anchored full-rect + 32/48px page `MarginContainer`.
- Hierarchy per screen: exactly one primary action (gold), the rest secondary. Aligned edges:
  titles left, primary action bottom-right (dialogs) or top-right (workspace screens).
- Text: labels `autowrap` on, `clip_text` off for body; truncate only metadata with ellipsis +
  tooltip. Buttons expand horizontally to their text — German-length strings must not clip.
- Every screen designs its empty, loading, and error states with real copy
  (e.g. "No cases on file yet — check back after the next update.",
  "Dust off the wire… contacting the precinct.",
  "The precinct isn't answering. Check your connection and try again. [Retry]").

## 7. Motion

- Fades/slides **150–250ms**, `ease-out`, one element at a time (screen fade 200ms;
  dialog slides up 8px + fades 180ms; error panel fades 150ms).
- No bouncing, no elastic easing, no looping decoration, no parallax, no animated backgrounds.
- Loading = static spinner + copy; skeleton shimmer is forbidden (a spinner + "Dusting for
  prints…" line is the pattern).
- Respect reduced-motion implicitly: transitions are purely cosmetic and never gate input.

## 8. Do not

- No default Godot gray theme visible anywhere (every control covered by the global Theme).
- No pure `#000` / `#FFF` — deepest color is `midnight`, brightest text is `primary`.
- No gradients as decoration (a vignette on a *provided* backdrop photo is art, not UI chrome —
  ask before adding even that).
- No more than two fonts; no emoji; no mixed icon styles.
- No lorem ipsum, no "Button"/"Label"/"Item 1" text anywhere, even temporarily.
- No AI-generated, downloaded, or procedurally drawn stand-in *illustrations* — missing art =
  palette-correct `StyleBoxFlat`/`ColorRect` placeholder with the asset name written on it.
- No per-node style overrides without a code comment citing the reason.
- No screen without a designed empty/loading/error state.

## 9. Asset swap contract (binding for Phase 1+)

Every art slot is referenced by exactly one constant in
`game/scripts/state/ArtPaths.gd` (e.g. `const PORTRAIT_VIVIEN := "res://assets/portraits/vivien_hart_512.png"`),
and every style by the one Theme resource. Dropping in a real file = change the constant /
assign the texture in the Theme. `NEEDED.md` lists the expected filename, size, format, subject,
and screen for each slot; `LICENSES.md` records source/author/license before any file is used.

## 10. Board look (evidence-board screens, amended)

The Investigation screen — and any screen showing case material — is a night
corkboard, not a settings page. Dark board, paper notes, red string.

- Board background: flat `#161318` (near-black warm charcoal, no texture, no
  gradient). Full-bleed; page margins do not apply on the board itself.
- Paper notes: `#E8DCC0` cream, `#D9C89A` yellow note, `#EFE6D2` white note.
  Ink on paper is `#2A2620` typewriter ink, Special Elite only. Notes sit at a
  slight rotation (±0.03 rad max) with a hard offset shadow like dialogs.
- Tape strips: `#C9B98A` at 0.7 alpha, thin rotated rects across note corners.
- Pins: 10px dots, `#B33A3A` red (evidence/suspects), gold for solved links.
- Red string: `#C0392B`, 2px lines between pinned cards. Meaning is fixed and
  labelled on the board legend: string runs from a testimony to the suspect it
  places, and from a breaking exhibit to the suspect it breaks.
- Photo cards: 6px `#EFE6D2` border, inner placeholder tile (panel fill + gold
  label) until commissioned art lands, red pin top-center.
- Case folder: kraft `#B99B5F` with dark-ink Special Elite labels, holds the
  briefing and the location list.
- Final Report clipboard: dark board behind, paper sheet, red `#A93226`
  CLOSE CASE button. Always an overlay — never covering the board's center.
- Archive Terminal (right rail): dark `panel` with paper-ink text rows:
  rank, level, XP, case status. Same data as the header, terminal voice.
- Bottom tab bar: CASE / EVIDENCE / STATEMENTS / DOCUMENTS / MAP. Inactive tab:
  transparent, secondary text. Active tab: charcoal fill, gold text, 2px gold
  top edge. Exactly one active.
- Verdict stamp: red `#A93226` Special Elite on transparent, 2px red border,
  rotated −0.06 rad, stamped over the report.
- Board motion: cards fade in staggered 50ms apart on first pin (150ms each);
  strings draw after their cards. Nothing else moves.
- Board cards are content placed by the board layout algorithm (grid +
  deterministic tilt in `board_ui.gd`), not hand-set pixel offsets; this is
  the single exception to §6's container rule, and it applies to pinned case
  material only — never to chrome, buttons, or dialogs.
