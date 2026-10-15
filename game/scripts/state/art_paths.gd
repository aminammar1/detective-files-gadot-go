class_name ArtPaths
extends RefCounted

## One-address book for every art slot (art bible §9).
## Swapping a placeholder for a real file = change one const here.

const FONT_DISPLAY := "res://assets/fonts/SpecialElite-Regular.ttf"
const FONT_BODY := "res://assets/fonts/IBMPlexSans-Variable.ttf"

const ICONS_DIR := "res://assets/icons/"
const SFX_DIR := "res://assets/audio/"
const MUSIC_DIR := "res://assets/audio/music/"

# Palette mirrors (art bible §1), for code-side tinting only.
const PRIMARY := Color(0.91373, 0.88627, 0.82353)
const SECONDARY := Color(0.69020, 0.68235, 0.74510)
const DISABLED := Color(0.43137, 0.44314, 0.50196)
const GOLD := Color(0.85098, 0.64314, 0.25490)
const GOLD_BRIGHT := Color(0.90980, 0.74510, 0.35294)
const PURPLE := Color(0.54510, 0.48235, 0.78039)
const ERROR := Color(0.87843, 0.47843, 0.41961)
const SUCCESS := Color(0.56078, 0.75294, 0.47843)

static var _icon_cache: Dictionary = {}


## White Kenney glyph tinted to a palette color. Cached.
static func icon(icon_name: String, color: Color = SECONDARY) -> Texture2D:
	var key := "%s|%s" % [icon_name, color.to_html()]
	if _icon_cache.has(key):
		return _icon_cache[key]
	var path := ICONS_DIR + icon_name + ".png"
	var tex: Texture2D = load(path)
	if tex == null:
		return null
	var img := tex.get_image()
	var data := img.get_data()
	# Multiply RGB by tint, keep alpha. Images are RGBA8.
	for i in range(0, data.size(), 4):
		data[i] = int(data[i] * color.r)
		data[i + 1] = int(data[i + 1] * color.g)
		data[i + 2] = int(data[i + 2] * color.b)
	var out := Image.create_from_data(img.get_width(), img.get_height(), img.has_mipmaps(), img.get_format(), data)
	var tinted := ImageTexture.create_from_image(out)
	_icon_cache[key] = tinted
	return tinted


static func bg_for(asset: String) -> String:
	# Case JSON asset refs are relative to assets/; real files land here,
	# placeholders resolve to "" so screens draw monogram tiles instead.
	if asset == "" or asset.begins_with("placeholder:"):
		return ""
	return "res://assets/" + asset
