extends Control

## Main Menu: badge header (name, rank, level, XP) + case actions.

var _header: Label
var _sub: Label
var _status_box: VBoxContainer
var _continue_btn: Button
var _active_case: Dictionary = {}


func _ready() -> void:
	Session.expired.connect(_on_expired)
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var parts := DFUI.centered_scroll(720)
	page.add_child(parts[0])
	var v: VBoxContainer = parts[1]
	v.add_child(BoardUI.logo("NIGHT DESK · GRAYWATER"))

	var cols := DFUI.hbox(24)
	v.add_child(cols)
	var left := DFUI.vbox(12)
	left.custom_minimum_size = Vector2(300, 0)
	cols.add_child(left)
	_header = DFUI.heading("…")
	left.add_child(_header)
	_sub = DFUI.small("…")
	left.add_child(_sub)
	_status_box = DFUI.vbox(8)
	left.add_child(_status_box)

	var right := DFUI.vbox(10)
	right.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	cols.add_child(right)
	_continue_btn = BoardUI.paper_button("Continue")
	_continue_btn.disabled = true
	right.add_child(_continue_btn)
	for spec in [["New case", "cases"], ["Case files", "cases"], ["Profile", "profile"],
			["Achievements", "profile"], ["Settings", "settings"]]:
		var b := BoardUI.paper_button(spec[0])
		right.add_child(b)
		b.pressed.connect(_on_nav.bind(spec[1]))
	var logout_btn := BoardUI.paper_button("Log out")
	right.add_child(logout_btn)
	logout_btn.pressed.connect(_on_logout)
	var exit_btn := BoardUI.paper_button("Exit")
	right.add_child(exit_btn)
	exit_btn.pressed.connect(func() -> void: get_tree().quit())

	_continue_btn.pressed.connect(_on_continue)
	_refresh()


func _on_expired() -> void:
	pass  # Nav routes to auth; nothing local to save.


func _refresh() -> void:
	_status_box.add_child(LoadingIndicator.make("Reading your file…"))
	var me := await ApiClient.api("GET", "/auth/me")
	for c in _status_box.get_children():
		c.queue_free()
	if not me.get("ok", false):
		var ep := ErrorPanel.make(str(me.get("message", "Could not read your file.")), true)
		_status_box.add_child(ep)
		ep.retried.connect(_refresh)
		_header.text = "Night desk"
		_sub.text = "Your file is unavailable."
		return
	var data: Dictionary = me["data"]
	var user: Dictionary = data.get("user", {})
	Session.user = user
	var profile: Dictionary = data.get("profile", {})
	Session.set_profile(profile)
	_header.text = str(profile.get("display_name", user.get("username", "Detective")))
	_sub.text = "Rank %s · Level %d · %d XP" % [
		str(profile.get("rank", "Rookie")), int(profile.get("level", 1)), int(profile.get("xp", 0))]
	Nav.play_music("fast_talkin")
	_load_cases()


func _load_cases() -> void:
	var res := await ApiClient.api("GET", "/cases")
	if not res.get("ok", false):
		var ep2 := ErrorPanel.make(str(res.get("message", "Case files unavailable.")), true)
		_status_box.add_child(ep2)
		ep2.retried.connect(_refresh)
		return
	for row in res["data"]:
		var c: Dictionary = row
		if c.get("started_at") != null and c.get("completed_at") == null:
			_active_case = c
			break
	if _active_case.is_empty():
		_continue_btn.disabled = true
		_continue_btn.text = "Continue — no open case"
	else:
		_continue_btn.disabled = false
		_continue_btn.text = "Continue: %s" % str(_active_case.get("title", "case"))
		_continue_btn.grab_focus()


func _on_continue() -> void:
	if _active_case.is_empty():
		return
	Nav.play_sfx("ui_click")
	Session.case_id = str(_active_case.get("id", ""))
	var detail := await ApiClient.api("GET", "/cases/" + Session.case_id)
	if detail.get("ok", false):
		Session.case_data = detail["data"]
	var state := await ApiClient.api("GET", "/cases/" + Session.case_id + "/state")
	if state.get("ok", false):
		Session.inv_state = state["data"]
	Nav.go("investigation")


func _on_nav(key: String) -> void:
	Nav.play_sfx("ui_click")
	Nav.go(key)


func _on_logout() -> void:
	Nav.play_sfx("ui_back")
	await ApiClient.api("POST", "/auth/logout")
	Session.logout_local()
	Nav.go("auth")
