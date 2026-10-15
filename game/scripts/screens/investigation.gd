extends Control

## Investigation as a night corkboard (art bible §10): pinned photos, paper
## notes, red string for broken alibis, case folder, archive terminal, final
## report clipboard. Tabs: CASE / EVIDENCE / STATEMENTS / DOCUMENTS / MAP.
## All state stays server-side; the board only renders backend answers.

const TABS := ["CASE", "EVIDENCE", "STATEMENTS", "DOCUMENTS", "MAP"]
const BREAKING := {
	"timetable": "edwin_marsh", "medicine_log": "edwin_marsh",
	"key_registry": "edwin_marsh", "matchbook": "vivienne_hart",
	"bond_stub": "tommy_beck",
}
const TESTIMONY_OF := {"tommy_timeline": "tommy_beck", "vivienne_sighting": "vivienne_hart"}

var _tab: int = 1
var _canvas: Control
var _overlay: Control
var _scroll: ScrollContainer
var _tabbar_box: HBoxContainer
var _terminal_box: VBoxContainer
var _clip: PanelContainer
var _clip_body: VBoxContainer
var _cards: Dictionary = {}  # "sus:ID" / "ev:ID" -> card
var _suspect_id: String = ""
var _picked_evidence: String = ""
var _accuse_pick: String = ""
var _responses: Dictionary = {}


func _ready() -> void:
	if Session.case_id == "":
		Nav.go("cases")
		return
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var layout := DFUI.vbox(12)
	layout.set_anchors_preset(Control.PRESET_FULL_RECT)
	page.add_child(layout)

	var top := DFUI.hbox(12)
	var mark := DFUI.icon_rect("location", ArtPaths.GOLD, 28)
	top.add_child(mark)
	var t := DFUI.title(str(Session.case_data.get("title", "Case file")))
	t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	top.add_child(t)
	var back := DFUI.button("Case files")
	top.add_child(back)
	back.pressed.connect(func() -> void: Nav.go("cases"))
	layout.add_child(top)

	var mid := DFUI.hbox(16)
	mid.size_flags_vertical = Control.SIZE_EXPAND_FILL
	layout.add_child(mid)
	_scroll = ScrollContainer.new()
	_scroll.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	mid.add_child(_scroll)
	_canvas = Control.new()
	_canvas.custom_minimum_size = Vector2(1500, 1150)
	_canvas.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_canvas.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_canvas.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_scroll.add_child(_canvas)
	_overlay = Control.new()
	_overlay.set_anchors_preset(Control.PRESET_FULL_RECT)
	_overlay.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_canvas.add_child(_overlay)

	var rail := DFUI.vbox(12)
	rail.custom_minimum_size = Vector2(290, 0)
	mid.add_child(rail)
	_terminal_box = DFUI.vbox(8)
	rail.add_child(_terminal_box)
	var report_btn := BoardUI.paper_button("Final report")
	rail.add_child(report_btn)
	report_btn.pressed.connect(_toggle_clip)
	var fill := Control.new()
	fill.size_flags_vertical = Control.SIZE_EXPAND_FILL
	fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	rail.add_child(fill)

	_tabbar_box = DFUI.hbox(4)
	_tabbar_box.alignment = BoxContainer.ALIGNMENT_CENTER
	layout.add_child(_tabbar_box)

	_clip = PanelContainer.new()
	BoardUI.panel_as(_clip, "NoteWhite")
	_clip.custom_minimum_size = Vector2(360, 0)
	_clip.visible = false
	_clip.anchor_left = 1.0
	_clip.anchor_right = 1.0
	_clip.anchor_top = 0.5
	_clip.anchor_bottom = 0.5
	_clip.offset_left = -392.0
	_clip.offset_right = -24.0
	_clip.offset_top = -215.0
	_clip.offset_bottom = 215.0
	_clip_body = DFUI.vbox(10)
	_clip.add_child(_clip_body)
	add_child(_clip)
	_boot()


