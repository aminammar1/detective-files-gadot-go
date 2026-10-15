class_name BoardUI
extends RefCounted

## Corkboard kit (art bible §10). The ONLY place that builds paper notes,
## photo cards, pins, tape, string, stamps, tabs, and terminal rows — always
## copied from the global Theme's Note*/FolderKraft/PhotoFrame/PaperButton/
## DangerButton/BoardTab*/Ink* entries via theme()/panel_as()/ink()/button_as().
## Board cards are content pinned by the layout algorithm below
## (grid + deterministic rotation), which §10 explicitly allows; all chrome
## elsewhere still uses containers and anchors.

const BOARD := Color(0.08627, 0.07451, 0.09412)   # #161318
const INK := Color(0.16471, 0.14902, 0.12549)     # #2A2620
const STRING_RED := Color(0.75294, 0.22353, 0.16863)
const PIN_RED := Color(0.70196, 0.22745, 0.22745)
const TAPE := Color(0.78824, 0.72549, 0.54118, 0.7)

static var _pin_styles: Dictionary = {}
static var _theme: Theme


## The .tres is the single source; Control-level custom-type lookup proved
## unreliable, so items are copied from the Theme resource as overrides —
## always here, never scattered (same dispensation as DFUI.primary()).
static func theme() -> Theme:
	if _theme == null:
		_theme = load("res://themes/noir_theme.tres")
	return _theme


static func panel_as(c: Control, type_name: String) -> void:
	if theme().has_stylebox("panel", type_name):
		c.add_theme_stylebox_override("panel", theme().get_stylebox("panel", type_name))


static func ink(l: Label, type_name: String) -> void:
	var th := theme()
	if th.has_color("font_color", type_name):
		l.add_theme_color_override("font_color", th.get_color("font_color", type_name))
	if th.has_font_size("font_size", type_name):
		l.add_theme_font_size_override("font_size", th.get_font_size("font_size", type_name))
	if th.has_font("font", type_name):
		l.add_theme_font_override("font", th.get_font("font", type_name))
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART


static func button_as(b: Button, type_name: String) -> void:
	var th := theme()
	for kind in ["normal", "hover", "pressed", "disabled", "focus"]:
		if th.has_stylebox(kind, type_name):
			b.add_theme_stylebox_override(kind, th.get_stylebox(kind, type_name))
	for color_name in ["font_color", "font_hover_color", "font_hover_pressed_color",
			"font_pressed_color", "font_disabled_color", "font_focus_color"]:
		if th.has_color(color_name, type_name):
			b.add_theme_color_override(color_name, th.get_color(color_name, type_name))
	if th.has_font_size("font_size", type_name):
		b.add_theme_font_size_override("font_size", th.get_font_size("font_size", type_name))
	if th.has_font("font", type_name):
		b.add_theme_font_override("font", th.get_font("font", type_name))


static func board_bg() -> ColorRect:
	var c := ColorRect.new()
	c.color = BOARD
	c.set_anchors_preset(Control.PRESET_FULL_RECT)
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return c


static func note(variation: String, title: String, body_text: String,
		foot: String = "", width: float = 300.0) -> PanelContainer:
	var p := PanelContainer.new()
	panel_as(p, variation)  # NoteCream / NoteYellow / NoteWhite
	var v := DFUI.vbox(8)
	if title != "":
		var t := Label.new()
		t.text = title
		ink(t, "InkTitle")
		v.add_child(t)
	if body_text != "":
		var b := Label.new()
		b.text = body_text
		ink(b, "InkLabel")
		v.add_child(b)
	if foot != "":
		var f := Label.new()
		f.text = foot
		ink(f, "InkSmall")
		v.add_child(f)
	p.add_child(v)
	p.custom_minimum_size = Vector2(width, 0)
	return p


static func folder(title: String, lines: Array) -> PanelContainer:
	var p := PanelContainer.new()
	panel_as(p, "FolderKraft")
	var v := DFUI.vbox(8)
	var t := Label.new()
	t.text = title
	ink(t, "InkTitle")
	v.add_child(t)
	for line in lines:
		var l := Label.new()
		l.text = str(line)
		ink(l, "InkLabel")
		v.add_child(l)
	p.add_child(v)
	return p


