package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// v3--在前两个版本的基础上将域名、端口、超时、时间格式提取到命令行参数选项

func main() {
	// command-line flag parsing
	var d = flag.String("d", "www.baidu.com", "`domain`: 域名 IP")
	var p = flag.Int("p", 443, "`port` : 端口")
	var t = flag.Int("t", 5, "`timeout` : 连接超时")
	var s = flag.Bool("s", false, "seconds since 1970-01-01 00:00:00 UTC")
	flag.Parse()

	if *d == "" {
		// 如果没有给-d参数，也就是说域名为空，那么打印使用方法和错误信息
		usage := "Usage: " + os.Args[0] + " -d domain [-p port] [-t timeout] [-s]\n"
		fmt.Fprintf(os.Stderr, usage) //usage信息输出到标准错误输出
		os.Exit(2)
	}
	// define dialer with timeout
	var dialer net.Dialer
	dialer.Timeout = time.Duration(*t * 1000 * 1000 * 1000)

	// dial with dialer
	addr := *d + ":" + strconv.Itoa(*p) // convert int to string port number
	conn, err := tls.DialWithDialer(&dialer, "tcp", addr, &tls.Config{})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		state := conn.ConnectionState()
		certs := state.PeerCertificates
		cert := *certs[0]
		notAfter := cert.NotAfter.Local()
		if *s {
			fmt.Println(cert.Subject)
			fmt.Println(cert.Issuer)
			fmt.Println(notAfter)
		} else {
			fmt.Println(cert.Subject)
			fmt.Println(cert.Issuer)
			fmt.Println(notAfter)
		}
		conn.Close()
	}
}
