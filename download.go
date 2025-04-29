package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func downloadImage(url, dir string) error {
	// 发送HTTP GET请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 从URL中提取文件名
	fileName := filepath.Base(url)
	// 如果URL中包含查询参数，去除它们
	if strings.Contains(fileName, "?") {
		fileName = strings.Split(fileName, "?")[0]
	}
	// 如果没有扩展名，添加默认的.jpg
	if !strings.Contains(fileName, ".") {
		fileName += ".jpg"
	}

	// 创建目标文件
	filePath := filepath.Join(dir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 将响应体复制到文件中
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("图片已保存到: %s\n", filePath)
	return nil
}
