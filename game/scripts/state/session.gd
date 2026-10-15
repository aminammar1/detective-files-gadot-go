extends Node

## Session autoload: token + user state. Token lives in user://, never logged.
## Any 401 from ApiClient funnels through expire() -> login screen.

signal changed
signal expired

const STORE_PATH := "user://session.cfg"

var token: String = ""
var user: Dictionary = {}
var profile: Dictionary = {}

# Per-case working state for the current visit (refreshed from server).
var case_id: String = ""
var case_data: Dictionary = {}
var inv_state: Dictionary = {}
var last_result: Dictionary = {}


func _ready() -> void:
	load_session()


func is_logged_in() -> bool:
	return token != ""


func auth_headers() -> Dictionary:
	if token == "":
		return {}
	return {"Authorization": "Bearer " + token}


func save_session() -> void:
	var cfg := ConfigFile.new()
	cfg.set_value("session", "token", token)
	cfg.set_value("session", "username", user.get("username", ""))
	cfg.save(STORE_PATH)


func load_session() -> void:
	var cfg := ConfigFile.new()
	if cfg.load(STORE_PATH) == OK:
		token = str(cfg.get_value("session", "token", ""))
		var uname := str(cfg.get_value("session", "username", ""))
		if uname != "":
			user = {"username": uname}


func set_auth(new_token: String, new_user: Dictionary) -> void:
	token = new_token
	user = new_user
	save_session()
	changed.emit()


func set_profile(new_profile: Dictionary) -> void:
	profile = new_profile
	if new_profile.has("display_name"):
		user["display_name"] = new_profile["display_name"]
	changed.emit()


func clear_case() -> void:
	case_id = ""
	case_data = {}
	inv_state = {}
	last_result = {}


func logout_local() -> void:
	token = ""
	user = {}
	profile = {}
	clear_case()
	var cfg := ConfigFile.new()
	cfg.load(STORE_PATH)
	# Delete the stored token on logout (spec: client rule).
	if DirAccess.dir_exists_absolute("user://"):
		DirAccess.remove_absolute(STORE_PATH)
	changed.emit()


func expire() -> void:
	# 401 anywhere -> drop credentials, tell Nav to show login.
	var was_in := is_logged_in()
	logout_local()
	if was_in:
		expired.emit()


## Screenshot/E2E helper only: seeds plausible auth state without network.
func debug_seed() -> void:
	token = "debug-token"
	user = {"id": "debug", "email": "sam@graywater.gov", "username": "sam_spade"}
	profile = {"display_name": "Sam Spade", "rank": "Gumshoe", "level": 2, "xp": 1185}
	changed.emit()
