class_name DFUI
extends RefCounted

## Central UI factory. Per art bible §4/§9/§10 this file is the ONLY place that
## applies per-node theme overrides, and only for these documented reasons:
##  1. Typography roles (Display/Heading/Subheading use Special Elite — the
##     project Theme can only set one default font, IBM Plex Sans).
##  2. Primary buttons (gold outline). The engine gives Button one style set,
##     used here for the quiet secondary look; the gold set lives in this same
##     Theme under "PrimaryButton/*" and is applied by make_primary().
##  3. Placeholder photo tiles (monogram_tile): panel fill + gold initials
##     until commissioned art lands (NEEDED.md).
## Board paper/danger/tab/ink styling is copied from the same Theme by
## scripts/ui/board_ui.gd (theme()/panel_as()/ink()/button_as) — see its header.

static var _display_font: FontFile
static var _body_font: FontFile


static func display_font() -> FontFile:
	if _display_font == null:
		_display_font = load(ArtPaths.FONT_DISPLAY)
	return _display_font


static func page() -> MarginContainer:
	var m := MarginContainer.new()
	m.set_anchors_preset(Control.PRESET_FULL_RECT)
	m.add_theme_constant_override("margin_left", 32)
	m.add_theme_constant_override("margin_right", 32)
	m.add_theme_constant_override("margin_top", 32)
	m.add_theme_constant_override("margin_bottom", 32)
	return m


static func vbox(sep: int = 16) -> VBoxContainer:
	var v := VBoxContainer.new()
	v.add_theme_constant_override("separation", sep)
	return v


static func hbox(sep: int = 16) -> HBoxContainer:
	var h := HBoxContainer.new()
	h.add_theme_constant_override("separation", sep)
	return h


static func _role(label: Label, font: Font, size: int, color: Color) -> Label:
	label.add_theme_font_override("font", font)
	label.add_theme_font_size_override("font_size", size)
	label.add_theme_color_override("font_color", color)
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return label


static func title(text: String) -> Label:
	# Display role: Special Elite 40, primary. One per screen.
	var l := Label.new()
	l.text = text
	_role(l, display_font(), 40, ArtPaths.PRIMARY)
	return l


static func heading(text: String) -> Label:
	var l := Label.new()
	l.text = text
	_role(l, display_font(), 28, ArtPaths.PRIMARY)
	return l


static func subheading(text: String) -> Label:
	# Card titles, suspect names: Special Elite 20, gold.
	var l := Label.new()
	l.text = text
	_role(l, display_font(), 20, ArtPaths.GOLD)
	return l


static func body(text: String) -> Label:
	# Inherits theme font; only color is guaranteed here.
	var l := Label.new()
	l.text = text
	l.add_theme_color_override("font_color", ArtPaths.PRIMARY)
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return l


static func small(text: String) -> Label:
	var l := Label.new()
	l.text = text
	l.add_theme_font_size_override("font_size", 14)
	l.add_theme_color_override("font_color", ArtPaths.SECONDARY)
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return l


static func error_label(text: String) -> Label:
	var l := Label.new()
	l.text = text
	l.add_theme_font_size_override("font_size", 14)
	l.add_theme_color_override("font_color", ArtPaths.ERROR)
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return l


static func success_label(text: String) -> Label:
	var l := Label.new()
	l.text = text
	l.add_theme_color_override("font_color", ArtPaths.SUCCESS)
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return l


static func button(text: String) -> Button:
	# Secondary (quiet) look comes from the Theme itself. No overrides.
	var b := Button.new()
	b.text = text
	b.focus_mode = Control.FOCUS_ALL
	return b


static func primary(btn: Button) -> Button:
	# Reason 2 (see file header): apply the gold variation from the Theme.
	# Must be called AFTER btn enters the tree, so lookups resolve against
	# the project theme (not the fallback gray).
	for kind in ["normal", "hover", "pressed", "disabled", "focus"]:
		if btn.has_theme_stylebox(kind, "PrimaryButton"):
			btn.add_theme_stylebox_override(kind, btn.get_theme_stylebox(kind, "PrimaryButton"))
	for color_name in ["font_color", "font_hover_color", "font_hover_pressed_color",
			"font_pressed_color", "font_disabled_color", "font_focus_color"]:
		if btn.has_theme_color(color_name, "PrimaryButton"):
			btn.add_theme_color_override(color_name, btn.get_theme_color(color_name, "PrimaryButton"))
	if btn.has_theme_font("font", "PrimaryButton"):
		btn.add_theme_font_override("font", btn.get_theme_font("font", "PrimaryButton"))
	return btn


static func input(placeholder_text: String) -> LineEdit:
	var e := LineEdit.new()
	e.placeholder_text = placeholder_text
	e.clear_button_enabled = true
	return e


static func card() -> PanelContainer:
	var p := PanelContainer.new()
	return p


static func monogram_tile(text: String, size: int = 64) -> PanelContainer:
	# Placeholder portrait tile: panel fill + gold initials (NEEDED.md).
	var p := PanelContainer.new()
	p.custom_minimum_size = Vector2(size, size)
	var c := CenterContainer.new()
	c.set_anchors_preset(Control.PRESET_FULL_RECT)
	var l := Label.new()
	var initials := ""
	for part in text.split(" "):
		if part != "":
			initials += part.left(1)
	l.text = initials.left(2).to_upper()
	l.add_theme_font_override("font", display_font())
	l.add_theme_font_size_override("font_size", 24)
	l.add_theme_color_override("font_color", ArtPaths.GOLD)
	c.add_child(l)
	p.add_child(c)
	return p


static func icon_rect(icon_name: String, color: Color = ArtPaths.SECONDARY, px: int = 24) -> TextureRect:
	var t := TextureRect.new()
	t.texture = ArtPaths.icon(icon_name, color)
	t.custom_minimum_size = Vector2(px, px)
	t.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
	t.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
	t.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return t


static func icon_button(icon_name: String, text: String, color: Color = ArtPaths.SECONDARY) -> Button:
	var b := button(text)
	b.icon = ArtPaths.icon(icon_name, color)
	b.expand_icon = true
	return b


static func centered_scroll(min_width: float) -> Array:
	# Full-rect vertical scroll whose content is centered when short and
	# scrolls when tall. Returns [scroll, content]. Content children fill width.
	var scroll := ScrollContainer.new()
	scroll.set_anchors_preset(Control.PRESET_FULL_RECT)
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	var outer := VBoxContainer.new()
	outer.alignment = BoxContainer.ALIGNMENT_CENTER
	outer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	outer.size_flags_vertical = Control.SIZE_EXPAND_FILL
	scroll.add_child(outer)
	var content := vbox(16)
	content.custom_minimum_size = Vector2(min_width, 0)
	content.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	content.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	outer.add_child(content)
	return [scroll, content]


static func list_scroll() -> Array:
	# Full-area vertical scroll with a top-aligned expanding list. [scroll, list].
	var scroll := ScrollContainer.new()
	scroll.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	var list := vbox(16)
	list.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	list.size_flags_vertical = Control.SIZE_EXPAND_FILL
	scroll.add_child(list)
	return [scroll, list]


static func spacer(px: int) -> Control:
	var c := Control.new()
	c.custom_minimum_size = Vector2(px, px)
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return c


static func hseparator() -> HSeparator:
	var s := HSeparator.new()
	s.add_theme_color_override("color", Color(0.22745, 0.24314, 0.32157, 1))
	return s