static func photo(caption: String, sub: String, w: float = 220.0,
		h: float = 140.0) -> PanelContainer:
	# White frame + dark placeholder tile (NEEDED.md) + ink caption.
	var p := PanelContainer.new()
	panel_as(p, "PhotoFrame")
	var v := DFUI.vbox(6)
	var tile := PanelContainer.new()
	tile.custom_minimum_size = Vector2(w, h)
	var l := Label.new()
	l.text = caption
	l.add_theme_font_override("font", DFUI.display_font())
	l.add_theme_font_size_override("font_size", 16)
	l.add_theme_color_override("font_color", ArtPaths.GOLD)
	l.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	l.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	tile.add_child(l)
	v.add_child(tile)
	var cap := Label.new()
	cap.text = sub
	ink(cap, "InkSmall")
	cap.custom_minimum_size = Vector2(maxf(w - 12.0, 40.0), 0)
	v.add_child(cap)
	p.add_child(v)
	return p


static func paper_button(text: String) -> Button:
	var b := Button.new()
	b.text = text
	b.focus_mode = Control.FOCUS_ALL
	button_as(b, "PaperButton")
	return b


static func danger_button(text: String) -> Button:
	var b := Button.new()
	b.text = text
	b.focus_mode = Control.FOCUS_ALL
	button_as(b, "DangerButton")
	return b


static func tab_bar(names: Array, active: int, on_pick: Callable) -> HBoxContainer:
	var h := DFUI.hbox(4)
	for i in names.size():
		var b := Button.new()
		b.text = str(names[i])
		b.toggle_mode = true
		b.button_pressed = i == active
		b.focus_mode = Control.FOCUS_ALL
		button_as(b, "BoardTabActive" if i == active else "BoardTab")
		h.add_child(b)
		b.pressed.connect(on_pick.bind(i))
	return h


static func terminal(rows: Array) -> PanelContainer:
	# Archive Terminal: dark panel, paper-ink data rows.
	var p := DFUI.card()
	var v := DFUI.vbox(6)
	var head := Label.new()
	head.text = "ARCHIVE TERMINAL"
	head.add_theme_font_override("font", DFUI.display_font())
	head.add_theme_font_size_override("font_size", 20)
	head.add_theme_color_override("font_color", ArtPaths.GOLD)
	v.add_child(head)
	for row in rows:
		var r: Array = row
		var l := DFUI.small("%s  %s" % [str(r[0]), str(r[1])])
		v.add_child(l)
	p.add_child(v)
	return p


static func stamp(text: String) -> PanelContainer:
	var p := PanelContainer.new()
	var sb := StyleBoxFlat.new()
	sb.draw_center = false
	sb.border_width_left = 2
	sb.border_width_top = 2
	sb.border_width_right = 2
	sb.border_width_bottom = 2
	sb.border_color = Color(0.66275, 0.19608, 0.14902)
	sb.content_margin_left = 16.0
	sb.content_margin_top = 8.0
	sb.content_margin_right = 16.0
	sb.content_margin_bottom = 8.0
	p.add_theme_stylebox_override("panel", sb)
	var l := Label.new()
	l.text = text
	l.add_theme_font_override("font", DFUI.display_font())
	l.add_theme_font_size_override("font_size", 28)
	l.add_theme_color_override("font_color", Color(0.66275, 0.19608, 0.14902))
	p.add_child(l)
	p.rotation = -0.06
	return p


static func pin_style(color: Color) -> StyleBoxFlat:
	var key := color.to_html()
	if not _pin_styles.has(key):
		var sb := StyleBoxFlat.new()
		sb.bg_color = color
		sb.corner_radius_top_left = 6
		sb.corner_radius_top_right = 6
		sb.corner_radius_bottom_right = 6
		sb.corner_radius_bottom_left = 6
		_pin_styles[key] = sb
	return _pin_styles[key]


