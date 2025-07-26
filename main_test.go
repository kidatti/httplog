package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateLogDir(t *testing.T) {
	// テスト用にカスタムログディレクトリを設定
	origLogDir := logDir
	logDir = "test_logs"
	defer func() {
		logDir = origLogDir
		cleanupLogs()
	}()
	
	dir, err := createLogDir()
	if err != nil {
		t.Fatalf("ディレクトリ作成エラー: %v", err)
	}
	
	if !strings.HasPrefix(dir, "test_logs/") {
		t.Errorf("期待されるディレクトリプレフィックスが異なります。取得: %s", dir)
	}
	
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("ディレクトリが作成されていません: %s", dir)
	}
}

func TestHandlerGETRequest(t *testing.T) {
	// テスト用にカスタムログディレクトリを設定
	origLogDir := logDir
	logDir = "test_logs"
	defer func() {
		logDir = origLogDir
		cleanupLogs()
	}()
	
	req, err := http.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ハンドラーが間違ったステータスコードを返しました: 取得 %v 期待 %v", status, http.StatusOK)
	}
	
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "success") || !strings.Contains(responseBody, "ログが保存されました") {
		t.Errorf("期待されるレスポンスボディが含まれていません: %s", responseBody)
	}
}

func TestHandlerPOSTRequest(t *testing.T) {
	// テスト用にカスタムログディレクトリを設定
	origLogDir := logDir
	logDir = "test_logs"
	defer func() {
		logDir = origLogDir
		cleanupLogs()
	}()
	
	postData := url.Values{}
	postData.Set("key1", "value1")
	postData.Set("key2", "value2")
	
	req, err := http.NewRequest("POST", "/test", strings.NewReader(postData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	handler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ハンドラーが間違ったステータスコードを返しました: 取得 %v 期待 %v", status, http.StatusOK)
	}
}

func TestHandlerJSONRequest(t *testing.T) {
	// テスト用にカスタムログディレクトリを設定
	origLogDir := logDir
	logDir = "test_logs"
	defer func() {
		logDir = origLogDir
		cleanupLogs()
	}()
	
	jsonData := map[string]interface{}{
		"test": "data",
		"number": 123,
	}
	jsonBytes, _ := json.Marshal(jsonData)
	
	req, err := http.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ハンドラーが間違ったステータスコードを返しました: 取得 %v 期待 %v", status, http.StatusOK)
	}
	
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "test_logs") {
		t.Errorf("レスポンスにログディレクトリパスが含まれていません: %s", responseBody)
	}
}

func TestLogFileCreation(t *testing.T) {
	// テスト用にカスタムログディレクトリを設定
	origLogDir := logDir
	logDir = "test_logs"
	defer func() {
		logDir = origLogDir
		cleanupLogs()
	}()
	
	req, err := http.NewRequest("GET", "/test?param=value", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "test")
	
	rr := httptest.NewRecorder()
	handler(rr, req)
	
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "test_logs") {
		t.Fatal("ログディレクトリパスがレスポンスに含まれていません")
	}

	// JSONレスポンスからlogDirを取得
	var response map[string]string
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		t.Fatalf("JSONレスポンスのパースエラー: %v", err)
	}

	logDirPath, ok := response["logDir"]
	if !ok {
		t.Fatal("logDirがレスポンスに含まれていません")
	}
	
	// logDirPathは上で取得済み
	logFilePath := filepath.Join(logDirPath, "log.txt")
	
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("ログファイルが作成されていません: %s", logFilePath)
	}
	
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("ログファイルの読み取りエラー: %v", err)
	}
	
	var logData RequestLog
	if err := json.Unmarshal(content, &logData); err != nil {
		t.Fatalf("ログデータのパースエラー: %v", err)
	}
	
	if logData.Method != "GET" {
		t.Errorf("期待されるメソッド: GET, 取得: %s", logData.Method)
	}
	
	if logData.URL != "/test?param=value" {
		t.Errorf("期待されるURL: /test?param=value, 取得: %s", logData.URL)
	}
}

func cleanupLogs() {
	os.RemoveAll("logs")
	os.RemoveAll("test_logs")
}

// カスタムログディレクトリのテスト
func TestCustomLogDirectory(t *testing.T) {
	customDir := "custom_test_logs"
	origLogDir := logDir
	logDir = customDir
	defer func() {
		logDir = origLogDir
		os.RemoveAll(customDir)
	}()

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ハンドラーが間違ったステータスコードを返しました: 取得 %v 期待 %v", status, http.StatusOK)
	}

	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, customDir) {
		t.Errorf("カスタムログディレクトリが使用されていません: %s", responseBody)
	}

	// JSONレスポンスの検証
	var response map[string]string
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		t.Errorf("JSONレスポンスのパースエラー: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("期待されるステータス: success, 取得: %s", response["status"])
	}
}

// OPTIONSリクエストのテスト
func TestHandlerOPTIONSRequest(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ハンドラーが間違ったステータスコードを返しました: 取得 %v 期待 %v", status, http.StatusOK)
	}

	// CORSヘッダーの検証
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("CORS Originヘッダーが正しく設定されていません: %s", origin)
	}

	if methods := rr.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "POST") {
		t.Errorf("CORS Methodsヘッダーが正しく設定されていません: %s", methods)
	}
}