extends Control

## Auth: login/register toggle, show-password, validation, loading state.

var _mode_register: bool = false
var _email: LineEdit
var _username: LineEdit
var _user_row: VBoxContainer
var _password: LineEdit
var _show_btn: Button
var _submit_btn: Button
var _toggle_btn: Button
var _form_error: Label
var _status_box: VBoxContainer
var _busy: bool = false


func _ready() -> void:
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var parts := DFUI.centered_scroll(460)
	page.add_child(parts[0])
	var v: VBoxContainer = parts[1]

	var term := BoardUI.terminal([["CASE", "the last train"], ["STATUS", "awaiting badge"]])
	term.custom_minimum_size = Vector2(460, 0)
	v.add_child(term)
	var hello := BoardUI.note("NoteYellow", "Sign in",
		"Show your badge at the night desk. New detective? Flip to Create a badge and pick a codename.",
		"", 460.0)
	v.add_child(hello)

	_toggle_btn = DFUI.button("New detective? Create a badge")
	v.add_child(_toggle_btn)

	_email = DFUI.input("e.g. violet_hart@example.com")
	v.add_child(_labeled("Badge email", _email))
	_username = DFUI.input("e.g. night_owl")
	_user_row = _labeled("Codename (register only)", _username)
	v.add_child(_user_row)
	_user_row.visible = false

	var pw_row := DFUI.hbox(8)
	_password = DFUI.input("Password, 8+ characters")
	_password.secret = true
	_password.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	pw_row.add_child(_password)
	_show_btn = DFUI.icon_button("show_password", "Show")
	pw_row.add_child(_show_btn)
	v.add_child(_labeled("Password", pw_row))

	_form_error = DFUI.error_label("")
	_form_error.visible = false
	v.add_child(_form_error)

	_status_box = DFUI.vbox(8)
	v.add_child(_status_box)

	var row := DFUI.hbox(12)
	_submit_btn = DFUI.button("Sign in")
	row.add_child(_submit_btn)
	DFUI.primary(_submit_btn)
	var back := DFUI.button("Back")
	row.add_child(back)
	v.add_child(row)

	_toggle_btn.pressed.connect(_on_toggle)
	_show_btn.pressed.connect(_on_show)
	_submit_btn.pressed.connect(_on_submit)
	back.pressed.connect(func() -> void: Nav.go("startup"))
	_password.text_submitted.connect(func(_t: String) -> void: _on_submit())


func _labeled(label_text: String, field: Control) -> VBoxContainer:
	var w := DFUI.vbox(8)
	var l := DFUI.small(label_text)
	w.add_child(l)
	field.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	w.add_child(field)
	return w


func _on_toggle() -> void:
	Nav.play_sfx("ui_select")
	_mode_register = not _mode_register
	_toggle_btn.text = "Have a badge? Sign in" if _mode_register else "New detective? Create a badge"
	_submit_btn.text = "Create badge" if _mode_register else "Sign in"
	_user_row.visible = _mode_register
	_clear_error()


func _on_show() -> void:
	_password.secret = not _password.secret
	_show_btn.text = "Hide" if not _password.secret else "Show"
	_show_btn.icon = ArtPaths.icon("hide_password" if not _password.secret else "show_password")


func _clear_error() -> void:
	_form_error.visible = false
	_form_error.text = ""


func _fail(msg: String) -> void:
	_form_error.text = msg
	_form_error.visible = true
	Nav.play_sfx("ui_error")


func _set_busy(b: bool) -> void:
	_busy = b
	_email.editable = not b
	_username.editable = not b
	_password.editable = not b
	_submit_btn.disabled = b
	_toggle_btn.disabled = b
	for c in _status_box.get_children():
		c.queue_free()
	if b:
		_status_box.add_child(LoadingIndicator.make("Dusting for prints…"))


func _on_submit() -> void:
	if _busy:
		return
	_clear_error()
	var email := _email.text.strip_edges()
	var uname := _username.text.strip_edges()
	var pw := _password.text
	if "@" not in email or "." not in email:
		_fail("That email won't reach any precinct. Check it and try again.")
		return
	if _mode_register:
		if not _valid_username(uname):
			_fail("Codenames run 3–20 letters, numbers, or underscores.")
			return
		if pw.length() < 8:
			_fail("Passwords need at least 8 characters. Make it a good one.")
			return
	elif pw == "":
		_fail("Enter your password.")
		return
	_set_busy(true)
	var res: Dictionary
	if _mode_register:
		res = await ApiClient.api("POST", "/auth/register",
			{"email": email, "username": uname, "password": pw}, false)
	else:
		res = await ApiClient.api("POST", "/auth/login",
			{"email": email, "password": pw}, false)
	_set_busy(false)
	if not res.get("ok", false):
		_fail(str(res.get("message", "Sign-in failed.")))
		return
	var data: Dictionary = res["data"]
	Session.set_auth(str(data.get("token", "")), data.get("user", {}))
	Nav.play_sfx("ui_confirm")
	# Pull the profile before entering.
	var me := await ApiClient.api("GET", "/auth/me")
	if me.get("ok", false):
		Session.set_profile((me["data"] as Dictionary).get("profile", {}))
	Nav.go("menu")


func _valid_username(u: String) -> bool:
	if u.length() < 3 or u.length() > 20:
		return false
	for i in u.length():
		var ch := u[i]
		if not ((ch >= "a" and ch <= "z") or (ch >= "A" and ch <= "Z")
				or (ch >= "0" and ch <= "9") or ch == "_"):
			return false
	return true
