extends Control

## Results: verdict, supporting evidence, score, XP.

const VERDICTS := {
	"guilty": ["CASE CLOSED — RIGHT", "Marsh never saw the morning edition. The bonds, the drug, the key, and two witnesses who stopped lying — you built a chain the railway lawyers couldn't cut. Vane gets his headline after all."],
	"unproven": ["RIGHT MAN, THIN FILE", "You named Marsh, and your gut was right — but the file couldn't hold him. A correct hunch without proof is just gossip with a badge. Next time, pin every exhibit and break every alibi first."],
	"wrong": ["CASE CLOSED — WRONG", "The file went to the railway court and fell apart in an hour. Somewhere Marsh is pouring rye with a steady hand. The desk keeps your badge, barely. Read the board again — the train only lies once."],
}


func _ready() -> void:
	add_child(BoardUI.board_bg())
	var page := DFUI.page()
	add_child(page)
	var parts := DFUI.centered_scroll(560)
	page.add_child(parts[0])
	var v: VBoxContainer = parts[1]

	var r: Dictionary = Session.last_result
	if r.is_empty():
		var n0 := BoardUI.note("NoteWhite", "No verdict on file",
			"There is no judgment to read. Work the case first.", "", 560.0)
		v.add_child(n0)
		var back0 := BoardUI.paper_button("Case files")
		v.add_child(back0)
		back0.pressed.connect(func() -> void: Nav.go("cases"))
		return

	var result := str(r.get("result", "wrong"))
	var words: Array = VERDICTS.get(result, VERDICTS["wrong"])
	var report := PanelContainer.new()
	BoardUI.panel_as(report, "NoteWhite")
	var rv := DFUI.vbox(12)
	rv.add_child(BoardUI.stamp(str(words[0])))
	rv.add_child(_ink_body(str(words[1])))
	rv.add_child(BoardUI.gold_rule())

	# Supporting evidence: what the file actually held at judgment.
	var ev: Array = Session.inv_state.get("discovered_evidence", [])
	if not ev.is_empty():
		var h := Label.new()
		h.text = "THE FILE HELD"
		BoardUI.ink(h, "InkSmall")
		rv.add_child(h)
		for e in ev:
			var row := DFUI.hbox(8)
			row.add_child(DFUI.icon_rect("evidence_pin", BoardUI.INK, 16))
			var l := Label.new()
			l.text = str(e["name"])
			BoardUI.ink(l, "InkSmall")
			l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
			row.add_child(l)
			rv.add_child(row)

	rv.add_child(BoardUI.gold_rule())
	var score_line := Label.new()
	score_line.text = "Score %d · +%d XP" % [int(r.get("score", 0)), int(r.get("xp_awarded", 0))]
	BoardUI.ink(score_line, "InkTitle")
	score_line.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	rv.add_child(score_line)
	report.add_child(rv)
	v.add_child(report)
	var fresh := DFUI.small("Filing the new totals…")
	fresh.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	v.add_child(fresh)

	var row := DFUI.hbox(12)
	row.alignment = BoxContainer.ALIGNMENT_CENTER
	var files := BoardUI.paper_button("Case files")
	row.add_child(files)
	var menu := DFUI.button("Night desk")
	row.add_child(menu)
	v.add_child(row)
	files.pressed.connect(func() -> void: Nav.go("cases"))
	menu.pressed.connect(func() -> void: Nav.go("menu"))

	_refresh_profile(fresh)
	if result == "guilty":
		Nav.play_sfx("ui_confirm")
	else:
		Nav.play_sfx("ui_error")


func _ink_body(text: String) -> Label:
	var l := Label.new()
	l.text = text
	BoardUI.ink(l, "InkLabel")
	l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return l


func _refresh_profile(line: Label) -> void:
	var me := await ApiClient.api("GET", "/auth/me")
	if me.get("ok", false):
		var data: Dictionary = me["data"]
		Session.set_profile(data.get("profile", {}))
		var p := Session.profile
		line.text = "%s · Level %d · %d XP total" % [
			str(p.get("rank", "Rookie")), int(p.get("level", 1)), int(p.get("xp", 0))]
