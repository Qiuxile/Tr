package tr

import "strings"

// iso639_1 lists the language codes recognized by the "-> <lang>" arrow
// syntax. Translating is done by the selected backend, so any code the
// backend understands can be passed through; this set only gates what the
// CLI treats as an arrow target instead of ordinary text.
var iso639_1 = map[string]bool{
	"aa": true, "ab": true, "ae": true, "af": true, "ak": true, "am": true, "an": true,
	"ar": true, "as": true, "av": true, "ay": true, "az": true, "ba": true, "be": true,
	"bg": true, "bh": true, "bi": true, "bm": true, "bn": true, "bo": true, "br": true,
	"bs": true, "ca": true, "ce": true, "ch": true, "co": true, "cr": true, "cs": true,
	"cu": true, "cv": true, "cy": true, "da": true, "de": true, "dv": true, "dz": true,
	"ee": true, "el": true, "en": true, "eo": true, "es": true, "et": true, "eu": true,
	"fa": true, "ff": true, "fi": true, "fj": true, "fo": true, "fr": true, "fy": true,
	"ga": true, "gd": true, "gl": true, "gn": true, "gu": true, "gv": true, "ha": true,
	"he": true, "hi": true, "ho": true, "hr": true, "ht": true, "hu": true, "hy": true,
	"hz": true, "ia": true, "id": true, "ie": true, "ig": true, "ii": true, "ik": true,
	"io": true, "is": true, "it": true, "iu": true, "ja": true, "jv": true, "ka": true,
	"kg": true, "ki": true, "kj": true, "kk": true, "kl": true, "km": true, "kn": true,
	"ko": true, "kr": true, "ks": true, "ku": true, "kv": true, "kw": true, "ky": true,
	"la": true, "lb": true, "lg": true, "li": true, "ln": true, "lo": true, "lt": true,
	"lu": true, "lv": true, "mg": true, "mh": true, "mi": true, "mk": true, "ml": true,
	"mn": true, "mr": true, "ms": true, "mt": true, "my": true, "na": true, "nb": true,
	"nd": true, "ne": true, "ng": true, "nl": true, "nn": true, "no": true, "nr": true,
	"nv": true, "ny": true, "oc": true, "oj": true, "om": true, "or": true, "os": true,
	"pa": true, "pi": true, "pl": true, "ps": true, "pt": true, "qu": true, "rm": true,
	"rn": true, "ro": true, "ru": true, "rw": true, "sa": true, "sc": true, "sd": true,
	"se": true, "sg": true, "si": true, "sk": true, "sl": true, "sm": true, "sn": true,
	"so": true, "sq": true, "sr": true, "ss": true, "st": true, "su": true, "sv": true,
	"sw": true, "ta": true, "te": true, "tg": true, "th": true, "ti": true, "tk": true,
	"tl": true, "tn": true, "to": true, "tr": true, "ts": true, "tt": true, "tw": true,
	"ty": true, "ug": true, "uk": true, "ur": true, "uz": true, "ve": true, "vi": true,
	"vo": true, "wa": true, "wo": true, "xh": true, "yi": true, "yo": true, "za": true,
	"zh": true, "zu": true,
}

