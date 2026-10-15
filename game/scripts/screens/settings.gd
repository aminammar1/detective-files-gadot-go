extends Control

## Settings: audio toggles, wire info, credits.

var _music_btn: Button
var _sfx_btn: Button


func _ready() -> void:
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var v := DFUI.vbox(16)
	page.add_child(v)
	var head := DFUI.hbox(12)
	var t := DFUI.title("DESK DRAWER")
	t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	head.add_child(t)
	var back := DFUI.button("Back")
	head.add_child(back)
	back.pressed.connect(func() -> void: Nav.go("menu"))
	v.add_child(head)
	v.add_child(DFUI.hseparator())

	v.add_child(DFUI.small("SOUND"))
	_music_btn = BoardUI.paper_button("Music")
	v.add_child(_music_btn)
	_music_btn.pressed.connect(_on_music)
	_sfx_btn = BoardUI.paper_button("Effects")
	v.add_child(_sfx_btn)
	_sfx_btn.pressed.connect(_on_sfx)
	_paint_toggles()

	v.add_child(DFUI.small("WIRE"))
	v.add_child(DFUI.body("Precinct wire: " + ApiClient.base_url))
	v.add_child(DFUI.small("If the desk can't reach the precinct, set DF_API_URL and restart."))

	v.add_child(BoardUI.note("NoteCream", "Credits",
		"Music: Kevin MacLeod (incompetech.com), CC BY 4.0 — “Fast Talkin” (the desk bed), “Dances and Dames” (the intro). Icons and sounds: Kenney (kenney.nl), CC0. Type: Special Elite (Apache-2.0) and IBM Plex Sans (OFL). All persons and cases fictional.",
		"Detective Files v1 — The Last Train.", 640.0))


func _paint_toggles() -> void:
	_music_btn.text = "Music: on" if Nav.music_on else "Music: off"
	_music_btn.icon = null
	_sfx_btn.text = "Effects: on" if Nav.sfx_on else "Effects: off"
	_sfx_btn.icon = null


func _on_music() -> void:
	Nav.set_audio(not Nav.music_on, Nav.sfx_on)
	if Nav.music_on:
		Nav.play_music("fast_talkin")
		Nav.play_sfx("ui_select")
	_paint_toggles()


func _on_sfx() -> void:
	Nav.set_audio(Nav.music_on, not Nav.sfx_on)
	_paint_toggles()
	Nav.play_sfx("ui_select")
