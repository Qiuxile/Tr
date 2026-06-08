package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var version string = "1.0.0"

// config 全局配置
var config = map[string]string{
	"source_lang": "en",
	"target_lang": "zh",
	"api_url":     "None", // "None" 表示使用 MyMemory，否则使用自定义 API
}

func getConfigPath() string {
	// 获取用户配置目录（Windows: C:\Users\<用户名>\AppData\Roaming\Tr\）
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		// 降级：当前目录下的 config.json
		return "config.json"
	}
	trDir := filepath.Join(userConfigDir, "Tr")
	// 确保目录存在
	_ = os.MkdirAll(trDir, 0755)
	return filepath.Join(trDir, "config.json")
}

// saveConfig 保存配置到文件
func saveConfig() error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	// 修正：调用 getConfigPath() 获取路径字符串
	return os.WriteFile(getConfigPath(), data, 0644)
}

// loadConfig 从文件加载配置
func loadConfig() error {
	// 修正：调用 getConfigPath() 获取路径字符串
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，创建默认配置文件
			if err := saveConfig(); err != nil {
				return fmt.Errorf("创建默认配置文件失败: %w", err)
			}
			return nil // 创建成功，无需再解析
		}
		return err
	}
	return json.Unmarshal(data, &config)
}

// runConfigCommand 处理 -config 子命令
func runConfigCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}
	subCmd := args[0]
	switch subCmd {
	case "show":
		fmt.Println("当前配置:")
		for k, v := range config {
			fmt.Printf("| %s\t =\t %s\t |\n", k, v)
		}
		fmt.Println("默认使用 MyMemory 进行 英=>中 翻译")
		return true
	case "set":
		if len(args) != 3 {
			fmt.Println("用法: -config set <key> <value>")
			fmt.Println("可用的 key: source_lang, target_lang, api_url")
			return true
		}
		key := args[1]
		val := args[2]
		if _, ok := config[key]; !ok {
			fmt.Printf("无效的配置项: %s，可用项: source_lang, target_lang, api_url\n", key)
			return true
		}
		config[key] = val
		if err := saveConfig(); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
		} else {
			fmt.Printf("设置成功: %s = %s\n", key, val)
		}
		return true
	default:
		fmt.Printf("未知的子命令: %s，可用: show, set\n", subCmd)
		return true
	}
}

// MyMemory 翻译请求
func translateWithMyMemory(text, sourceLang, targetLang string) (string, error) {
	apiURL := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=%s|%s",
		url.QueryEscape(text), sourceLang, targetLang)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ResponseData.TranslatedText == "" {
		return "", fmt.Errorf("翻译结果为空")
	}
	return result.ResponseData.TranslatedText, nil
}

// 通用 API 翻译（兼容 LibreTranslate/DeepLX 格式）
func translateWithGenericAPI(text, sourceLang, targetLang, apiURL string) (string, error) {
	reqBody := map[string]string{
		"q":      text,
		"source": sourceLang,
		"target": targetLang,
		"format": "text",
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 尝试 LibreTranslate 格式
	var resultLibre struct {
		TranslatedText string `json:"translatedText"`
	}
	// 尝试 DeepLX 格式
	var resultDeepLX struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}

	if err := json.Unmarshal(body, &resultLibre); err == nil && resultLibre.TranslatedText != "" {
		return resultLibre.TranslatedText, nil
	}
	if err := json.Unmarshal(body, &resultDeepLX); err == nil && resultDeepLX.Code == 200 && resultDeepLX.Data != "" {
		return resultDeepLX.Data, nil
	}
	return "", fmt.Errorf("无法解析 API 响应: %s", string(body))
}

// translate 根据配置翻译文本
func translate(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("文本不能为空")
	}

	source := config["source_lang"]
	target := config["target_lang"]
	apiURL := config["api_url"]

	if apiURL == "None" {
		return translateWithMyMemory(text, source, target)
	}
	return translateWithGenericAPI(text, source, target, apiURL)
}

func help() {
	fmt.Println("用法:")
	fmt.Println("| tr [文本]   # 翻译文本并输出翻译结果")
	fmt.Println("| tr -config show          # 显示当前配置")
	fmt.Println("| tr -config set <key> <value>   # 修改配置")
	fmt.Println("| tr -help   # 获得帮助")
	fmt.Println("| tr -about    # 关于")
	fmt.Println("| tr -version  # 版本")
	fmt.Println("其他写法:")
	fmt.Println("| -config  =  --config  =  -c")
	fmt.Println("| -help	=  --help    =  -h")
	fmt.Println("| -version =  --version =  -v")
	fmt.Println("| -about   =  --about   =  -a")
}

func about() {
	fmt.Printf("Tr version: %s", version)
	fmt.Println("是由Surile开发的一个运行在终端的翻译App")
	fmt.Println("采用 MIT LICENSE")
	fmt.Println("Github 仓库链接为:")
	fmt.Println("https://github.com/Qiuxile/Tr")
}

func getVersion() {
	fmt.Printf("Tr version: %s\n", version)
}

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v，使用默认配置\n", err)
	}

	// 处理子命令
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "-config", "--config", "-c":
			runConfigCommand(os.Args[2:])
			return
		case "-help", "--help", "-h":
			help()
			return
		case "-version", "--version", "-v":
			getVersion()
			return
		case "-about", "--about", "-a":
			about()
			return
		}
	}

	// 默认：翻译传入的所有参数
	args := os.Args[1:]
	content := strings.Join(args, " ")
	translated, err := translate(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "翻译失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(translated)
}
