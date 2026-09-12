package tr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// fileBlock is a half-open range of line indexes.
type fileBlock struct {
	start int // inclusive
	end   int // exclusive
}

// TranslateFile translates a text file.
//
// When outPath is empty the translation goes to stdout; otherwise it is written
// to outPath, which may be either a file path or a directory (the directory
// receives a file named after the source file). target is the per-call target
// language: empty means "use the configured one", set means "use exactly this"
// (see Translate).
//
// The file is translated in blocks so that long documents neither exceed the
// model context nor spend tokens on a single oversized request; each block
// still goes through the normal pipeline, so caching, offline mode and
// per-call language overrides all apply.
func TranslateFile(cfg Config, opts Options, cache *Cache, inPath, outPath, target string) error {
	lang := cfg.UILang

	raw, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("%s", T(lang, "err.file_read", err))
	}
	text, hadBOM := stripBOM(string(raw))
	if strings.TrimSpace(text) == "" {
		return ErrEmptyInput
	}
	crlf := strings.Contains(text, "\r\n")
	text = strings.ReplaceAll(text, "\r\n", "\n")

	lines := strings.Split(text, "\n")
	outLines := make([]string, len(lines))
	copy(outLines, lines)

	blocks := splitBlocks(lines, aiRequestChars)
	failed := 0
	var firstErr error
	progress := isStderrTerminal()

	for i, b := range blocks {
		if progress && len(blocks) > 1 {
			fmt.Fprintf(os.Stderr, "\r%s", T(lang, "file.progress", i+1, len(blocks)))
		}
		translated, err := translateLineBlock(cfg, opts, cache, lines[b.start:b.end], target)
		if err != nil {
			failed++
			if firstErr == nil {
				firstErr = err
			}
			continue // keep the original lines for this block
		}
		copy(outLines[b.start:b.end], translated)
	}
	if progress && len(blocks) > 1 {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	if failed == len(blocks) {
		return firstErr
	}
	if failed > 0 {
		fmt.Fprintln(os.Stderr, T(lang, "file.partial", failed, len(blocks)))
	}

	out := strings.Join(outLines, "\n")
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	if hadBOM {
		out = "\ufeff" + out
	}

	if strings.TrimSpace(outPath) == "" {
		fmt.Println(out)
		return nil
	}

	dest, err := resolveOutputPath(inPath, outPath)
	if err != nil {
		return fmt.Errorf("%s", T(lang, "err.file_write", err))
	}
	if dir := filepath.Dir(dest); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%s", T(lang, "err.file_write", err))
		}
	}
	if err := os.WriteFile(dest, []byte(out), 0644); err != nil {
		return fmt.Errorf("%s", T(lang, "err.file_write", err))
	}
	fmt.Fprintln(os.Stderr, T(lang, "file.written", dest))
	return nil
}

// translateLineBlock translates one block of lines, keeping the line count
// intact. It first sends the whole block in a single request; if the reply does
// not preserve the line structure — or simply echoes the input back, which the
// free backends do for some multi-line requests — it falls back to translating
// line by line.
func translateLineBlock(cfg Config, opts Options, cache *Cache, lines []string, target string) ([]string, error) {
	if len(lines) == 1 {
		return translateLinesIndividually(cfg, opts, cache, lines, target)
	}

	blockText := strings.Join(lines, "\n")
	res, err := Translate(cfg, opts, cache, blockText, target)
	if err == nil {
		translated := strings.Split(res.Text, "\n")
		if len(translated) == len(lines) && strings.Join(translated, "\n") != blockText {
			return alignBlankLines(lines, translated), nil
		}
	}
	return translateLinesIndividually(cfg, opts, cache, lines, target)
}

// translateLinesIndividually translates non-blank lines one at a time and keeps
// blank lines (the paragraph structure) exactly as they were.
func translateLinesIndividually(cfg Config, opts Options, cache *Cache, lines []string, target string) ([]string, error) {
	out := make([]string, len(lines))
	var firstErr error

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			out[i] = line
			continue
		}
		res, err := Translate(cfg, opts, cache, line, target)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			out[i] = line // keep the original line
			continue
		}
		out[i] = res.Text
	}
	if firstErr != nil && allLinesUnchanged(lines, out) {
		return out, firstErr
	}
	return out, nil
}

// alignBlankLines forces blank lines of the source to stay blank in the result.
func alignBlankLines(source, translated []string) []string {
	for i := range translated {
		if i < len(source) && strings.TrimSpace(source[i]) == "" {
			translated[i] = ""
		}
	}
	return translated
}

// allLinesUnchanged reports whether nothing was translated at all.
func allLinesUnchanged(source, out []string) bool {
	for i := range source {
		if strings.TrimSpace(source[i]) != "" && source[i] != out[i] {
			return false
		}
	}
	return true
}

// splitBlocks groups lines into blocks of at most maxChars characters. Blank
// lines carry paragraph structure rather than content, so they never start or
// split a block; they travel inside one.
func splitBlocks(lines []string, maxChars int) []fileBlock {
	var blocks []fileBlock
	start, size := -1, 0

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if start < 0 {
			start, size = i, 0
		}
		size += len([]rune(line)) + 1
		if size >= maxChars {
			blocks = append(blocks, fileBlock{start: start, end: i + 1})
			start, size = -1, 0
		}
	}
	if start >= 0 {
		blocks = append(blocks, fileBlock{start: start, end: len(lines)})
	}
	return blocks
}

// resolveOutputPath decides whether out is a file or a directory. An existing
// directory, or a path written with a trailing separator, receives a file named
// after the source file; anything else is treated as the target file path.
func resolveOutputPath(inPath, out string) (string, error) {
	if isDirLike(out) {
		if err := os.MkdirAll(out, 0755); err != nil {
			return "", err
		}
		return filepath.Join(out, filepath.Base(inPath)), nil
	}
	return out, nil
}

// isDirLike reports whether a path should be treated as a directory.
func isDirLike(p string) bool {
	if strings.HasSuffix(p, "/") || strings.HasSuffix(p, `\`) {
		return true
	}
	if st, err := os.Stat(p); err == nil && st.IsDir() {
		return true
	}
	return false
}

// stripBOM removes a UTF-8 byte order mark, reporting whether one was present.
func stripBOM(s string) (string, bool) {
	if strings.HasPrefix(s, "\ufeff") {
		return strings.TrimPrefix(s, "\ufeff"), true
	}
	return s, false
}

// isStderrTerminal reports whether stderr is an interactive terminal, used to
// decide if progress output is welcome.
func isStderrTerminal() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}
