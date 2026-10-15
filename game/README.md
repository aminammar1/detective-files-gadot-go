# Detective Files — Godot client

Godot 4.7, GDScript, no browser UI. Art contract: `docs/art-bible.md`.
One global Theme (`themes/noir_theme.tres`); role helpers in
`scripts/ui/df_ui.gd` are the only sanctioned per-node overrides.

## Run

```bash
godot --path game
DF_API_URL=http://localhost:8080 godot --path game   # default wire
```

## Checks

```bash
godot --headless --path game --quit    # parse check (errors fail it)
# Screenshots (needs API up; flows play the real backend):
xvfb-run -a godot --rendering-driver opengl3 --resolution 1280x720 --path game \
  res://tools/screenshot.tscn -- --scene=res://scenes/auth.tscn --shot=auth_720
# --flow=menu|case|board|verdict|tunis seeds a throwaway badge and plays that far.
# --view=<tab index> frames an investigation tab; --suspect=<id> picks a
# suspect (STATEMENTS tab); --clip opens the Final Report clipboard.
```

## Structure

- `scenes/` — one minimal root per screen; all layout built in code.
- `scripts/api/` — `ApiClient` autoload (single HTTPRequest, queued).
- `scripts/state/` — `Session` (token in `user://`), `Nav` (transitions + audio),
  `ArtPaths` (every art slot is one constant here).
- `scripts/ui/` — `DFUI` factories, `NoirTheme` role notes, `LoadingIndicator`,
  `ErrorPanel`, `DFDialog`.
- `scripts/screens/` — startup, auth, main_menu, case_select, investigation,
  results, profile, settings.
- `assets/` — shipped fonts/icons/audio + `LICENSES.md`; `NEEDED.md` lists
  commission slots (portraits, backdrops, evidence art).