func _boot() -> void:
	_canvas.add_child(LoadingIndicator.make("Opening the case file…"))
	if Session.case_data.is_empty():
		var detail := await ApiClient.api("GET", "/cases/" + Session.case_id)
		if not detail.get("ok", false):
			return
		Session.case_data = detail["data"]
	var state := await ApiClient.api("GET", "/cases/" + Session.case_id + "/state")
	if not state.get("ok", false):
		var st := await ApiClient.api("POST", "/cases/" + Session.case_id + "/start")
		if st.get("ok", false):
			state = await ApiClient.api("GET", "/cases/" + Session.case_id + "/state")
	if state.get("ok", false):
		Session.inv_state = state["data"]
	var sus: Array = Session.case_data.get("suspects", [])
	if not sus.is_empty():
		_suspect_id = str(sus[0]["id"])
	_refresh_terminal()
	_build_tabs()
	_show_tab(_tab)


func _build_tabs() -> void:
	for c in _tabbar_box.get_children():
		c.queue_free()
	_tabbar_box.add_child(BoardUI.tab_bar(TABS, _tab, _on_tab))


func _on_tab(i: int) -> void:
	Nav.play_sfx("ui_select")
	_show_tab(i)


func _show_tab(i: int) -> void:
	_tab = i
	_cards.clear()
	for c in _canvas.get_children():
		if c != _overlay:
			c.queue_free()
	for c in _overlay.get_children():
		c.queue_free()
	_build_tabs()
	match TABS[i]:
		"CASE":
			_build_case()
		"EVIDENCE":
			_build_evidence()
		"STATEMENTS":
			_build_statements()
		"DOCUMENTS":
			_build_documents()
		"MAP":
			_build_map()
	await get_tree().process_frame
	await get_tree().process_frame
	_draw_board_dressing()


func _draw_board_dressing() -> void:
	var items: Array = []
	for k in _cards:
		items.append({"card": _cards[k], "pin": "gold" if k.begins_with("test:") else "red",
			"tape": not k.begins_with("sus:")})
	BoardUI.decorate(_canvas, _overlay, items)
	for k in _cards:
		if k.begins_with("ev:") and BREAKING.has(k.trim_prefix("ev:")):
			BoardUI.link(_canvas, _overlay, _cards[k], _cards["sus:" + BREAKING[k.trim_prefix("ev:")]])
		if k.begins_with("test:") and TESTIMONY_OF.has(k.trim_prefix("test:")):
			BoardUI.link(_canvas, _overlay, _cards[k], _cards["sus:" + TESTIMONY_OF[k.trim_prefix("test:")]])


# --- CASE tab: folder + briefing ---

func _build_case() -> void:
	var d: Dictionary = Session.case_data
	var f := BoardUI.folder("CASE FILE — " + str(d.get("title", "")).to_upper(),
		[str(d.get("briefing", "")), "— Exhibit: the audit letter —",
			str(d.get("audit_letter", ""))])
	f.custom_minimum_size = Vector2(640, 0)
	BoardUI.place(_canvas, f, Vector2(0, 0), Vector2(720, 200), 3)
	_cards["folder"] = f
	var legend := BoardUI.note("NoteYellow", "How to read this board",
		"Photos are people and exhibits. Red string runs from a breaking exhibit to the suspect it breaks, and from a signed statement to the suspect it places. Pin everything. Trust nothing.",
		"red string = a liar caught", 380.0)
	BoardUI.place(_canvas, legend, Vector2(2, 0), Vector2(720, 200), 7)
	_cards["legend"] = legend


# --- EVIDENCE tab: the board ---

