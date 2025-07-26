package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RequestLog struct {
	Method string            `json:"method"`
	URL    string            `json:"url"`
	Header map[string]string `json:"header"`
	Get    map[string]string `json:"get"`
	Post   map[string]string `json:"post,omitempty"`
	Data   string            `json:"data,omitempty"`
}

var (
	logDir string = "logs"
)

func createLogDir() (string, error) {
	baseDir := filepath.Join(logDir, time.Now().Format("20060102-150405"))
	num := 1
	
	for {
		dirLog := fmt.Sprintf("%s-%03d", baseDir, num)
		if _, err := os.Stat(dirLog); os.IsNotExist(err) {
			err := os.MkdirAll(dirLog, 0777)
			return dirLog, err
		}
		num++
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// CORS ヘッダーを設定
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
	
	// OPTIONSリクエスト（プリフライト）に対する処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	dirLog, err := createLogDir()
	if err != nil {
		http.Error(w, "ディレクトリ作成に失敗しました", http.StatusInternalServerError)
		return
	}
	
	var body []byte
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "リクエストボディの読み取りに失敗しました", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
	}
	
	// URLを安全に取得
	requestURL := r.RequestURI
	if requestURL == "" && r.URL != nil {
		requestURL = r.URL.String()
	}
	
	request := RequestLog{
		Method: r.Method,
		URL:    requestURL,
		Header: make(map[string]string),
		Get:    make(map[string]string),
		Post:   make(map[string]string),
		Data:   string(body),
	}
	
	for name, values := range r.Header {
		if len(values) > 0 {
			// ヘッダー名を正規化してJSONキーとして安全にする
			cleanName := strings.ReplaceAll(name, "\"", "")
			cleanName = strings.ReplaceAll(cleanName, "\\", "")
			cleanName = strings.TrimSpace(cleanName)
			
			// ヘッダー値も正規化
			cleanValue := strings.ReplaceAll(values[0], "\"", "\\\"")
			cleanValue = strings.TrimSpace(cleanValue)
			
			if cleanName != "" {
				request.Header[cleanName] = cleanValue
			}
		}
	}
	
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			cleanKey := strings.TrimSpace(key)
			cleanValue := strings.TrimSpace(values[0])
			if cleanKey != "" {
				request.Get[cleanKey] = cleanValue
			}
		}
	}
	
	if r.Method == "POST" {
		if err := r.ParseForm(); err == nil {
			for key, values := range r.PostForm {
				if len(values) > 0 {
					cleanKey := strings.TrimSpace(key)
					cleanValue := strings.TrimSpace(values[0])
					if cleanKey != "" {
						request.Post[cleanKey] = cleanValue
					}
				}
			}
		}
	}
	
	logData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		http.Error(w, "ログデータのシリアライズに失敗しました", http.StatusInternalServerError)
		return
	}
	
	logFile := filepath.Join(dirLog, "log.txt")
	if err := os.WriteFile(logFile, logData, 0644); err != nil {
		http.Error(w, "ログファイルの書き込みに失敗しました", http.StatusInternalServerError)
		return
	}
	
	if len(body) > 0 {
		dataFile := filepath.Join(dirLog, "data.txt")
		if err := os.WriteFile(dataFile, body, 0644); err != nil {
			http.Error(w, "データファイルの書き込みに失敗しました", http.StatusInternalServerError)
			return
		}
	}
	
	// レスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	// JSON形式でレスポンスを返す
	response := map[string]string{
		"status":  "success",
		"message": "ログが保存されました",
		"logDir":  dirLog,
	}
	
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	
	w.Write(jsonResponse)
}

func main() {
	var (
		port     = flag.String("port", "8080", "サーバーのポート番号")
		logPath  = flag.String("logdir", "logs", "ログファイルの保存ディレクトリ")
		help     = flag.Bool("help", false, "ヘルプを表示")
		version  = flag.Bool("version", false, "バージョン情報を表示")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *version {
		fmt.Println("httplog version 1.0.0")
		return
	}

	// 環境変数からポート番号を取得（引数が優先）
	if *port == "8080" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		}
	}

	// ログディレクトリを設定
	logDir = *logPath

	// ログディレクトリの作成確認
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("ログディレクトリの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/", handler)
	
	fmt.Printf("HTTPログサーバーを開始します\n")
	fmt.Printf("ポート: %s\n", *port)
	fmt.Printf("ログ保存先: %s\n", logDir)
	fmt.Printf("停止するには Ctrl+C を押してください\n\n")
	
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Printf("サーバーの開始に失敗しました: %v\n", err)
		os.Exit(1)
	}
}