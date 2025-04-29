package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

func main() {
	//如何跳过已经抓取过的页面
	//用于存储已发现的URL，避免重复处理
	visitedURLs := make(map[string]bool)
	visitedMutex := &sync.Mutex{}

	var (
		defaultDir = "download" // Directorio predeterminado para guardar los archivos descargados
		currentDir string
		outputDir  string
	)

	c := colly.NewCollector(
		colly.AllowedDomains("club.autohome.com.cn", "autohome.com.cn", "www.autohome.com.cn"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.95 Safari/537.36 QIHU 360SE"),
		colly.MaxDepth(2),  // Limita la profundidad de la navegación a 2 niveles
		colly.Async(false), // Habilita la ejecución asíncrona
	)

	//抓取列表页（分类页面）的帖子链接
	c.OnHTML("ul.content li a", func(e *colly.HTMLElement) {
		//ul class="content"
		link := e.Request.AbsoluteURL(e.Attr("href"))
		e.Request.Visit(link)
	})

	//抓取列表页的分页链接
	c.OnHTML("div.pages a", func(e *colly.HTMLElement) {
		// div class="pages" id="x-pages1"
		link := e.Request.AbsoluteURL(e.Attr("href"))
		
		//如何跳过已经抓取过的页面
		//检查是否访问过
		visitedMutex.Lock()
		_, visited := visitedURLs[link]
		visitedMutex.Unlock()
		if !visited {
			visitedURLs[link] = true
			e.Request.Visit(link)
		}
	})

	//抓取帖子的标题
	c.OnHTML("div.toolbar-left-title", func(e *colly.HTMLElement) {
		//decodeStr: 常用三个编码解码，解决乱码问题
		//sanitizeStr: 去除特殊字符
		currentDir = decodeStr(e.Attr("title"))
		outputDir = filepath.Join(defaultDir, sanitizeStr(currentDir))
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Printf("目录创建失败: %v\n", err)
			return
		}
	})

	/*
	现在把文本内容和图片放在一个OnHTML中
	视频中的思路是直接找到文本内容的div，这样无法避免重复生成txt文件
	现在的思路是找到更上级的div，文本内容先循环存在变量中，最后再写入文件
	由于现在用的是上级选择器，所以图片也可以放在同一个OnHTML中
	*/
	c.OnHTML("div.post-container", func(e *colly.HTMLElement) {
		// 抓取帖子的文本内容
		tmpText := ""
		//先循环找到文本
		e.ForEach("div.tz-paragraph, p.editor-paragraph", func(_ int, el *colly.HTMLElement) {
			tmpText += el.Text + "\n"
		})
		//再写入文件
		outputFile := filepath.Join(outputDir, "text.txt")
		err := os.WriteFile(outputFile, []byte(tmpText), 0644)
		if err != nil {
			fmt.Printf("文件写入失败: %v\n", err)
			return
		}

		//抓取帖子的图片
		e.ForEach("div.tz-picture img, div.editor-image img", func(_ int, el *colly.HTMLElement) {		
			imgSrc := el.Request.AbsoluteURL(el.Attr("data-src"))
			imgSrc = strings.ReplaceAll(imgSrc, "820_", "")
			err := downloadImage(imgSrc, outputDir)
			if err != nil {
				fmt.Printf("图片下载失败: %v\n", err)
				return
			}
		})
	})

	//抓取帖子的图片
	// c.OnHTML("div.tz-picture img", func(e *colly.HTMLElement) {
	// 	// imgSrc := e.Attr("data-src")
	// 	// imgSrc = "https:" + imgSrc
	// 	imgSrc := e.Request.AbsoluteURL(e.Attr("data-src"))
	// 	imgSrc = strings.ReplaceAll(imgSrc, "820_", "")
	// 	err := downloadImage(imgSrc, outputDir)
	// 	if err != nil {
	// 		fmt.Printf("图片下载失败: %v\n", err)
	// 		return
	// 	}
	// })

	c.OnRequest(func(r *colly.Request) {
		println("正在抓取", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		println("结束：", r.Request.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit("https://club.autohome.com.cn/bbs/thread/e3c463f8ebc587c7/111225288-1.html#pvareaid=102410")
}
