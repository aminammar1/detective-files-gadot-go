extends Control

## Startup: masthead, backend health check, graceful offline state.

var _status_box: VBoxContainer
var _continue_btn: Button
var _retry_btn: Button
var _online: bool = false


func _ready() -> void:
	Nav.play_music("dances_and_dames")  # intro bed; the desk switches to fast_talkin
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var parts := DFUI.centered_scroll(560)
	page.add_child(parts[0])
	var v: VBoxContainer = parts[1]
	v.add_child(BoardUI.logo())
	v.add_child(DFUI.spacer(8))

	_status_box = DFUI.vbox(12)
	v.add_child(_status_box)

	var row := DFUI.hbox(12)
	row.alignment = BoxContainer.ALIGNMENT_CENTER
	_continue_btn = DFUI.button("Continue")
	_continue_btn.disabled = true
	row.add_child(_continue_btn)
	DFUI.primary(_continue_btn)
	_retry_btn = DFUI.button("Retry")
	_retry_btn.visible = false
	row.add_child(_retry_btn)
	var exit_btn := DFUI.button("Exit")
	row.add_child(exit_btn)
	v.add_child(row)

	_continue_btn.pressed.connect(_on_continue)
	_retry_btn.pressed.connect(_on_retry)
	exit_btn.pressed.connect(func() -> void: get_tree().quit())
	_check()


func _set_status(node: Node) -> void:
	for c in _status_box.get_children():
		c.queue_free()
	_status_box.add_child(node)


func _check() -> void:
	_online = false
	_continue_btn.disabled = true
	_retry_btn.visible = false
	_set_status(LoadingIndicator.make("Dust off the wire… contacting the precinct."))
	var h := await ApiClient.health()
	if not h.get("ok", false):
		_offline(str(h.get("message", "The precinct isn't answering.")))
		return
	var r := await ApiClient.readiness()
	if not r.get("ok", false):
		_offline("The precinct picked up, but the files room is dark. Try again shortly.")
		return
	var data: Dictionary = r.get("data", {})
	if not bool(data.get("postgres", false)) or not bool(data.get("redis", false)):
		_offline("The precinct picked up, but the files room is dark. Try again shortly.")
		return
	_online = true
	# Session present? Validate it before offering the menu.
	if Session.is_logged_in():
		_set_status(LoadingIndicator.make("Checking your badge…"))
		var me := await ApiClient.api("GET", "/auth/me")
		if me.get("ok", false):
			var d: Dictionary = me["data"]
			Session.set_profile(d.get("profile", {}))
			_ready_to_go("Badge verified. The night desk is yours.")
			return
		# Invalid token: Session.expire() already cleared it; fall to login.
	_set_status(DFUI.small("Precinct wire is live. Sign in to open the case files."))
	_ready_to_go("")


func _ready_to_go(_note: String) -> void:
	_continue_btn.disabled = false
	_continue_btn.grab_focus()


func _offline(msg: String) -> void:
	var p := ErrorPanel.make(msg, false)
	_set_status(p)
	_retry_btn.visible = true
	_retry_btn.grab_focus()


func _on_retry() -> void:
	Nav.play_sfx("ui_back")
	_check()


func _on_continue() -> void:
	if not _online:
		return
	Nav.play_sfx("ui_click")
	if Session.is_logged_in():
		Nav.go("menu")
	else:
		Nav.go("auth")
