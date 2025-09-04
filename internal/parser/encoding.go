package parser

import (
	"bytes"
	"io"
	"unicode/utf16"
	"unicode/utf8"
)

// DetectAndConvertEncoding 检测并转换文件编码
func DetectAndConvertEncoding(reader io.Reader) (io.Reader, error) {
	// 读取前几个字节来检测编码
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// 检测是否为UTF-16 LE (Little Endian)
	if n >= 2 && buf[0] == 0xFF && buf[1] == 0xFE {
		// UTF-16 LE BOM detected
		return convertUTF16LEToUTF8(reader, buf[2:n])
	}

	// 检测是否为UTF-16 BE (Big Endian)
	if n >= 2 && buf[0] == 0xFE && buf[1] == 0xFF {
		// UTF-16 BE BOM detected
		return convertUTF16BEToUTF8(reader, buf[2:n])
	}

	// 检测是否为UTF-16 LE (无BOM，通过模式检测)
	if isUTF16LE(buf[:n]) {
		return convertUTF16LEToUTF8(reader, buf[:n])
	}

	// 检测是否为UTF-16 BE (无BOM，通过模式检测)
	if isUTF16BE(buf[:n]) {
		return convertUTF16BEToUTF8(reader, buf[:n])
	}

	// 默认为UTF-8，创建一个新的reader包含已读取的数据
	return io.MultiReader(bytes.NewReader(buf[:n]), reader), nil
}

// isUTF16LE 检测是否为UTF-16 LE编码（无BOM）
func isUTF16LE(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// 检查是否有典型的UTF-16 LE模式：每隔一个字节为0
	nullCount := 0
	for i := 1; i < len(data) && i < 100; i += 2 {
		if data[i] == 0 {
			nullCount++
		}
	}

	// 如果超过50%的奇数位置是0，可能是UTF-16 LE
	return nullCount > (len(data)/4)
}

// isUTF16BE 检测是否为UTF-16 BE编码（无BOM）
func isUTF16BE(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// 检查是否有典型的UTF-16 BE模式：每隔一个字节为0
	nullCount := 0
	for i := 0; i < len(data) && i < 100; i += 2 {
		if data[i] == 0 {
			nullCount++
		}
	}

	// 如果超过50%的偶数位置是0，可能是UTF-16 BE
	return nullCount > (len(data)/4)
}

// convertUTF16LEToUTF8 将UTF-16 LE转换为UTF-8
func convertUTF16LEToUTF8(reader io.Reader, initialData []byte) (io.Reader, error) {
	// 读取所有数据
	allData := initialData
	restData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	allData = append(allData, restData...)

	// 确保数据长度为偶数
	if len(allData)%2 != 0 {
		allData = allData[:len(allData)-1]
	}

	// 转换为uint16切片
	uint16Data := make([]uint16, len(allData)/2)
	for i := 0; i < len(uint16Data); i++ {
		uint16Data[i] = uint16(allData[i*2]) | uint16(allData[i*2+1])<<8
	}

	// 转换为UTF-8
	utf8Data := utf16.Decode(uint16Data)
	utf8Bytes := make([]byte, 0, len(utf8Data)*4)
	for _, r := range utf8Data {
		if utf8.ValidRune(r) {
			buf := make([]byte, 4)
			n := utf8.EncodeRune(buf, r)
			utf8Bytes = append(utf8Bytes, buf[:n]...)
		}
	}

	return bytes.NewReader(utf8Bytes), nil
}

// convertUTF16BEToUTF8 将UTF-16 BE转换为UTF-8
func convertUTF16BEToUTF8(reader io.Reader, initialData []byte) (io.Reader, error) {
	// 读取所有数据
	allData := initialData
	restData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	allData = append(allData, restData...)

	// 确保数据长度为偶数
	if len(allData)%2 != 0 {
		allData = allData[:len(allData)-1]
	}

	// 转换为uint16切片（大端序）
	uint16Data := make([]uint16, len(allData)/2)
	for i := 0; i < len(uint16Data); i++ {
		uint16Data[i] = uint16(allData[i*2])<<8 | uint16(allData[i*2+1])
	}

	// 转换为UTF-8
	utf8Data := utf16.Decode(uint16Data)
	utf8Bytes := make([]byte, 0, len(utf8Data)*4)
	for _, r := range utf8Data {
		if utf8.ValidRune(r) {
			buf := make([]byte, 4)
			n := utf8.EncodeRune(buf, r)
			utf8Bytes = append(utf8Bytes, buf[:n]...)
		}
	}

	return bytes.NewReader(utf8Bytes), nil
}

// ParseTestLogWithEncoding 解析测试日志，自动处理编码
func ParseTestLogWithEncoding(reader io.Reader) (*TestResult, error) {
	// 检测并转换编码
	convertedReader, err := DetectAndConvertEncoding(reader)
	if err != nil {
		return nil, err
	}

	// 使用转换后的reader解析
	return ParseTestLog(convertedReader)
}

// ValidateTestLogWithEncoding 验证测试日志，自动处理编码
func ValidateTestLogWithEncoding(reader io.Reader) error {
	// 检测并转换编码
	convertedReader, err := DetectAndConvertEncoding(reader)
	if err != nil {
		return err
	}

	// 使用转换后的reader验证
	return ValidateTestLog(convertedReader)
}