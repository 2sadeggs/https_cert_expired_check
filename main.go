package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// 默认并发数量
const defaultConcurrency = 42

// 格式化输出过期时间 计算单位小时shortly 天soon
const (
	errExpiringShortly = "%s: ** '%s' (S/N %X) expires in %d hours! **"
	errExpiringSoon    = "%s: '%s' (S/N %X) expires in roughly %d days."
	errSunsetAlg       = "%s: '%s' (S/N %X) expires after the sunset date for its signature algorithm '%s'."
)

// 定义一个结构体 有name和过期时间两个字段
type sigAlgSunset struct {
	name      string
	sunsetsAt time.Time
}

// 初始化一些算法map key为算法名 value为上述结构体
var sunsetSigAlgs = map[x509.SignatureAlgorithm]sigAlgSunset{
	x509.MD2WithRSA: sigAlgSunset{
		name:      "MD2 with RSA",
		sunsetsAt: time.Now(),
	},
	x509.MD5WithRSA: sigAlgSunset{
		name:      "MD5 with RSA",
		sunsetsAt: time.Now(),
	},
	x509.SHA1WithRSA: sigAlgSunset{
		name:      "SHA1 with RSA",
		sunsetsAt: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	x509.DSAWithSHA1: sigAlgSunset{
		name:      "DSA with SHA1",
		sunsetsAt: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	x509.ECDSAWithSHA1: sigAlgSunset{
		name:      "ECDSA with SHA1",
		sunsetsAt: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// 命令行参数
var (
	//hostsFile   = flag.String("h", "./hosts.txt", "The path to the file containing a list of hosts to check.")
	hostsFile   = flag.String("h", "", "The path to the file containing a list of hosts to check.")
	timeOut     = flag.Int("t", 3, "Dial connect timeout seconds")
	warnYears   = flag.Int("y", 0, "Warn if the certificate will expire within this many years.")
	warnMonths  = flag.Int("m", 0, "Warn if the certificate will expire within this many months.")
	warnDays    = flag.Int("d", 30, "Warn if the certificate will expire within this many days.")
	checkSigAlg = flag.Bool("s", true, "Verify that non-root certificates are using a good signature algorithm.")
	concurrency = flag.Int("c", defaultConcurrency, "Maximum number of hosts to check at once.")
)

// 定义一个证书错误结构体 返回证书的name和error
type certErrors struct {
	commonName string
	errs       []error
}

// 定义一个域名检测结果结构体 包含域名 错误 证书错误相关信息
type hostResult struct {
	host  string
	err   error
	certs []certErrors
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 判断hostFile值的长度 如果为0 表示没有给hostFile 但是命令行给了默认值
	// 打印命令行使用格式
	if len(*hostsFile) == 0 {
		flag.Usage()
		return
	}
	// 如果给定的期限为负数 那么将负数改为0
	if *warnYears < 0 {
		*warnYears = 0
	}
	if *warnMonths < 0 {
		*warnMonths = 0
	}
	if *warnDays < 0 {
		*warnDays = 0
	}
	// 如果三个给定的期限都为0 那么默认使用“天”的默认期限
	if *warnYears == 0 && *warnMonths == 0 && *warnDays == 0 {
		*warnDays = 30
	}
	// 如果给定的并发数小于0 那么将使用并发的默认值
	if *concurrency < 0 {
		*concurrency = defaultConcurrency
	}

	// 开始进行域名证书有效期查询 封装成一个函数
	processHosts()
}
func processHosts() {
	// define a null channel without cache for end of task
	done := make(chan struct{})
	defer close(done)

	hosts := queueHosts(done)
	results := make(chan hostResult)

	var wg sync.WaitGroup
	wg.Add(*concurrency)
	for i := 0; i < *concurrency; i++ {
		go func() {
			processQueue(done, hosts, results)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			log.Printf("%s: %#v\n", r.host, r.err)
			continue
		}
		for _, cert := range r.certs {
			for _, err := range cert.errs {
				log.Println(err)
			}
		}
	}
}

// 函数queueHosts带一个done的参数 返回整个域名文件的域名列表到一个管道
func queueHosts(done <-chan struct{}) <-chan string {
	hosts := make(chan string)
	go func() {
		defer close(hosts)
		fileContents, err := os.ReadFile(*hostsFile)
		if err != nil {
			return
		}
		lines := strings.Split(string(fileContents), "\n")
		for _, line := range lines {
			host := strings.TrimSpace(line)
			// 判断空行或注释行 跳过
			if len(host) == 0 || host[0] == '#' {
				continue
			}
			select {
			case hosts <- host:
			case <-done:
				// 此处利用的是关闭channel的方式退出goroutine
				// 因为done本身是个无缓冲管道 除非在别的协程里为done发送一个值 否则此case一直堵塞
				// 但是有一种情况例外 就是管道channel被关闭了
				// 那么在哪里done被关闭了呢 在刚定义完管道后的defer clone
				return
			}
		}
	}()
	return hosts
}

// 函数processQueue用checkHost遍历所有的域名列表 将结果返回到results管道 同样采用关闭管道done的方式结束select
func processQueue(done <-chan struct{}, hosts <-chan string, results chan<- hostResult) {
	for host := range hosts {
		select {
		case results <- checkHost(host):
		case <-done:
			return
		}
	}
}

// 函数checkHost 遍历每一个域名 将结果返回到result hostResult是个结构体
func checkHost(host string) (result hostResult) {
	result = hostResult{
		host:  host,
		certs: []certErrors{},
	}
	//conn, err := tls.Dial("tcp", host, nil)
	var dialer net.Dialer
	dialer.Timeout = time.Duration(*timeOut * 1000 * 1000 * 1000)
	conn, err := tls.DialWithDialer(&dialer, "tcp", host, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		result.err = err
		return
	}
	defer conn.Close()

	timeNow := time.Now()
	checkedCerts := make(map[string]struct{})
	//log.Printf("VerifiedChains: %#v", conn.ConnectionState().VerifiedChains)
	for _, chain := range conn.ConnectionState().VerifiedChains {
		//log.Printf("current chain: %#v", chain)
		for certNum, cert := range chain {
			//log.Printf("current certNum: %#v", certNum)
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				//log.Printf("current cert checked: %#v", checked)
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}
			cErrs := []error{}

			// Check the expiration.
			if timeNow.AddDate(*warnYears, *warnMonths, *warnDays).After(cert.NotAfter) {
				expiresIn := int64(cert.NotAfter.Sub(timeNow).Hours())
				if expiresIn <= 48 {
					cErrs = append(cErrs, fmt.Errorf(errExpiringShortly, host, cert.Subject.CommonName, cert.SerialNumber, expiresIn))
				} else {
					cErrs = append(cErrs, fmt.Errorf(errExpiringSoon, host, cert.Subject.CommonName, cert.SerialNumber, expiresIn/24))
				}
			}

			// Check the signature algorithm, ignoring the root certificate.
			if alg, exists := sunsetSigAlgs[cert.SignatureAlgorithm]; *checkSigAlg && exists && certNum != len(chain)-1 {
				if cert.NotAfter.Equal(alg.sunsetsAt) || cert.NotAfter.After(alg.sunsetsAt) {
					cErrs = append(cErrs, fmt.Errorf(errSunsetAlg, host, cert.Subject.CommonName, cert.SerialNumber, alg.name))
				}
			}

			result.certs = append(result.certs, certErrors{
				commonName: cert.Subject.CommonName,
				errs:       cErrs,
			})
		}
	}
	return
}
