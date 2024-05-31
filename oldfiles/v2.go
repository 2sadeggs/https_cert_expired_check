package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"time"
)

// v2--改用HTTP替换tls建立连接获取证书
// 并增加flag命令行解析 用法示例
// go run .\v2.go -url https://www.baidu.com

func main() {
	trans := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // something like: curl -k
		},
	}
	client := &http.Client{
		Transport: trans,
		Timeout:   time.Second * 3, // 添加超时控制 防止域名访问不可达情况
	}
	var seedURL string
	flag.StringVar(&seedURL, "url", "https://www.baidu.com", "默认inSuite首页")
	flag.Parse()
	resp, err := client.Get(seedURL)
	if err != nil {
		fmt.Errorf(seedURL, "Client http get err")
		panic(err)
	}
	defer resp.Body.Close()
	certInfo := resp.TLS.PeerCertificates[0]
	fmt.Println("证书颁发机构信息：", certInfo.Issuer)
	fmt.Println("域名证书组织信息：", certInfo.Subject)
	fmt.Println("域名证书过期时间：", certInfo.NotAfter)
}