func _build_evidence() -> void:
	var sus: Array = Session.case_data.get("suspects", [])
	for i in sus.size():
		var s: Dictionary = sus[i]
		var ph := BoardUI.photo(str(s["name"]), "%s · %s" % [str(s["name"]), str(s["role"])])
		ph.custom_minimum_size = Vector2(240, 0)
		BoardUI.place(_canvas, ph, Vector2(i, 0), Vector2(340, 200), i + 1)
		_cards["sus:" + str(s["id"])] = ph
	var ev := _evidence()
	if ev.is_empty():
		var bare := BoardUI.note("NoteYellow", "The board is bare",
			"Nothing pinned yet. Open MAP, walk the train, search every car — every car has something to say.",
			"", 380.0)
		BoardUI.place(_canvas, bare, Vector2(0, 1), Vector2(340, 300), 11)
		_cards["hint"] = bare
		return
	for i in ev.size():
		var e: Dictionary = ev[i]
		var col := i % 3
		var row := 1 + i / 3
		var ph2 := BoardUI.photo(str(e["name"]), "%s — %s" % [str(e["name"]), str(e.get("proves", ""))])
		BoardUI.place(_canvas, ph2, Vector2(col, row), Vector2(340, 300), 20 + i)
		_cards["ev:" + str(e["id"])] = ph2
	var tests := _testimonies()
	for i in tests.size():
		var t: Dictionary = tests[i]
		var n := BoardUI.note("NoteWhite", "SIGNED — " + str(t.get("title", "")),
			str(t.get("text", "")), "", 380.0)
		BoardUI.place(_canvas, n, Vector2(i, 3), Vector2(340, 300), 40 + i)
		_cards["test:" + str(t["id"])] = n


# --- STATEMENTS tab: interrogation ---

func _suspect(suspect_id: String) -> Dictionary:
	for s in Session.case_data.get("suspects", []):
		if str(s["id"]) == suspect_id:
			return s
	return {}


func _build_statements() -> void:
	var sus: Array = Session.case_data.get("suspects", [])
	for i in sus.size():
		var s: Dictionary = sus[i]
		var b := BoardUI.paper_button(str(s["name"]))
		b.toggle_mode = true
		b.button_pressed = str(s["id"]) == _suspect_id
		BoardUI.place(_canvas, b, Vector2(i, 0), Vector2(420, 200), 50 + i)
		b.pressed.connect(_on_pick_suspect.bind(str(s["id"])))
	var s2 := _suspect(_suspect_id)
	if s2.is_empty():
		return
	_cards["sus:" + _suspect_id] = BoardUI.photo(str(s2["name"]), str(s2["role"]))
	_cards["sus:" + _suspect_id].custom_minimum_size = Vector2(300, 0)
	var head := BoardUI.note("NoteCream", "%s — dossier" % str(s2["name"]),
		"%s\n\nSTATES: “%s”" % [str(s2["dossier"]), str(s2["alibi"])], "", 560.0)
	BoardUI.place(_canvas, head, Vector2(0, 1), Vector2(640, 260), 60)
	_cards["dossier"] = head
	var i := 0
	for t in s2.get("topics", []):
		var q := BoardUI.paper_button(str(t["prompt"]))
		BoardUI.place(_canvas, q, Vector2(0, 2), Vector2(640, 120), 70 + i)
		q.pressed.connect(_on_ask.bind(_suspect_id, str(t["id"])))
		i += 1
	var hist: Array = _responses.get(_suspect_id, [])
	var shown := 0
	for j in hist.size():
		var r: int = hist.size() - j
		var n := BoardUI.note("NoteWhite", "Says", str(hist[r - 1]), "", 560.0)
		BoardUI.place(_canvas, n, Vector2(1, 1 + j), Vector2(640, 260), 80 + j)
		_cards["resp%d" % j] = n
		shown += 1
	for key in _responses:
		if str(key).begins_with(_suspect_id + "|"):
			for txt in _responses[key]:
				var n2 := BoardUI.note("NoteYellow", "Pressed with proof", str(txt), "", 560.0)
				BoardUI.place(_canvas, n2, Vector2(1, 1 + shown), Vector2(640, 260), 85 + shown)
				_cards["confr%d" % shown] = n2
				shown += 1
	# Present-evidence strip.
	var ev := _evidence()
	if not ev.is_empty():
		var k := 0
		for e in ev:
			var eb := BoardUI.paper_button("Present: " + str(e["name"]))
			eb.toggle_mode = true
			eb.button_pressed = str(e["id"]) == _picked_evidence
			BoardUI.place(_canvas, eb, Vector2(1, 2), Vector2(640, 110), 90 + k)
			eb.pressed.connect(_on_pick_evidence.bind(str(e["id"])))
			k += 1
		var go := BoardUI.paper_button("Present it")
		go.disabled = _picked_evidence == ""
		BoardUI.place(_canvas, go, Vector2(1, 3), Vector2(640, 110), 99)
		go.pressed.connect(_on_present)


