extends Control

## Visual-verification rig. Loads a scene path from the --scene= user arg,
## optionally seeds a session (--mock) or plays a full API flow (--flow=menu,
## --flow=case, --flow=board), waits, saves the viewport PNG to /tmp/dfshots/.
##
## Usage:
##   xvfb-run -a godot --rendering-driver opengl3 --path game \
##     res://tools/screenshot.tscn -- --scene=res://scenes/auth.tscn [--mock] \
##     [--shot=name] [--flow=menu|case|board]

var _shot := "shot"
var _scene := ""


func _ready() -> void:
	var args := OS.get_cmdline_user_args()
	var mock := false
	var flow := ""
	for a in args:
		if a.begins_with("--scene="):
			_scene = a.trim_prefix("--scene=")
		elif a == "--mock":
			mock = true
		elif a.begins_with("--shot="):
			_shot = a.trim_prefix("--shot=")
		elif a.begins_with("--flow="):
			flow = a.trim_prefix("--flow=")
	if _scene == "":
		push_error("screenshot: missing --scene=res://scenes/<name>.tscn")
		get_tree().quit(2)
		return
	if mock or flow != "":
		Session.debug_seed()
	if flow != "":
		await _run_flow(flow)
	var packed: PackedScene = load(_scene)
	if packed == null:
		push_error("screenshot: cannot load " + _scene)
		get_tree().quit(2)
		return
	add_child(packed.instantiate())
	var inst := get_child(-1)
	var pending_view := ""
	var pending_suspect := ""
	var pending_clip := false
	for a in args:
		if a.begins_with("--view="):
			pending_view = a.trim_prefix("--view=")
		elif a.begins_with("--suspect="):
			pending_suspect = a.trim_prefix("--suspect=")
		elif a == "--clip":
			pending_clip = true
	# Let HTTP-less screens settle; flows already awaited their API.
	await get_tree().process_frame
	await get_tree().process_frame
	await get_tree().create_timer(0.6).timeout
	# Apply after the screen's own boot (which ends on its default view).
	if pending_suspect != "" and inst.has_method("_on_pick_suspect"):
		inst._on_pick_suspect(pending_suspect)
	elif pending_view != "" and inst.has_method("_show_tab"):
		inst._show_tab(int(pending_view))
	elif pending_view != "" and inst.has_method("_show"):
		inst._show(pending_view)
	if pending_clip and inst.has_method("_toggle_clip"):
		inst._toggle_clip()
	await get_tree().process_frame
	await get_tree().process_frame
	var img := get_viewport().get_texture().get_image()
	var dir := "/tmp/dfshots"
	DirAccess.make_dir_recursive_absolute(dir)
	var path := "%s/%s.png" % [dir, _shot]
	img.save_png(path)
	print("screenshot: saved ", path)
	get_tree().quit()


func _run_flow(flow: String) -> void:
	# Real backend flows so screens render populated (not just empty states).
	# Registers a throwaway badge (dev DB only) to get a valid token.
	var stamp := str(int(Time.get_unix_time_from_system())) + str(randi() % 997)
	var reg := await ApiClient.api("POST", "/auth/register", {
		"email": "shot_%s@example.com" % stamp,
		"username": "shot_%s" % stamp,
		"password": "trenchcoat99",
	}, false)
	if not reg.get("ok", false):
		return
	var rd: Dictionary = reg["data"]
	Session.set_auth(str(rd.get("token", "")), rd.get("user", {}))
	if flow == "menu":
		var me := await ApiClient.api("GET", "/auth/me")
		if me.get("ok", false):
			Session.set_profile((me["data"] as Dictionary).get("profile", {}))
		return
	# case/board: open the file, start, sweep every location, break Tommy.
	var detail := await ApiClient.api("GET", "/cases/the_last_train")
	if not detail.get("ok", false):
		return
	Session.case_id = "the_last_train"
	Session.case_data = detail["data"]
	await ApiClient.api("POST", "/cases/the_last_train/start")
	for loc in Session.case_data.get("locations", []):
		await ApiClient.api("POST", "/cases/the_last_train/actions",
			{"action": "search", "location_id": str(loc["id"])})
	if flow == "board" or flow == "verdict":
		await ApiClient.api("POST", "/cases/the_last_train/present-evidence",
			{"suspect_id": "tommy_beck", "evidence_id": "bond_stub"})
		await ApiClient.api("POST", "/cases/the_last_train/present-evidence",
			{"suspect_id": "vivienne_hart", "evidence_id": "matchbook"})
	var state := await ApiClient.api("GET", "/cases/the_last_train/state")
	if state.get("ok", false):
		Session.inv_state = state["data"]
	if flow == "tunis":
		await ApiClient.api("POST", "/cases/the_last_train/start")
		await ApiClient.api("POST", "/cases/the_last_train/accusations", {"suspect_id": "tommy_beck"})
		var detail2 := await ApiClient.api("GET", "/cases/the_tunis_ledger")
		if detail2.get("ok", false):
			Session.case_id = "the_tunis_ledger"
			Session.case_data = detail2["data"]
			await ApiClient.api("POST", "/cases/the_tunis_ledger/start")
			for loc in Session.case_data.get("locations", []):
				await ApiClient.api("POST", "/cases/the_tunis_ledger/actions",
					{"action": "search", "location_id": str(loc["id"])})
			await ApiClient.api("POST", "/cases/the_tunis_ledger/present-evidence",
				{"suspect_id": "bilal_haddad", "evidence_id": "banknote"})
			await ApiClient.api("POST", "/cases/the_tunis_ledger/present-evidence",
				{"suspect_id": "yasmine_benammar", "evidence_id": "stub"})
			var st3 := await ApiClient.api("GET", "/cases/the_tunis_ledger/state")
			if st3.get("ok", false):
				Session.inv_state = st3["data"]
	if flow == "verdict":
		for pair in [["edwin_marsh", "timetable"], ["edwin_marsh", "medicine_log"],
				["edwin_marsh", "key_registry"]]:
			await ApiClient.api("POST", "/cases/the_last_train/present-evidence",
				{"suspect_id": pair[0], "evidence_id": pair[1]})
		var acc := await ApiClient.api("POST", "/cases/the_last_train/accusations",
			{"suspect_id": "edwin_marsh"})
		if acc.get("ok", false):
			Session.last_result = acc["data"]
			var st2 := await ApiClient.api("GET", "/cases/the_last_train/state")
			if st2.get("ok", false):
				Session.inv_state = st2["data"]
		_scene = "res://scenes/results.tscn"
