class_name ErrorPanel
extends PanelContainer

## Designed error state: real message + optional Retry. Fades in 150ms.

signal retried


static func make(message: String, with_retry: bool = false) -> ErrorPanel:
	var p := ErrorPanel.new()
	var v := DFUI.vbox(12)
	var row := DFUI.hbox(12)
	row.add_child(DFUI.icon_rect("error_dot", ArtPaths.ERROR, 24))
	var l := DFUI.error_label(message)
	l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(l)
	v.add_child(row)
	if with_retry:
		var retry := DFUI.button("Retry")
		retry.pressed.connect(func() -> void: p.retried.emit())
		var h := DFUI.hbox()
		h.alignment = BoxContainer.ALIGNMENT_END
		h.add_child(retry)
		v.add_child(h)
	p.add_child(v)
	p.modulate.a = 0.0
	var tw := p.create_tween().set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_OUT)
	tw.tween_property(p, "modulate:a", 1.0, 0.15)
	return p
