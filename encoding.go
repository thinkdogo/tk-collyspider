package main

import (
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// decodeStr 解码标题
func decodeStr(text string) string {
	// 尝试UTF-8
	if utf8.ValidString(text) {
		return text
	}

	// 尝试GBK
	if gbk, err := decodeEncoding(text, simplifiedchinese.GBK); err == nil {
		return gbk
	}

	// 尝试GB2312
	if gb2312, err := decodeEncoding(text, simplifiedchinese.HZGB2312); err == nil {
		return gb2312
	}

	return text
}

// decodeEncoding 解码特定编码
func decodeEncoding(text string, enc encoding.Encoding) (string, error) {
	reader := transform.NewReader(strings.NewReader(text), enc.NewDecoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(b), nil
}


// sanitizeFilename 去特殊字符
func sanitizeStr(name string) string {
	// 替换非法字符
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	// 去除控制字符
	name = strings.Map(func(r rune) rune {
		if r <= 31 || r == 127 {
			return -1
		}
		return r
	}, name)

	// 限制长度
	name = strings.TrimSpace(name)
	if len(name) > 100 {
		name = name[:100]
	}
	if name == "" {
		name = "untitled_" + time.Now().Format("20060102")
	}

	return name
}