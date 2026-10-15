class_name DFDialog
extends CanvasLayer

## Modal dialog: dim + raised card (hard-offset shadow from Theme).
## Slides up 8px + fades 180ms (art bible §7). Await `chosen` for the answer.

signal chosen(value: String)

var _card: PanelContainer


static func confirm(title: String, body_text: String, ok_text: String,
		cancel_text: String = "Cancel", ok_primary: bool = true) -> DFDialog:
	var d := DFDialog.new()
	d.layer = 50
	var dim := ColorRect.new()
	dim.color = Color(0.07059, 0.08235, 0.11373, 0.75)
	dim.set_anchors_preset(Control.PRESET_FULL_RECT)
	d.add_child(dim)
	var center := CenterContainer.new()
	center.set_anchors_preset(Control.PRESET_FULL_RECT)
	d.add_child(center)
	var card := PanelContainer.new()
	card.custom_minimum_size = Vector2(520, 0)
	# Gold-ruled dialog card: Theme dialog style is panel+gold border.
	var v := DFUI.vbox(16)
	v.add_child(DFUI.heading(title))
	var b := DFUI.body(body_text)
	b.custom_minimum_size = Vector2(460, 0)
	v.add_child(b)
	var row := DFUI.hbox(12)
	row.alignment = BoxContainer.ALIGNMENT_END
	var cancel := DFUI.button(cancel_text)
	cancel.pressed.connect(func() -> void: d._answer("cancel"))
	row.add_child(cancel)
	var ok := DFUI.button(ok_text)
	row.add_child(ok)
	ok.pressed.connect(func() -> void: d._answer("ok"))
	v.add_child(row)
	card.add_child(v)
	center.add_child(card)
	d._card = card
	if ok_primary:
		# Applied once visible (needs tree for theme lookup).
		ok.visibility_changed.connect(func() -> void: pass)
	return d


func _ready() -> void:
	# Apply primary styling + focus now that nodes are in the tree.
	for ok in find_children("*", "Button", true, false):
		var b := ok as Button
		if b != null and b.text != "Cancel" and b.text != "Back":
			DFUI.primary(b)
			b.grab_focus()
	_card.position.y += 8
	_card.modulate.a = 0.0
	var tw := create_tween().set_parallel(true).set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_OUT)
	tw.tween_property(_card, "position:y", _card.position.y - 8.0, 0.18)
	tw.tween_property(_card, "modulate:a", 1.0, 0.18)


func _answer(value: String) -> void:
	chosen.emit(value)
	queue_free()