## Places a card on the board canvas at a grid cell with deterministic tilt.
## cell: Vector2(col, row) in cells of (cell_w, cell_h) with origin at margin.
static func place(board: Control, card: Control, cell: Vector2,
		cell_size: Vector2, tilt_seed: int = 0) -> void:
	var margin := Vector2(48, 48)
	var pos := margin + Vector2(cell.x * cell_size.x, cell.y * cell_size.y)
	# Deterministic pseudo-jitter from the seed: ±10px, ±0.03 rad.
	var jx := float((tilt_seed * 37) % 21 - 10)
	var jy := float((tilt_seed * 53) % 21 - 10)
	var tilt := float((tilt_seed * 29) % 7 - 3) * 0.01
	card.position = pos + Vector2(jx, jy)
	card.rotation = tilt
	if card.get_parent() == null:
		board.add_child(card)


## After all cards are placed and laid out (await two frames first), pins
## tape + red string get drawn on the overlay in board-local coordinates.
static func decorate(board: Control, overlay: Control, cards: Array) -> void:
	for item in cards:
		var card: Control = item["card"]
		var center := _board_center(board, card)
		var w: float = card.size.x
		var h: float = card.size.y
		if String(item.get("pin", "red")) != "none":
			var dot := Panel.new()
			var pc := ArtPaths.GOLD if item.get("pin") == "gold" else PIN_RED
			dot.add_theme_stylebox_override("panel", pin_style(pc))
			dot.custom_minimum_size = Vector2(12, 12)
			dot.position = center + Vector2(-6, -h / 2 - 4)
			dot.mouse_filter = Control.MOUSE_FILTER_IGNORE
			overlay.add_child(dot)
		if bool(item.get("tape", false)):
			for sx in [-0.32, 0.32]:
				var strip := ColorRect.new()
				strip.color = TAPE
				strip.custom_minimum_size = Vector2(48, 14)
				strip.position = center + Vector2(sx * w - 24, -h / 2 + 4)
				strip.rotation = -0.2 if sx < 0.0 else 0.25
				strip.mouse_filter = Control.MOUSE_FILTER_IGNORE
				overlay.add_child(strip)


static func link(board: Control, overlay: Control, a: Control, b: Control) -> void:
	var line := Line2D.new()
	line.points = PackedVector2Array([_board_center(board, a), _board_center(board, b)])
	line.width = 2.0
	line.default_color = STRING_RED
	overlay.add_child(line)


static func _board_center(board: Control, card: Control) -> Vector2:
	var r: Rect2 = card.get_global_rect()
	var o: Rect2 = board.get_global_rect()
	return r.get_center() - o.position


static func gold_rule() -> ColorRect:
	var r := ColorRect.new()
	r.color = ArtPaths.GOLD
	r.custom_minimum_size = Vector2(0, 2)
	r.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	r.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return r


static func logo(tagline: String = "GRAYWATER · 1947") -> VBoxContainer:
	# The game logo: gold lens mark + typewriter lockup. One definition —
	# every screen uses this, nothing hand-builds a masthead.
	var v := DFUI.vbox(8)
	v.alignment = BoxContainer.ALIGNMENT_CENTER
	var row := DFUI.hbox(16)
	row.alignment = BoxContainer.ALIGNMENT_CENTER
	row.add_child(DFUI.icon_rect("logo_mark", ArtPaths.GOLD, 52))
	var stack := DFUI.vbox(2)
	var title := Label.new()
	title.text = "DETECTIVE FILES"
	title.add_theme_font_override("font", DFUI.display_font())
	title.add_theme_font_size_override("font_size", 44)
	title.add_theme_color_override("font_color", ArtPaths.PRIMARY)
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	stack.add_child(title)
	stack.add_child(gold_rule())
	var tag := Label.new()
	tag.text = tagline
	tag.add_theme_font_override("font", DFUI.display_font())
	tag.add_theme_font_size_override("font_size", 16)
	tag.add_theme_color_override("font_color", ArtPaths.GOLD)
	tag.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	stack.add_child(tag)
	row.add_child(stack)
	var flip := DFUI.icon_rect("logo_mark", ArtPaths.GOLD, 52)
	flip.flip_h = true
	row.add_child(flip)
	v.add_child(row)
	return v
