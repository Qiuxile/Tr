package tr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// maxFileBytes bounds a single translation job so a huge or accidental binary
// file cannot exhaust memory (or a user's API budget).
const maxFileBytes = 16 << 20 // 16 MiB

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
	// File translation must keep indentation and leading/trailing whitespace.
	opts.PreserveWhitespace = true

	info, statErr := os.Stat(inPath)
	if statErr == nil && info.Size() > maxFileBytes {
		return fmt.Errorf("%s", T(lang, "err.file_too_large", maxFileBytes/1024/1024))
	}
	raw, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("%s", T(lang, "err.file_read", err))
	}
	if len(raw) > maxFileBytes {
		return fmt.Errorf("%s", T(lang, "err.file_too_large", maxFileBytes/1024/1024))
	}

	text, hadBOM := stripBOM(string(raw))
	if !utf8.ValidString(text) {
		return fmt.Errorf("%s", T(lang, "err.file_encoding"))
	}
	if strings.IndexByte(text, 0) >= 0 {
		return fmt.Errorf("%s", T(lang, "err.file_binary"))
	}
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
			// Authentication, rate-limit, network and empty-input errors will
			// fail for every following block too. Stop instead of multiplying
			// requests (which is especially harmful under rate limiting).
			if isFatalTranslationError(err) {
				if progress && len(blocks) > 1 {
					fmt.Fprint(os.Stderr, "\r\033[K")
				}
				return err
			}
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

	if len(blocks) > 0 && failed == len(blocks) {
		if firstErr != nil {
			return firstErr
		}
		return fmt.Errorf("%s", T(lang, "file.partial", failed, len(blocks)))
	}

	out := strings.Join(outLines, "\n")
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	if hadBOM {
		out = "\ufeff" + out
	}

	if strings.TrimSpace(outPath) == "" {
		fmt.Print(out)
		if !strings.HasSuffix(out, "\n") {
			fmt.Println()
		}
		if failed > 0 {
			return fmt.Errorf("%s", T(lang, "file.partial", failed, len(blocks)))
		}
		return nil
	}

	dest, err := resolveOutputPath(inPath, outPath)
	if err != nil {
		return fmt.Errorf("%s", T(lang, "err.file_write", err))
	}
	if err := writeFileAtomic(dest, []byte(out), 0o644); err != nil {
		return fmt.Errorf("%s", T(lang, "err.file_write", err))
	}
	fmt.Fprintln(os.Stderr, T(lang, "file.written", dest))
	if failed > 0 {
		return fmt.Errorf("%s", T(lang, "file.partial", failed, len(blocks)))
	}
	return nil
}

// translateLineBlock translates one block of lines, keeping the line count
// intact. It first sends the whole block in a single request; if the reply does
// not preserve the line structure -- or simply echoes the input back -- it
// falls back to translating line by line. Genuine API/network failures are not
// retried per line: they would just multiply the failure and the cost.
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
	if err != nil && isFatalTranslationError(err) {
		return nil, err
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
			if isFatalTranslationError(err) {
				return out, err
			}
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
//
// The end of a block is the last non-blank line, not len(lines): trailing blank
// lines must stay outside the request. Otherwise the model's reply legitimately
// has fewer lines than the block and tr wastes the request on a per-line
// fallback.
func splitBlocks(lines []string, maxChars int) []fileBlock {
	var blocks []fileBlock
	start, size, last := -1, 0, -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if start < 0 {
			start, size, last = i, 0, -1
		}
		size += len([]rune(line)) + 1
		last = i
		if size >= maxChars {
			blocks = append(blocks, fileBlock{start: start, end: last + 1})
			start, size, last = -1, 0, -1
		}
	}
	if start >= 0 {
		blocks = append(blocks, fileBlock{start: start, end: last + 1})
	}
	return blocks
}

// resolveOutputPath decides whether out is a file or a directory. An existing
// directory, or a path written with a trailing separator, receives a file named
// after the source file; anything else is treated as the target file path.
func resolveOutputPath(inPath, out string) (string, error) {
	if isDirLike(out) {
		if err := os.MkdirAll(out, 0o755); err != nil {
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
