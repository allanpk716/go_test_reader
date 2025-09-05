package parser

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"unicode/utf16"
)

// TestDetectAndConvertEncoding_UTF8 测试UTF-8编码检测
func TestDetectAndConvertEncoding_UTF8(t *testing.T) {
	// Arrange
	utf8Data := "Hello, 世界!"
	reader := strings.NewReader(utf8Data)

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if string(output) != utf8Data {
		t.Errorf("Expected %q, got %q", utf8Data, string(output))
	}
}

// TestDetectAndConvertEncoding_UTF16LE_WithBOM 测试UTF-16 LE带BOM的检测和转换
func TestDetectAndConvertEncoding_UTF16LE_WithBOM(t *testing.T) {
	// Arrange
	text := "Hello"
	utf16Data := utf16.Encode([]rune(text))
	buf := bytes.NewBuffer([]byte{0xFF, 0xFE}) // UTF-16 LE BOM
	for _, r := range utf16Data {
		buf.WriteByte(byte(r))
		buf.WriteByte(byte(r >> 8))
	}
	reader := bytes.NewReader(buf.Bytes())

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if string(output) != text {
		t.Errorf("Expected %q, got %q", text, string(output))
	}
}

// TestDetectAndConvertEncoding_UTF16BE_WithBOM 测试UTF-16 BE带BOM的检测和转换
func TestDetectAndConvertEncoding_UTF16BE_WithBOM(t *testing.T) {
	// Arrange
	text := "Hello"
	utf16Data := utf16.Encode([]rune(text))
	buf := bytes.NewBuffer([]byte{0xFE, 0xFF}) // UTF-16 BE BOM
	for _, r := range utf16Data {
		buf.WriteByte(byte(r >> 8))
		buf.WriteByte(byte(r))
	}
	reader := bytes.NewReader(buf.Bytes())

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if string(output) != text {
		t.Errorf("Expected %q, got %q", text, string(output))
	}
}

// TestDetectAndConvertEncoding_EmptyInput 测试空输入
func TestDetectAndConvertEncoding_EmptyInput(t *testing.T) {
	// Arrange
	reader := strings.NewReader("")

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if len(output) != 0 {
		t.Errorf("Expected empty output, got %q", string(output))
	}
}

