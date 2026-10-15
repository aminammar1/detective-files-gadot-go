extends CanvasLayer

## Nav autoload: scene transitions (200ms fades per art bible §7) + audio.
## Audio lives here (rather than a 4th autoload) so music survives scene swaps.

const SCENES := {
	"startup": "res://scenes/startup.tscn",
	"auth": "res://scenes/auth.tscn",
	"menu": "res://scenes/main_menu.tscn",
	"cases": "res://scenes/case_select.tscn",
	"investigation": "res://scenes/investigation.tscn",
	"results": "res://scenes/results.tscn",
	"profile": "res://scenes/profile.tscn",
	"settings": "res://scenes/settings.tscn",
}

const SETTINGS_PATH := "user://settings.cfg"

var music_on: bool = true
var sfx_on: bool = true

var _fade: ColorRect
var _music: AudioStreamPlayer
var _navigating: bool = false


func _ready() -> void:
	layer = 100
	_fade = ColorRect.new()
	_fade.color = Color(0.07059, 0.08235, 0.11373, 1.0)  # midnight
	_fade.set_anchors_preset(Control.PRESET_FULL_RECT)
	_fade.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_fade.modulate.a = 0.0
	add_child(_fade)
	_music = AudioStreamPlayer.new()
	_music.bus = "Master"
	add_child(_music)
	_load_settings()
	Session.expired.connect(_on_session_expired)


func go(key: String) -> void:
	if not SCENES.has(key) or _navigating:
		return
	_navigating = true
	var tw := create_tween().set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_OUT)
	tw.tween_property(_fade, "modulate:a", 1.0, 0.2)
	await tw.finished
	get_tree().change_scene_to_file(SCENES[key])
	await get_tree().process_frame
	await get_tree().process_frame
	var tw2 := create_tween().set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_OUT)
	tw2.tween_property(_fade, "modulate:a", 0.0, 0.2)
	await tw2.finished
	_navigating = false


func reload() -> void:
	var cur := get_tree().current_scene
	if cur != null and cur.scene_file_path != "":
		_navigating = false
		go(_key_for(cur.scene_file_path))


func _key_for(path: String) -> String:
	for k in SCENES:
		if SCENES[k] == path:
			return k
	return "menu"


func _on_session_expired() -> void:
	go("auth")


# --- Audio (one-shots + one music bed) ---

func play_sfx(sfx_name: String) -> void:
	if not sfx_on:
		return
	var p := AudioStreamPlayer.new()
	p.stream = load(ArtPaths.SFX_DIR + sfx_name + ".ogg")
	p.finished.connect(p.queue_free)
	add_child(p)
	p.play()


func play_music(track: String) -> void:
	if not music_on:
		return
	var want := ArtPaths.MUSIC_DIR + track + ".mp3"
	if _music.playing and _music.stream != null and _music.stream.resource_path == want:
		return
	_music.stream = load(want)
	if _music.stream is AudioStreamMP3:
		(_music.stream as AudioStreamMP3).loop = true
	elif _music.stream is AudioStreamOggVorbis:
		(_music.stream as AudioStreamOggVorbis).loop = true
	_music.play()


func stop_music() -> void:
	_music.stop()


func set_audio(new_music_on: bool, new_sfx_on: bool) -> void:
	music_on = new_music_on
	sfx_on = new_sfx_on
	if not music_on:
		stop_music()
	var cfg := ConfigFile.new()
	cfg.set_value("audio", "music_on", music_on)
	cfg.set_value("audio", "sfx_on", sfx_on)
	cfg.save(SETTINGS_PATH)


func _load_settings() -> void:
	var cfg := ConfigFile.new()
	if cfg.load(SETTINGS_PATH) == OK:
		music_on = bool(cfg.get_value("audio", "music_on", true))
		sfx_on = bool(cfg.get_value("audio", "sfx_on", true))
