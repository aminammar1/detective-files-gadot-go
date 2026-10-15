# Asset picks — Detective Files (v1, awaiting approval)

Shortlist researched 2026-10-15. Nothing downloaded yet — no files will be
fetched until this list is approved. Rule for download day: fetch the asset's
own page first, confirm the license on that page, download from the official
link only, extract only the files used, record each in `game/assets/LICENSES.md`
with source URL / author / license / date, and keep the pack's own
license/readme in `game/assets/_licenses/<pack-name>/`.

Coherence decision (binding once approved): **one Kenney UI look for all
chrome + icons + SFX, the two Google Fonts, Kevin MacLeod CC-BY music, and
palette-correct `StyleBoxFlat` placeholders for portraits/backdrops/evidence
art** (see §4–5 for why). Nothing AI-generated, no emoji icons, no mixed sets.

## 1. UI kit (9-slice panels + buttons)

- **A. Kenney "UI Pack" (RECOMMENDED).** Page: https://kenney.nl/assets/ui-pack —
  Preview: same page (preview.png / sample.png) — License: **CC0** (confirmed on
  the page itself, "Creative Commons CC0", fetched 2026-10-15). Reason: 430
  sprites, true 9-slice panel/button/slider assets with vector sources, flat
  cartoon-noir-adjacent shapes that recolor cleanly to the art-bible palette
  via `StyleBoxTexture`; the pack the art bible §3/§4 9-slice plan was written
  against.
- **B. Kenney "UI Pack – Adventure".** Page: https://kenney.nl/assets/ui-pack-adventure —
  Preview: same page — License: **CC0** (same site-wide Kenney CC0 grant; to
  confirm on the page at download time). Reason: same structure as A but with a
  parchment/aged-paper tone that suits 1940s case files; fallback if the base
  pack reads too casual once themed.