func _on_pick_suspect(suspect_id: String) -> void:
	Nav.play_sfx("ui_select")
	_suspect_id = suspect_id
	_picked_evidence = ""
	_show_tab(2)  # suspect picking lives on the STATEMENTS tab


func _on_ask(suspect_id: String, topic_id: String) -> void:
	Nav.play_sfx("ui_click")
	var res := await ApiClient.api("POST", "/cases/" + Session.case_id + "/interrogate",
		{"suspect_id": suspect_id, "topic_id": topic_id})
	if res.get("ok", false):
		if not _responses.has(suspect_id):
			_responses[suspect_id] = []
		(_responses[suspect_id] as Array).append(str((res["data"] as Dictionary).get("response", "")))
	await _reload_state()
	_show_tab(_tab)


func _on_pick_evidence(ev_id: String) -> void:
	Nav.play_sfx("ui_select")
	_picked_evidence = "" if _picked_evidence == ev_id else ev_id
	_show_tab(_tab)


func _on_present() -> void:
	if _picked_evidence == "":
		return
	Nav.play_sfx("ui_click")
	var res := await ApiClient.api("POST", "/cases/" + Session.case_id + "/present-evidence",
		{"suspect_id": _suspect_id, "evidence_id": _picked_evidence})
	if res.get("ok", false):
		var data: Dictionary = res["data"]
		var key := "%s|%s" % [_suspect_id, _picked_evidence]
		if not _responses.has(key):
			_responses[key] = []
		(_responses[key] as Array).append(str(data.get("response", "")))
		if bool(data.get("contradiction", false)):
			Nav.play_sfx("ui_confirm")
		else:
			Nav.play_sfx("ui_error")
	_picked_evidence = ""
	await _reload_state()
	_show_tab(_tab)


# --- DOCUMENTS tab: journal ---

func _build_documents() -> void:
	var ev := _evidence()
	var prog := BoardUI.note("NoteYellow", "Journal",
		"%d exhibits pinned · %d statements signed. The file grows or the case dies." % [ev.size(), _testimonies().size()],
		"", 380.0)
	BoardUI.place(_canvas, prog, Vector2(0, 0), Vector2(340, 200), 100)
	_cards["prog"] = prog
	for i in ev.size():
		var e: Dictionary = ev[i]
		var n := BoardUI.note("NoteCream", str(e["name"]),
			"%s\n\nIt proves: %s" % [str(e.get("description", "")), str(e.get("proves", ""))], "", 380.0)
		BoardUI.place(_canvas, n, Vector2(i % 3, 1 + i / 3), Vector2(340, 420), 110 + i)
		_cards["ev:" + str(e["id"])] = n
	for i in _testimonies().size():
		var t: Dictionary = _testimonies()[i]
		var n2 := BoardUI.note("NoteWhite", "SIGNED — " + str(t.get("title", "")),
			str(t.get("text", "")), "", 380.0)
		BoardUI.place(_canvas, n2, Vector2(i, 3), Vector2(340, 420), 130 + i)
		_cards["test:" + str(t["id"])] = n2


# --- MAP tab: locations + search ---

func _loc(loc_id: String) -> Dictionary:
	for l in Session.case_data.get("locations", []):
		if str(l["id"]) == loc_id:
			return l
	return {}


func _build_map() -> void:
	var locs: Array = Session.case_data.get("locations", [])
	for i in locs.size():
		var l: Dictionary = locs[i]
		var ph := BoardUI.photo(str(l["name"]), str(l["description"]), 260.0, 150.0)
		BoardUI.place(_canvas, ph, Vector2(0, i), Vector2(480, 430), 140 + i)
		_cards["loc:" + str(l["id"])] = ph
		var s := BoardUI.paper_button("Search " + str(l["name"]))
		BoardUI.place(_canvas, s, Vector2(1, i), Vector2(480, 430), 150 + i)
		s.pressed.connect(_on_search.bind(str(l["id"])))


