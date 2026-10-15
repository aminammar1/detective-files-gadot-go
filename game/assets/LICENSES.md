# Asset licenses

Every third-party file under `game/assets/` is listed here **before** it is used
by any scene. License texts / source notes live in `_licenses/<pack-name>/`.
Download date for all rows below: 2026-10-15.

| File(s) | Source URL | Author | License |
|---------|-----------|--------|---------|
| `fonts/SpecialElite-Regular.ttf` | https://github.com/google/fonts/tree/main/apache/specialelite | Astigmatic (Brian J. Bonislawsky) | Apache-2.0 — see `_licenses/google-fonts/SpecialElite-Apache-2.0-LICENSE.txt`. NOTE: not on the original allowed-list (CC0/CC-BY/OFL/MIT); shipped because the spec mandates Special Elite as one of the two fonts. Permits commercial use + embedding with notice preserved. |
| `fonts/IBMPlexSans-Variable.ttf` (renamed from `IBMPlexSans[wdth,wght].ttf`) | https://github.com/google/fonts/tree/main/ofl/ibmplexsans | IBM (Mike Abbink, Bold Monday) | OFL 1.1 — see `_licenses/google-fonts/IBMPlexSans-OFL.txt`. Variable font; Regular (400) used directly, SemiBold via `FontVariation` wght=600. |
| `icons/*.png` (26 files, white 2x, canonical names per `docs/art-bible.md` §5) | https://kenney.nl/assets/game-icons | Kenney | CC0 1.0 — see `_licenses/kenney-game-icons/SOURCE.txt`. Tinted at display time to `secondary`/`gold`/`disabled`/`ink`; never shipped in default colors. `loading_spinner.png` is `contrast.png` rotated in code (documented substitute: pack has no spinner glyph). `logo_mark.png` is the pack's magnifier glyph, used only in the game logo. |
| `portraits/*.png` (4 files, 512×512) | https://opengameart.org/content/portrait-pack | Calciumtrice | CC BY 3.0 — see `_licenses/calciumtrice/CREDIT.txt`. One painted hand across all four; cast: mustached railway man (Marsh), young long-haired porter (Tommy), silver-blonde singer (Vivienne), young brunette captain (Kowalski). |
| `audio/ui_*.ogg` (8 files) | https://kenney.nl/assets/interface-sounds and https://opengameart.org/content/interface-sounds | Kenney | CC0 1.0 — see `_licenses/kenney-interface-sounds/`. |
| `audio/music/fast_talkin.mp3`, `audio/music/dances_and_dames.mp3` | https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100590 and ...?isrc=USUAN1100595 | Kevin MacLeod | CC BY 4.0 — credit "Kevin MacLeod (incompetech.com), CC BY 4.0"; see `_licenses/incompetech/CREDIT.txt`. Shown on the Settings screen. |

Evaluated but NOT shipped (no files in `assets/` from these):
- Kenney UI Pack 2.0 (https://kenney.nl/assets/ui-pack, CC0) — light-grey cartoon
  chrome does not map to the noir palette; v1 chrome is `StyleBoxFlat` per the
  art-bible fallback clause. See `_licenses/SOURCE-kenney-ui-pack-evaluated-not-shipped.txt`.
- Kenney UI Audio 50-sound pack (https://kenney.nl/assets/ui-audio, CC0) — the
  newer Interface Sounds set covers all needs; one set only.

Rule: if a file exists in `game/assets/` and is not in this table, the game must
not reference it — ask the project owner for the source first.
