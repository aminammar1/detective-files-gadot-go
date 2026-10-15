extends Node

## ApiClient autoload: the single HTTPRequest-based client.
## Sequential queue (one HTTPRequest serves one call at a time).
## Never logs tokens or passwords. 401s funnel to Session.expire().

class ApiCall extends RefCounted:
	signal finished(result: Dictionary)
	var method: int
	var path: String
	var body: Dictionary
	var with_auth: bool


const DEFAULT_BASE := "http://localhost:8080"
const API_PREFIX := "/api/v1"  # /health and /ready are root-mounted (docs/API.md).

const METHODS := {
	"GET": HTTPClient.METHOD_GET,
	"POST": HTTPClient.METHOD_POST,
	"PATCH": HTTPClient.METHOD_PATCH,
	"DELETE": HTTPClient.METHOD_DELETE,
}

var base_url: String = DEFAULT_BASE
var timeout_secs: float = 15.0

var _http: HTTPRequest
var _queue: Array[ApiCall] = []
var _active: ApiCall = null


func _ready() -> void:
	var env_url := OS.get_environment("DF_API_URL").strip_edges()
	if env_url != "":
		base_url = env_url.rstrip("/")
	_http = HTTPRequest.new()
	_http.timeout = timeout_secs
	add_child(_http)
	_http.request_completed.connect(_on_completed)


func api(method: String, path: String, body: Dictionary = {}, with_auth: bool = true) -> Dictionary:
	var call := ApiCall.new()
	call.method = int(METHODS.get(method.to_upper(), HTTPClient.METHOD_GET))
	call.path = path
	call.body = body
	call.with_auth = with_auth
	_queue.append(call)
	_pump()
	return await call.finished


func health() -> Dictionary:
	# Root-mounted, no auth (docs/API.md).
	return await api("GET", "/health", {}, false)


func readiness() -> Dictionary:
	return await api("GET", "/ready", {}, false)


func _pump() -> void:
	if _active != null or _queue.is_empty():
		return
	_active = _queue.pop_front()
	var headers: PackedStringArray = ["Content-Type: application/json", "Accept: application/json"]
	if _active.with_auth:
		for k in Session.auth_headers():
			headers.append("%s: %s" % [k, Session.auth_headers()[k]])
	var payload := ""
	if not _active.body.is_empty():
		var scrubbed := _active.body.duplicate()
		# Never put secrets on the wire twice or in logs: body goes, logs don't.
		payload = JSON.stringify(scrubbed)
	var full_path := _active.path
	if full_path != "/health" and full_path != "/ready":
		full_path = API_PREFIX + full_path
	var err := _http.request(base_url + full_path, headers, _active.method, payload)
	if err != OK:
		var call := _active
		_active = null
		call.finished.emit({"ok": false, "transport": true, "code": "connection",
			"message": "The precinct isn't answering. Check your connection and try again."})
		_pump()


func _on_completed(_result: int, status: int, _headers: PackedStringArray, raw: PackedByteArray) -> void:
	var call := _active
	_active = null
	if call == null:
		_pump()
		return
	var out: Dictionary
	if status == 0:
		out = {"ok": false, "transport": true, "code": "connection",
			"message": "The precinct isn't answering. Check your connection and try again."}
	else:
		var parsed: Variant = null
		if raw.size() > 0:
			parsed = JSON.parse_string(raw.get_string_from_utf8())
		if status == 401:
			Session.expire()
			out = {"ok": false, "status": status, "code": "unauthorized",
				"message": "Your session went cold. Sign in again."}
		elif status >= 200 and status < 300:
			out = {"ok": true, "status": status, "data": parsed}
		else:
			var emsg := "Something went wrong at the precinct."
			var ecode := "internal"
			if parsed is Dictionary and (parsed as Dictionary).has("error"):
				var e: Dictionary = (parsed as Dictionary)["error"]
				ecode = str(e.get("code", ecode))
				emsg = str(e.get("message", emsg))
			out = {"ok": false, "status": status, "code": ecode, "message": emsg}
	call.finished.emit(out)
	_pump()