func _on_search(loc_id: String) -> void:
	Nav.play_sfx("ui_click")
	var res := await ApiClient.api("POST", "/cases/" + Session.case_id + "/actions",
		{"action": "search", "location_id": loc_id})
	if res.get("ok", false) and not ((res["data"] as Dictionary).get("found", []) as Array).is_empty():
		Nav.play_sfx("ui_confirm")
	await _reload_state()
	_show_tab(_tab)


# --- Terminal + clipboard ---

func _evidence() -> Array:
	return Session.inv_state.get("discovered_evidence", [])


func _testimonies() -> Array:
	return Session.inv_state.get("testimonies", [])


func _is_closed() -> bool:
	var inv: Dictionary = Session.inv_state.get("investigation", {})
	return str(inv.get("status", "active")) != "active"


func _refresh_terminal() -> void:
	for c in _terminal_box.get_children():
		c.queue_free()
	var p := Session.profile
	var rows := [
		["RANK", str(p.get("rank", "Rookie"))],
		["LEVEL", str(int(p.get("level", 1)))],
		["XP", str(int(p.get("xp", 0)))],
		["CASE", "CLOSED" if _is_closed() else "OPEN"],
		["EXHIBITS", "%d pinned" % _evidence().size()],
		["WITNESSES", "%d signed" % _testimonies().size()],
	]
	_terminal_box.add_child(BoardUI.terminal(rows))


func _toggle_clip() -> void:
	Nav.play_sfx("ui_open")
	_clip.visible = not _clip.visible
	if _clip.visible:
		_build_clip()


func _build_clip() -> void:
	for c in _clip_body.get_children():
		c.queue_free()
	_clip_body.add_child(BoardUI.stamp("FINAL REPORT"))
	if _is_closed():
		var l := Label.new()
		l.text = "Filed and judged. Read the board."
		BoardUI.ink(l, "InkLabel")
		l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
		_clip_body.add_child(l)
		return
	for s in Session.case_data.get("suspects", []):
		var b := BoardUI.paper_button(str(s["name"]))
		b.toggle_mode = true
		b.button_pressed = str(s["id"]) == _accuse_pick
		_clip_body.add_child(b)
		b.pressed.connect(_on_accuse_pick.bind(str(s["id"])))
	var conf := Label.new()
	var strong := _evidence().size() >= 7 and _testimonies().size() >= 2
	conf.text = "CONFIDENCE: %s" % ("HOLDING" if strong else "THIN")
	BoardUI.ink(conf, "InkSmall")
	_clip_body.add_child(conf)
	var chk := Label.new()
	chk.text = "Key evidence: %d pinned · %d signed" % [_evidence().size(), _testimonies().size()]
	BoardUI.ink(chk, "InkSmall")
	chk.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_clip_body.add_child(chk)
	var go := BoardUI.danger_button("CLOSE CASE")
	go.disabled = _accuse_pick == ""
	_clip_body.add_child(go)
	go.pressed.connect(_on_accuse)


func _on_accuse_pick(suspect_id: String) -> void:
	Nav.play_sfx("ui_select")
	_accuse_pick = "" if _accuse_pick == suspect_id else suspect_id
	_build_clip()


func _on_accuse() -> void:
	if _accuse_pick == "":
		return
	var s := _suspect(_accuse_pick)
	var d := DFDialog.confirm("Pass judgment",
		"Name %s as the killer of Silas Vane? The desk closes the file the moment you speak." % str(s["name"]),
		"Name them")
	add_child(d)
	if await d.chosen != "ok":
		return
	Nav.play_sfx("ui_confirm")
	var res := await ApiClient.api("POST", "/cases/" + Session.case_id + "/accusations",
		{"suspect_id": _accuse_pick})
	if not res.get("ok", false):
		return
	Session.last_result = res["data"]
	await _reload_state()
	Nav.go("results")


func _reload_state() -> void:
	var state := await ApiClient.api("GET", "/cases/" + Session.case_id + "/state")
	if state.get("ok", false):
		Session.inv_state = state["data"]
		_refresh_terminal()
