# Needed assets

Slots the game expects. Drop files at the exact paths; each slot is referenced by
one constant in `game/scripts/state/ArtPaths.gd` (Phase 1), so no scene edits are
needed. After adding a file, record its license in `LICENSES.md`.

Shipped (do not replace without approval): `fonts/`, `icons/*.png`, `audio/`.

## Fonts (shipped)

`game/assets/fonts/SpecialElite-Regular.ttf`, `IBMPlexSans-Variable.ttf`.

## Chrome (intentionally StyleBoxFlat — no files needed)

Kenney UI Pack 2.0 was evaluated and rejected for v1 (light-grey cartoon chrome
does not map to the noir palette). Panels/buttons/inputs are `StyleBoxFlat` per
`docs/art-bible.md` §3–§4. Do not add 9-slice PNGs without an art-bible amendment.

## Portraits — 512×512 PNG, single consistent hand (SHIPPED)

Portrait Pack by Calciumtrice (CC BY 3.0, see LICENSES.md). Landed 2026-10-15.

| File | Shows / screen |
|------|---------------|
| `game/assets/portraits/edwin_marsh_512.png` | Mustached railway man (Marsh) / Board, statements |
| `game/assets/portraits/tommy_beck_512.png` | Young long-haired porter (Tommy) / Board, statements |
| `game/assets/portraits/vivienne_hart_512.png` | Silver-blonde singer (Vivienne) / Board, statements |
| `game/assets/portraits/ada_kowalski_512.png` | Young brunette captain (Kowalski) / Menu, case file |

Victim needs no portrait (seen only as coroner's-slip text). Evidence art
below still uses placeholder tiles by design.

## The Tunis Ledger slots (case 2 — placeholders by design)

Same one-hand rule will apply when these are commissioned; until then the
board shows monogram tiles (suspects) and dark tiles (places, exhibits),
exactly like v1 did before its portraits landed.

| File | Shows |
|------|-------|
| `game/assets/portraits/marco_silvestri_512.png` | Italian cargo clerk, 45, jacket buttoned |
| `game/assets/portraits/yasmine_benammar_512.png` | Bar singer, 34, Corail Brûlé |
| `game/assets/portraits/bilal_haddad_512.png` | Night watchman, 28, cap straight |
| `game/assets/backgrounds/hotel_office_1920.png` | Aubert's office, razored ledger |
| `game/assets/backgrounds/hotel_bar_1920.png` | Majestic bar after closing |
| `game/assets/backgrounds/marina_quay_1920.png` | Dawn quay, harbormaster's hut |
| `game/assets/backgrounds/cinema_etoile_1920.png` | Cinéma after the late show |
| `game/assets/backgrounds/service_stair_1920.png` | Service stair, loose step |
| `game/assets/evidence/anisette_800.png` | Cloudy anisette glass |
| `game/assets/evidence/doctor_slip_800.png` | Amended doctor's slip |
| `game/assets/evidence/ledger_800.png` | Razored harbor ledger |
| `game/assets/evidence/manifests_800.png` | Two forged bills of lading |
| `game/assets/evidence/dispensary_800.png` | Bar dispensary book |
| `game/assets/evidence/stub_800.png` | Torn cinema stub, row C |
| `game/assets/evidence/banknote_800.png` | Thousand-franc safe-run note |
| `game/assets/evidence/wax_key_800.png` | Sealing-wax key impression |

## Backdrops — 1920×1080, one painterly noir set (commission)

| File | Shows / screen |
|------|---------------|
| `game/assets/backgrounds/dining_car_1920.png` | Night dining car, white cloths, icebox / Investigation locations |
| `game/assets/backgrounds/berth_c_1920.png` | Sleeping berth, made bunk, morning light / Investigation locations |
| `game/assets/backgrounds/observation_deck_1920.png` | Open tail deck, marsh wind, night / Investigation locations |
| `game/assets/backgrounds/vestibule_car3_1920.png` | Rattling car joint, conductor's cubby / Investigation locations |

Until these land, location headers are charcoal bands with the location name.

## Evidence close-ups — 800×600 (commission)

| File | Shows |
|------|-------|
| `game/assets/evidence/flask_800.png` | Monogrammed silver hip flask |
| `game/assets/evidence/pillow_note_800.png` | Doctor's amended examination slip |
| `game/assets/evidence/satchel_800.png` | Slit leather payroll satchel |
| `game/assets/evidence/timetable_800.png` | Timetable with circled 3:40 stop |
| `game/assets/evidence/medicine_log_800.png` | Steward's medicine-chest log |
| `game/assets/evidence/bond_stub_800.png` | Torn rail-bond half |
| `game/assets/evidence/key_registry_800.png` | Conductor's key log |
| `game/assets/evidence/matchbook_800.png` | Lipstick-marked matchbook |

Until these land, evidence rows show the `evidence_pin` icon tile.

## Audio — commissioned sting (nothing shipped)

A 3–5 second logo sting (muted trumpet + vinyl crackle, noir) for the Startup
logo moment. Evaluated and rejected: Kenney Music Jingles (8-bit/hits only).
Currently the intro bed ("Dances and Dames") covers the logo moment.

## Icon gaps (pack has no glyph — use text label, do not mix sets)

No map-pin (location uses `target`), no eye (show/hide password use zoom
magnifier pair), no spinner (`loading_spinner` = `contrast` rotated in code).
If Kenney's Game Icons Expansion fills these in the same hand, propose an
amendment to `docs/asset-picks.md` first.