// TestIsUTF16LE_ValidPattern 测试UTF-16 LE模式检测
func TestIsUTF16LE_ValidPattern(t *testing.T) {
	// Arrange - 创建典型的UTF-16 LE模式数据
	data := []byte{'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0}

	// Act
	result := isUTF16LE(data)

	// Assert
	if !result {
		t.Error("Expected true for UTF-16 LE pattern")
	}
}

// TestIsUTF16LE_InvalidPattern 测试非UTF-16 LE模式
func TestIsUTF16LE_InvalidPattern(t *testing.T) {
	// Arrange - 普通UTF-8数据
	data := []byte("Hello World")

	// Act
	result := isUTF16LE(data)

	// Assert
	if result {
		t.Error("Expected false for non-UTF-16 LE pattern")
	}
}

// TestIsUTF16LE_TooShort 测试数据太短的情况
func TestIsUTF16LE_TooShort(t *testing.T) {
	// Arrange
	data := []byte{0x48}

	// Act
	result := isUTF16LE(data)

	// Assert
	if result {
		t.Error("Expected false for too short data")
	}
}

// TestIsUTF16BE_ValidPattern 测试UTF-16 BE模式检测
func TestIsUTF16BE_ValidPattern(t *testing.T) {
	// Arrange - 创建典型的UTF-16 BE模式数据
	data := []byte{0, 'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o'}

	// Act
	result := isUTF16BE(data)

	// Assert
	if !result {
		t.Error("Expected true for UTF-16 BE pattern")
	}
}

// TestIsUTF16BE_InvalidPattern 测试非UTF-16 BE模式
func TestIsUTF16BE_InvalidPattern(t *testing.T) {
	// Arrange - 普通UTF-8数据
	data := []byte("Hello World")

	// Act
	result := isUTF16BE(data)

	// Assert
	if result {
		t.Error("Expected false for non-UTF-16 BE pattern")
	}
}

// TestIsUTF16BE_TooShort 测试数据太短的情况
func TestIsUTF16BE_TooShort(t *testing.T) {
	// Arrange
	data := []byte{0x00, 0x48}

	// Act
	result := isUTF16BE(data)

	// Assert
	if result {
		t.Error("Expected false for too short data")
	}
}

// TestConvertUTF16LEToUTF8_ValidData 测试UTF-16 LE到UTF-8的转换
func TestConvertUTF16LEToUTF8_ValidData(t *testing.T) {
	// Arrange
	text := "Hello"
	utf16Data := utf16.Encode([]rune(text))
	buf := bytes.NewBuffer(nil)
	for _, r := range utf16Data {
		buf.WriteByte(byte(r))
		buf.WriteByte(byte(r >> 8))
	}
	reader := strings.NewReader("")
	initialData := buf.Bytes()

	// Act
	result, err := convertUTF16LEToUTF8(reader, initialData)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if string(output) != text {
		t.Errorf("Expected %q, got %q", text, string(output))
	}
}

// TestConvertUTF16LEToUTF8_OddLength 测试奇数长度数据的处理
func TestConvertUTF16LEToUTF8_OddLength(t *testing.T) {
	// Arrange - 奇数长度的数据
	initialData := []byte{'H', 0, 'i', 0, 'X'} // 最后一个字节会被丢弃
	reader := strings.NewReader("")

	// Act
	result, err := convertUTF16LEToUTF8(reader, initialData)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	expected := "Hi"
	if string(output) != expected {
		t.Errorf("Expected %q, got %q", expected, string(output))
	}
}

// TestConvertUTF16BEToUTF8_ValidData 测试UTF-16 BE到UTF-8的转换
func TestConvertUTF16BEToUTF8_ValidData(t *testing.T) {
	// Arrange
	text := "Hello"
	utf16Data := utf16.Encode([]rune(text))
	buf := bytes.NewBuffer(nil)
	for _, r := range utf16Data {
		buf.WriteByte(byte(r >> 8))
		buf.WriteByte(byte(r))
	}
	reader := strings.NewReader("")
	initialData := buf.Bytes()

	// Act
	result, err := convertUTF16BEToUTF8(reader, initialData)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if string(output) != text {
		t.Errorf("Expected %q, got %q", text, string(output))
	}
}

// TestConvertUTF16BEToUTF8_OddLength 测试奇数长度数据的处理
func TestConvertUTF16BEToUTF8_OddLength(t *testing.T) {
	// Arrange - 奇数长度的数据
	initialData := []byte{0, 'H', 0, 'i', 'X'} // 最后一个字节会被丢弃
	reader := strings.NewReader("")

	// Act
	result, err := convertUTF16BEToUTF8(reader, initialData)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	expected := "Hi"
	if string(output) != expected {
		t.Errorf("Expected %q, got %q", expected, string(output))
	}
}

// TestParseTestLogWithEncoding_UTF8 测试带编码的解析功能（UTF-8）
func TestParseTestLogWithEncoding_UTF8(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`
	reader := strings.NewReader(testInput)

	// Act
	result, err := ParseTestLogWithEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalTests != 1 {
		t.Errorf("Expected TotalTests=1, got %d", result.TotalTests)
	}
	if result.PassedTests != 1 {
		t.Errorf("Expected PassedTests=1, got %d", result.PassedTests)
	}
}

// TestValidateTestLogWithEncoding_UTF8 测试带编码的验证功能（UTF-8）
func TestValidateTestLogWithEncoding_UTF8(t *testing.T) {
	// Arrange
	testInput := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}
{"Time":"2024-01-15T10:30:02Z","Action":"pass","Package":"example/pkg","Test":"TestExample1","Elapsed":0.001}`
	reader := strings.NewReader(testInput)

	// Act
	err := ValidateTestLogWithEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestValidateTestLogWithEncoding_InvalidJSON 测试带编码的验证功能（无效JSON）
func TestValidateTestLogWithEncoding_InvalidJSON(t *testing.T) {
	// Arrange - 提供多行数据，其中大部分是无效JSON（超过50%）
	testInput := `invalid json line 1
invalid json line 2
invalid json line 3
{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"pkg","Test":"test"}`
	reader := strings.NewReader(testInput)

	// Act
	err := ValidateTestLogWithEncoding(reader)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

// TestDetectAndConvertEncoding_ReadError 测试读取错误的处理
func TestDetectAndConvertEncoding_ReadError(t *testing.T) {
	// Arrange - 创建一个会产生读取错误的reader
	reader := &errorReader{}

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

// TestConvertUTF16LEToUTF8_ReadError 测试UTF-16 LE转换时的读取错误
func TestConvertUTF16LEToUTF8_ReadError(t *testing.T) {
	// Arrange
	reader := &errorReader{}
	initialData := []byte{}

	// Act
	result, err := convertUTF16LEToUTF8(reader, initialData)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

// TestConvertUTF16BEToUTF8_ReadError 测试UTF-16 BE转换时的读取错误
func TestConvertUTF16BEToUTF8_ReadError(t *testing.T) {
	// Arrange
	reader := &errorReader{}
	initialData := []byte{}

	// Act
	result, err := convertUTF16BEToUTF8(reader, initialData)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

// TestDetectAndConvertEncoding_UTF16LE_NoBOM 测试无BOM的UTF-16 LE检测
func TestDetectAndConvertEncoding_UTF16LE_NoBOM(t *testing.T) {
	// Arrange - 创建足够长的UTF-16 LE数据以触发模式检测
	data := make([]byte, 200)
	for i := 0; i < 100; i += 2 {
		data[i] = byte('A' + (i/2)%26)  // ASCII字符
		data[i+1] = 0                   // 高字节为0（LE模式）
	}
	reader := bytes.NewReader(data)

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证包含预期字符
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

// TestDetectAndConvertEncoding_UTF16BE_NoBOM 测试无BOM的UTF-16 BE检测
func TestDetectAndConvertEncoding_UTF16BE_NoBOM(t *testing.T) {
	// Arrange - 创建足够长的UTF-16 BE数据以触发模式检测
	data := make([]byte, 200)
	for i := 0; i < 100; i += 2 {
		data[i] = 0                     // 高字节为0（BE模式）
		data[i+1] = byte('A' + (i/2)%26) // ASCII字符
	}
	reader := bytes.NewReader(data)

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证包含预期字符
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

// TestParseTestLogWithEncoding_UTF16LE 测试UTF-16 LE编码的解析
func TestParseTestLogWithEncoding_UTF16LE(t *testing.T) {
	// Arrange - 创建UTF-16 LE编码的测试日志
	testJSON := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}`
	utf16Data := utf16.Encode([]rune(testJSON))
	buf := bytes.NewBuffer([]byte{0xFF, 0xFE}) // UTF-16 LE BOM
	for _, r := range utf16Data {
		buf.WriteByte(byte(r))
		buf.WriteByte(byte(r >> 8))
	}
	reader := bytes.NewReader(buf.Bytes())

	// Act
	result, err := ParseTestLogWithEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if len(result.Packages) == 0 {
		t.Error("Expected at least one package")
	}
}

// TestValidateTestLogWithEncoding_UTF16BE 测试UTF-16 BE编码的验证
func TestValidateTestLogWithEncoding_UTF16BE(t *testing.T) {
	// Arrange - 创建UTF-16 BE编码的测试日志
	testJSON := `{"Time":"2024-01-15T10:30:00Z","Action":"run","Package":"example/pkg","Test":"TestExample1"}`
	utf16Data := utf16.Encode([]rune(testJSON))
	buf := bytes.NewBuffer([]byte{0xFE, 0xFF}) // UTF-16 BE BOM
	for _, r := range utf16Data {
		buf.WriteByte(byte(r >> 8))
		buf.WriteByte(byte(r))
	}
	reader := bytes.NewReader(buf.Bytes())

	// Act
	err := ValidateTestLogWithEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestParseTestLogWithEncoding_EncodingError 测试编码检测错误
func TestParseTestLogWithEncoding_EncodingError(t *testing.T) {
	// Arrange
	reader := &errorReader{}

	// Act
	result, err := ParseTestLogWithEncoding(reader)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

// TestValidateTestLogWithEncoding_EncodingError 测试编码检测错误
func TestValidateTestLogWithEncoding_EncodingError(t *testing.T) {
	// Arrange
	reader := &errorReader{}

	// Act
	err := ValidateTestLogWithEncoding(reader)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestDetectAndConvertEncoding_ShortData 测试短数据的处理
func TestDetectAndConvertEncoding_ShortData(t *testing.T) {
	// Arrange - 只有一个字节的数据
	reader := strings.NewReader("A")

	// Act
	result, err := DetectAndConvertEncoding(reader)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// 读取结果并验证
	output, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	if string(output) != "A" {
		t.Errorf("Expected 'A', got %q", string(output))
	}
}

// errorReader 是一个总是返回错误的Reader，用于测试错误处理
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}