// langAliases maps human-friendly names to ISO 639-1 codes so the arrow
// syntax accepts things like "tr 你好 -> 英文" or "tr hello -> 日本語".
var langAliases = map[string]string{
	// codes with regional variants
	"zh-cn": "zh", "zh-tw": "zh", "zh-hk": "zh", "zh-sg": "zh", "zh-hans": "zh", "zh-hant": "zh",
	// English
	"en": "en", "english": "en", "英文": "en", "英": "en",
	// Chinese
	"zh": "zh", "chinese": "zh", "中文": "zh", "汉语": "zh", "简体": "zh", "中": "zh",
	// Japanese
	"ja": "ja", "jp": "ja", "jpn": "ja", "japanese": "ja", "日本語": "ja", "日文": "ja", "日語": "ja", "日": "ja",
	// Korean
	"ko": "ko", "kr": "ko", "kor": "ko", "korean": "ko", "한국어": "ko", "韩文": "ko", "韓文": "ko", "韩": "ko",
	// Russian
	"ru": "ru", "rus": "ru", "russian": "ru", "русский": "ru", "俄文": "ru", "俄語": "ru", "俄": "ru",
	// French
	"fr": "fr", "fra": "fr", "french": "fr", "français": "fr", "francais": "fr", "法文": "fr", "法語": "fr", "法": "fr",
	// German
	"de": "de", "ger": "de", "deu": "de", "german": "de", "deutsch": "de", "德文": "de", "德语": "de", "德語": "de", "德": "de",
	// Spanish
	"es": "es", "spa": "es", "spanish": "es", "español": "es", "espanol": "es", "西文": "es", "西班牙语": "es", "西": "es",
	// Italian
	"it": "it", "ita": "it", "italian": "it", "italiano": "it", "意文": "it", "意大利语": "it", "意": "it",
	// Portuguese
	"pt": "pt", "por": "pt", "portuguese": "pt", "português": "pt", "portugues": "pt", "葡文": "pt", "葡萄牙语": "pt", "葡": "pt",
	// Arabic
	"ar": "ar", "ara": "ar", "arabic": "ar", "العربية": "ar", "阿拉伯语": "ar", "阿": "ar",
	// Thai
	"th": "th", "tha": "th", "thai": "th", "ไทย": "th", "泰文": "th", "泰语": "th", "泰": "th",
	// Hindi
	"hi": "hi", "hin": "hi", "hindi": "hi", "हिन्दी": "hi", "印地语": "hi",
	// Vietnamese
	"vi": "vi", "vie": "vi", "vietnamese": "vi", "tiếng việt": "vi", "越南语": "vi", "越": "vi",
}

// normalizeLangArg canonicalizes a user-supplied language argument
// (code or human-friendly name) to an ISO 639-1 code. Returns
// ("", false) when the argument is not recognized.
func normalizeLangArg(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if code, ok := langAliases[s]; ok {
		return code, true
	}
	if iso639_1[s] {
		return s, true
	}
	return "", false
}

// DetectLang guesses the language of text from its writing system using
// Unicode script blocks. It returns an ISO 639-1 code, falling back to
// "en" when the text has no distinctive script (or is Latin-script text).
func DetectLang(text string) string {
	var kana, han, hangul, cyr, arab, greek, thai, deva int
	for _, r := range text {
		switch {
		case r >= 0x3040 && r <= 0x30FF, r == 0x30FC, r >= 0x31F0 && r <= 0x31FF: // hiragana / katakana
			kana++
		case r >= 0x4E00 && r <= 0x9FFF, r >= 0x3400 && r <= 0x4DBF: // CJK unified ideographs
			han++
		case r >= 0xAC00 && r <= 0xD7A3, r >= 0x1100 && r <= 0x11FF, r >= 0x3130 && r <= 0x318F: // hangul
			hangul++
		case r >= 0x0400 && r <= 0x052F: // Cyrillic
			cyr++
		case r >= 0x0600 && r <= 0x06FF: // Arabic
			arab++
		case r >= 0x0370 && r <= 0x03FF: // Greek
			greek++
		case r >= 0x0E00 && r <= 0x0E7F: // Thai
			thai++
		case r >= 0x0900 && r <= 0x097F: // Devanagari
			deva++
		}
	}

	// Any kana at all strongly suggests Japanese, even among kanji.
	if kana > 0 {
		return "ja"
	}

	best, bestN := "en", 0
	for code, n := range map[string]int{
		"zh": han, "ko": hangul, "ru": cyr, "ar": arab, "el": greek, "th": thai, "hi": deva,
	} {
		if n > bestN {
			best, bestN = code, n
		}
	}
	return best
}
