package tr

import "testing"

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		mode      Mode
		offline   bool
		fileIn    string
		fileOut   string
		remaining []string
		parseErr  bool
	}{
		{"file with offline before", []string{"-f", "-o", "file.txt"}, ModeFile, true, "file.txt", "", nil, false},
		{"file with arrow", []string{"-f", "file.txt", "->", "en"}, ModeFile, false, "file.txt", "", []string{"->", "en"}, false},
		{"file named config", []string{"-f", "config", "show"}, ModeFile, false, "config", "show", nil, false},
		{"file named cache", []string{"-f", "cache", "clear"}, ModeFile, false, "cache", "clear", nil, false},
		{"file with output and arrow", []string{"-f", "file.txt", "out.txt", "->", "en"}, ModeFile, false, "file.txt", "out.txt", []string{"->", "en"}, false},
		{"file with extra positional", []string{"-f", "file.txt", "out.txt", "extra"}, ModeFile, false, "file.txt", "out.txt", nil, true},
		{"dashdash config is text", []string{"--", "config"}, ModeTranslate, false, "", "", []string{"config"}, false},
		{"dashdash tui is text", []string{"--", "tui"}, ModeTranslate, false, "", "", []string{"tui"}, false},
		{"config subcommand", []string{"config", "show"}, ModeConfig, false, "", "", nil, false},
		{"bare config is text", []string{"config"}, ModeTranslate, false, "", "", []string{"config"}, false},
		{"tui", []string{"tui"}, ModeTUI, false, "", "", nil, false},
		{"cache subcommand", []string{"cache", "clear"}, ModeCache, false, "", "", nil, false},
		{"unknown flag", []string{"hello", "--bogus"}, ModeTranslate, false, "", "", []string{"hello"}, true},
		{"help wins over subcommand", []string{"-help", "config", "show"}, ModeHelp, false, "", "", nil, false},
		{"legacy config", []string{"-config", "show"}, ModeConfig, false, "", "", nil, false},
		{"offline only", []string{"-o"}, ModeTranslate, true, "", "", nil, false},
		{"missing file", []string{"-f"}, ModeFile, false, "", "", nil, false},
		{"dashdash flag is text", []string{"--", "-o"}, ModeTranslate, false, "", "", []string{"-o"}, false},
		{"arrow kept", []string{"hello", "->", "ja"}, ModeTranslate, false, "", "", []string{"hello", "->", "ja"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, remaining := parseArgs("en", tc.args)
			if opts.Mode != tc.mode {
				t.Fatalf("mode = %v, want %v", opts.Mode, tc.mode)
			}
			if opts.Offline != tc.offline {
				t.Fatalf("offline = %v, want %v", opts.Offline, tc.offline)
			}
			if opts.FileIn != tc.fileIn || opts.FileOut != tc.fileOut {
				t.Fatalf("file = %q -> %q, want %q -> %q", opts.FileIn, opts.FileOut, tc.fileIn, tc.fileOut)
			}
			if opts.ParseError != tc.parseErr {
				t.Fatalf("ParseError = %v, want %v", opts.ParseError, tc.parseErr)
			}
			if len(remaining) != len(tc.remaining) {
				t.Fatalf("remaining = %q, want %q", remaining, tc.remaining)
			}
			for i := range remaining {
				if remaining[i] != tc.remaining[i] {
					t.Fatalf("remaining = %q, want %q", remaining, tc.remaining)
				}
			}
		})
	}
}
