package tr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// translateWithMyMemory calls the free MyMemory translation API.
func translateWithMyMemory(text, sourceLang, targetLang, lang string) (string, error) {
	apiURL := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=%s|%s",
		url.QueryEscape(text), sourceLang, targetLang)

	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrNetworkFailed, T(lang, "api.mm_request"), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrNetworkFailed, T(lang, "api.mm_status", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrNetworkFailed, T(lang, "api.mm_read"), err)
	}

	var result struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("%s: %w", T(lang, "api.mm_parse"), err)
	}
	if result.ResponseData.TranslatedText == "" {
		return "", fmt.Errorf(T(lang, "api.mm_empty", text))
	}
	return result.ResponseData.TranslatedText, nil
}
