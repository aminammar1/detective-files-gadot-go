extends Control

## Profile: badge, record, codename edit, achievements case.

var _box: VBoxContainer
var _name_edit: LineEdit


func _ready() -> void:
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var v := DFUI.vbox(16)
	page.add_child(v)
	var head := DFUI.hbox(12)
	var t := DFUI.title("BADGE & RECORD")
	t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	head.add_child(t)
	var back := DFUI.button("Back")
	head.add_child(back)
	back.pressed.connect(func() -> void: Nav.go("menu"))
	v.add_child(head)
	v.add_child(DFUI.hseparator())
	var parts := DFUI.list_scroll()
	v.add_child(parts[0])
	_box = parts[1]
	_load()


func _load() -> void:
	for c in _box.get_children():
		c.queue_free()
	_box.add_child(LoadingIndicator.make("Reading your file…"))
	var prof := await ApiClient.api("GET", "/player/profile")
	var stats := await ApiClient.api("GET", "/player/stats")
	var ach := await ApiClient.api("GET", "/player/achievements")
	for c in _box.get_children():
		c.queue_free()
	if not prof.get("ok", false):
		var p := ErrorPanel.make(str(prof.get("message", "File unavailable.")), true)
		_box.add_child(p)
		p.retried.connect(_load)
		return
	var profile: Dictionary = (prof["data"] as Dictionary).get("profile", {})
	Session.set_profile(profile)
	var badge := PanelContainer.new()
	BoardUI.panel_as(badge, "FolderKraft")
	var h := DFUI.hbox(16)
	h.add_child(DFUI.monogram_tile(str(profile.get("display_name", "?")), 72))
	var v := DFUI.vbox(4)
	v.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var nm := Label.new()
	nm.text = str(profile.get("display_name", "Detective"))
	BoardUI.ink(nm, "InkTitle")
	v.add_child(nm)
	var rk := Label.new()
	rk.text = "Rank %s · Level %d · %d XP" % [
		str(profile.get("rank", "Rookie")), int(profile.get("level", 1)), int(profile.get("xp", 0))]
	BoardUI.ink(rk, "InkSmall")
	v.add_child(rk)
	h.add_child(v)
	badge.add_child(h)
	_box.add_child(badge)

	if stats.get("ok", false):
		var s: Dictionary = stats["data"]
		_box.add_child(DFUI.small("RECORD"))
		_box.add_child(DFUI.body("Cases opened %d · solved %d · fumbled %d · %d lifetime XP" % [
			int(s.get("cases_started", 0)), int(s.get("cases_solved", 0)),
			int(s.get("cases_failed", 0)), int(s.get("total_xp", 0))]))

	_box.add_child(DFUI.small("CODENAME"))
	var row := DFUI.hbox(8)
	_name_edit = DFUI.input("A new codename, 1–30 characters")
	_name_edit.text = str(profile.get("display_name", ""))
	_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(_name_edit)
	var save := DFUI.button("File it")
	row.add_child(save)
	DFUI.primary(save)
	_box.add_child(row)
	save.pressed.connect(_on_save)

	_box.add_child(DFUI.small("ACHIEVEMENTS"))
	if ach.get("ok", false):
		for a in ach["data"]:
			_box.add_child(_ach_row(a))
	else:
		_box.add_child(DFUI.small("The commendation drawer is stuck."))


func _ach_row(a: Dictionary) -> Control:
	var unlocked: bool = a.get("unlocked_at") != null
	var row := PanelContainer.new()
	BoardUI.panel_as(row, "NoteCream" if unlocked else "NoteWhite")
	var h := DFUI.hbox(12)
	h.add_child(DFUI.icon_rect("trophy", BoardUI.INK if unlocked else ArtPaths.DISABLED, 24))
	var v := DFUI.vbox(2)
	v.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var n := Label.new()
	n.text = str(a.get("name", "Commendation"))
	BoardUI.ink(n, "InkLabel")
	v.add_child(n)
	var d := Label.new()
	d.text = str(a.get("description", ""))
	BoardUI.ink(d, "InkSmall")
	d.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	v.add_child(d)
	h.add_child(v)
	if not unlocked:
		var l := Label.new()
		l.text = "LOCKED"
		l.custom_minimum_size = Vector2(70, 0)
		BoardUI.ink(l, "InkSmall")
		h.add_child(l)
	row.add_child(h)
	return row


func _on_save() -> void:
	var want := _name_edit.text.strip_edges()
	if want.length() < 1 or want.length() > 30:
		Nav.play_sfx("ui_error")
		return
	Nav.play_sfx("ui_click")
	var res := await ApiClient.api("PATCH", "/player/profile", {"display_name": want})
	if res.get("ok", false):
		Nav.play_sfx("ui_confirm")
		_load()
	else:
		Nav.play_sfx("ui_error")