- **C. Kenney "Fantasy UI Borders".** Page: https://kenney.nl/assets/fantasy-ui-borders —
  Preview: same page — License: **CC0 1.0 Universal** (stated on the
  itch.io mirror https://kenney-assets.itch.io/fantasy-ui-borders). Reason:
  130+ window/dialog borders made for 9-slice; fallback if A/B panels don't
  frame dossier cards well.

Pick: **A**, with B as the only approved alternate. One pack ships — never mix A+B borders.

## 2. Icons (fixed list, art bible §5)

- **A. Kenney "Game Icons" (RECOMMENDED).** Page: https://kenney.nl/assets/game-icons —
  Preview: same page — License: **CC0** (confirmed via page + OpenGameArt mirror
  https://opengameart.org/content/all-cc0-uploader-kenney). Reason: 105+
  single-weight line icons in one hand; covers lock, trophy, music, settings,
  suspects/locations with a single 2px stroke at 24px, exactly the icon spec.
- **B. Kenney "Game Icons (Expansion)".** Page: via https://kenney.nl/assets/tag:icon —
  Preview: same page — License: **CC0** (same grant; confirm on page at
  download). Reason: same hand as A, fills any gaps (journal, board, present)
  without mixing styles.
- **C. Kenney "Board Game Icons".** Page: https://kenney-assets.itch.io/board-game-icons —
  Preview: same page — License: **CC0 1.0 Universal** (stated on the page).
  Reason: 250+ card/dice/action symbols; fallback only for evidence-board
  symbology.

Pick: **A, extended by B only for missing glyphs**. Recolor to `secondary` /
`gold` at import; never ship Kenney's default colors.

## 3. Fonts (exactly two, no substitutes)

- **A. Special Elite + IBM Plex Sans from Google Fonts (RECOMMENDED, only option).**
  Pages: https://fonts.google.com/specimen/Special+Elite and
  https://fonts.google.com/specimen/IBM+Plex+Sans —
  License: **OFL** (both are Google Fonts OFL families; Special Elite is a
  CC0-style typewriter face by Astigmatic, Plex is IBM's OFL superfamily —
  confirm the OFL text in the downloaded TTF metadata at download time).
  Reason: the art bible's mandated pair — typewriter display + neutral grotesk
  body. Files: `SpecialElite-Regular.ttf`, `IBMPlexSans-Regular.ttf`,
  `IBMPlexSans-SemiBold.ttf`.
- B/C: none — the spec fixes these two fonts, so there are no alternates to
  shortlist. If Google Fonts is unreachable, say so and stop; do not substitute.

## 4. Portraits (3 suspects + detective, 512×512, one consistent hand)

No allowed source offers a coherent set of 1940s-noir realistic portraits under
CC0/CC-BY/OFL/MIT. What exists is mismatched in style (pixel heads, fantasy
avatars, greyscale public-domain face crops), and mixing them would violate the
one-style rule. So:

- **A. Palette-correct placeholders (RECOMMENDED for v1).** `StyleBoxFlat` /
  `ColorRect` in `panel` with the suspect name set in Special Elite 20 gold —
  per the art bible "do not" list and the spec's placeholder rule. Real
  portraits get commissioned later against `NEEDED.md`. Reason: honest, coherent,
  one-line swap via `ArtPaths.gd`.
- **B. OpenGameArt "CC0 Portraits" collection** (https://opengameart.org/content/cc0-portraits) —
  License: CC0 per collection, **but per-item licenses vary and styles clash**
  (pixel, painted, photo-crop). Reason listed only for completeness: a source
  for *individual* CC-BY-compatible placeholders if you prefer faces over
  monograms — requires per-item license confirmation, so not recommended.
- **C. itch.io "noir/detective/1920s" search** — no candidate vetted: every
  result found in research either lacked a stated license or carried NC/ND or
  "personal use only" terms, all of which are **not allowed**. Listed here so
  the search is on record; do not download from itch.io without a page that
  explicitly states CC0/CC-BY/free-for-any-use.

Pick: **A**. Portraits ship as designed placeholders in v1.

## 5. Backdrops / evidence art (1920×1080 locations, evidence close-ups)

Same situation as §4:

- **A. Palette-correct placeholders (RECOMMENDED for v1).** Charcoal/midnight
  bands with the location name — the Investigation screen is designed to read
  as a desk of files, so it works without photos. Reason: keeps one coherent
  look; swap is one constant each.
- **B. OpenGameArt "CC0 Backgrounds" collection** (https://opengameart.org/content/cc0-backgrounds) —
  License: CC0 per collection umbrella, per-item confirmation required.
  Reason: only usable if 3–4 same-style interiors (train/office/platform) can
  be confirmed from one artist; research found train *tilesets/models*
  (e.g. Kenney "Train Kit", 3D) but no matching 2D noir interior set — so B is
  a documented dead end for v1, not a pick.
- **C. Commission/custom shoot later** — `NEEDED.md` carries exact filenames,
  sizes, subjects, and screens for each slot.

Pick: **A**.

## 6. Music (one loopable noir bed + one tension sting)

- **A. Kevin MacLeod via incompetech.com (RECOMMENDED).** Catalogue:
  https://incompetech.com/music/royalty-free/music.html — License: **CC BY 4.0**
  (attribution credit in-game + in LICENSES.md; confirm per-track license on
  its track page at download). Shortlist tracks: "Fast Talkin" (1940s private-
  detective feel, bass/drums/brass — https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100590),
  "Dances and Dames" (https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100595),
  "Investigations". Reason: period-correct jazz/blues beds written for exactly
  this mood, from the spec's preferred music source.
- **B. OpenGameArt "Detective Game" collection music** (https://opengameart.org/content/detective-game —
  lists "Noire Crime Time", "A Conversation with Saul (Jazz/Blues Shuffle)",
  "Bar jazz piano song") — License: per-track CC0/CC-BY, confirm per item.
  Reason: already-curated noir jazz loops; fallback if incompetech licensing
  display changes.
- **C. Kenney "Music Loops / Music Jingles"** (https://kenney.nl/assets/category:Audio) —
  License: CC0. Reason: CC0 means zero attribution friction, but the loops are
  generic casual-game beds, not noir — fallback only.

Pick: **A (two tracks max)**. Assignment: "Dances and Dames" is the intro/title
bed (plays on Startup), "Fast Talkin" is the desk bed (menu + investigation).
A dedicated seconds-long logo sting was evaluated and rejected: Kenney's Music
Jingles pack (https://kenney.nl/assets/music-jingles, CC0, downloaded
2026-10-15) holds only 8-bit and impact-hit jingles — wrong century for noir —
and no OGA sting combined a confirmable CC0/CC-BY license with a fitting
period sound, so per the download rules nothing was taken. A commissioned
sting is logged in `game/assets/NEEDED.md`.

## 7. SFX (UI clicks, paper, stamps, door/chime)

- **A. Kenney "Interface Sounds" + "UI Audio" (RECOMMENDED).**
  Pages: https://kenney.nl/assets/ui-audio and via
  https://opengameart.org/content/interface-sounds —
  Preview: audio previews on both pages — License: **CC0** (Kenney site grant +
  OGA page states CC0). Reason: 100 + 50 button/switch/paper-click sounds in one
  hand, normalized volume, OGG files Godot plays natively; the Godot-ready
  mirror https://github.com/Calinou/kenney-ui-audio (CC0, Godot AssetLib entry
  https://godotengine.org/asset-library/asset/796) exists but **no addon will
  be added without your say-so** — raw OGGs copied into `game/assets/audio/`.
- **B. OpenGameArt "100 CC0 SFX"** (referenced from the Detective Game
  collection, https://opengameart.org/content/detective-game) — License: CC0,
  confirm per item. Reason: broader foley (doors, rain) beyond UI clicks.
- **C. freesound.org CC0/CC-BY only** — no specific pack shortlisted: every
  sound there needs per-item license confirmation at download, so it is the
  gap-filler for specific foley (train brake, stamp thud) only.

Pick: **A** for all UI SFX; C only for named gaps, one file at a time.

## 8. Godot addons

None proposed. The only candidate encountered (Calinou's kenney-ui-audio
packaging, Godot AssetLib) is a file-layout convenience, not a code dependency,
and will **not** be added without explicit approval — raw files only.

## 9. Download plan (after approval, Phase 1 start)

1. Fonts (§3A) → `game/assets/fonts/` (+ OFL text to `_licenses/google-fonts/`).
2. Kenney UI Pack (§1A) → extract only used 9-slice panel/button assets →
   `game/assets/ui/` (+ license to `_licenses/kenney-ui-pack/`).
3. Kenney Game Icons (§2A + gaps from §2B) → `game/assets/icons/` (recolored
   imports; originals + license in `_licenses/kenney-game-icons/`).
4. Kenney Interface Sounds + UI Audio (§7A) → only used OGGs →
   `game/assets/audio/` (+ license in `_licenses/kenney-audio/`).
5. MacLeod tracks (§6A) → `game/assets/audio/music/` (+ CC-BY credit).
6. Fill `LICENSES.md` (file, source URL, author, license, date) and update
   `NEEDED.md` with portrait/backdrop/evidence slots for commissioning.
7. If any URL is dead or its license page changed: stop, report, do not
   substitute — propose an amended pick.
