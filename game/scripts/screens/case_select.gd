extends Control

## Case Selection: locked/unlocked, difficulty, status, resume.

var _list: VBoxContainer
var _cases: Array = []


func _ready() -> void:
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var v := DFUI.vbox(16)
	page.add_child(v)
	var head := DFUI.hbox(12)
	var t := DFUI.title("CASE FILES")
	t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	head.add_child(t)
	var back := DFUI.button("Back")
	head.add_child(back)
	back.pressed.connect(func() -> void: Nav.go("menu"))
	v.add_child(head)
	v.add_child(BoardUI.gold_rule())
	var parts := DFUI.list_scroll()
	v.add_child(parts[0])
	_list = parts[1]
	_load()


func _load() -> void:
	for c in _list.get_children():
		c.queue_free()
	_list.add_child(LoadingIndicator.make("Pulling the files…"))
	var res := await ApiClient.api("GET", "/cases")
	for c in _list.get_children():
		c.queue_free()
	if not res.get("ok", false):
		var p := ErrorPanel.make(str(res.get("message", "Case files unavailable.")), true)
		_list.add_child(p)
		p.retried.connect(_load)
		return
	_cases = res["data"]
	if _cases.is_empty():
		_list.add_child(DFUI.body("No cases on file yet — check back after the next update."))
		return
	for row in _cases:
		_list.add_child(_case_card(row))


func _case_card(c: Dictionary) -> PanelContainer:
	var card := PanelContainer.new()
	BoardUI.panel_as(card, "FolderKraft")
	var h := DFUI.hbox(16)
	var tile := DFUI.monogram_tile(str(c.get("title", "?")), 72)
	h.add_child(tile)
	var v := DFUI.vbox(8)
	v.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var title := Label.new()
	title.text = str(c.get("title", "Untitled"))
	BoardUI.ink(title, "InkTitle")
	v.add_child(title)
	var status := str(c.get("status", "unlocked"))
	var line := Label.new()
	line.text = "Difficulty %s · %s" % [str(c.get("difficulty", "—")), status.capitalize()]
	if c.get("completed_at") != null:
		line.text += " · judged %s" % ("right" if bool(c.get("correct", false)) else "wrong")
	BoardUI.ink(line, "InkSmall")
	v.add_child(line)
	h.add_child(v)
	var started: bool = c.get("started_at") != null
	var done: bool = c.get("completed_at") != null
	var sealed: bool = str(c.get("status", "unlocked")) == "locked"
	var btn := BoardUI.paper_button("Sealed" if sealed else ("Review" if done else ("Resume" if started else "Open the file")))
	btn.disabled = sealed
	if sealed:
		btn.icon = ArtPaths.icon("lock", ArtPaths.DISABLED)
	h.add_child(btn)
	if not sealed:
		btn.pressed.connect(_on_open.bind(str(c.get("id", "")), done))
	if sealed:
		var req := Label.new()
		req.text = "Close The Last Train first."
		BoardUI.ink(req, "InkSmall")
		v.add_child(req)
	card.add_child(h)
	return card


func _on_open(case_id: String, done: bool) -> void:
	Nav.play_sfx("ui_open")
	Session.case_id = case_id
	var detail := await ApiClient.api("GET", "/cases/" + case_id)
	if not detail.get("ok", false):
		return
	Session.case_data = detail["data"]
	if not done:
		var st := await ApiClient.api("POST", "/cases/" + case_id + "/start")
		if st.get("ok", false):
			var state := await ApiClient.api("GET", "/cases/" + case_id + "/state")
			if state.get("ok", false):
				Session.inv_state = state["data"]
	else:
		var state2 := await ApiClient.api("GET", "/cases/" + case_id + "/state")
		if state2.get("ok", false):
			Session.inv_state = state2["data"]
	Nav.go("investigation")
