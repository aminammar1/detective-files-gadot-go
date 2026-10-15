class_name LoadingIndicator
extends HBoxContainer

## Static spinner + copy (art bible §7: no skeleton shimmer).
## The contrast-glyph icon rotates; purely cosmetic, never gates input.

var _icon: TextureRect
var _spinning: bool = true


static func make(text: String) -> LoadingIndicator:
	var row := LoadingIndicator.new()
	row.add_theme_constant_override("separation", 12)
	var icon := DFUI.icon_rect("loading_spinner", ArtPaths.SECONDARY, 24)
	row.add_child(icon)
	row._icon = icon
	var l := DFUI.small(text)
	l.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(l)
	return row


func _process(delta: float) -> void:
	if _spinning and is_instance_valid(_icon):
		_icon.rotation += delta * 2.5


func stop() -> void:
	_spinning = false
