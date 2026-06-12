package tr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// translateWithGenericAPI calls a LibreTranslate/DeepLX-compatible translation API.
func translateWithGenericAPI(text, sourceLang, targetLang, apiURL, lang string) (string, error) {
	reqBody := map[string]string{
		"q":      text,
		"source": sourceLang,
		"target": targetLang,
		"format": "text",
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrNetworkFailed, T(lang, "api.gen_request"), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrNetworkFailed, T(lang, "api.gen_status", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrNetworkFailed, T(lang, "api.gen_read"), err)
	}

	// Try LibreTranslate format first
	var libre struct {
		TranslatedText string `json:"translatedText"`
	}
	if err := json.Unmarshal(body, &libre); err == nil && libre.TranslatedText != "" {
		return libre.TranslatedText, nil
	}

	// Try DeepLX format
	var deepLX struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &deepLX); err == nil && deepLX.Code == 200 && deepLX.Data != "" {
		return deepLX.Data, nil
	}

	return "", fmt.Errorf(T(lang, "api.gen_format", string(body)))
}
