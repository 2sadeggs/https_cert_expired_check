# https_cert_expired_check

**域名证书有效期检查shell脚本和命令行工具**

**Shell版**

* https_cert_check_v2_openssl.sh and https_cert_check_v2_curl.sh
  ```bash
  # v2版可以批量处理多个域名 记录日志 并记录检查结果到csv文件
  # 用法如下 执行完会输出带时间戳的日志文件与检查结果csv文件
  bash https_cert_check_v2_openssl.sh
  bash https_cert_check_v2_curl.sh
  # curl版超时控制友好点 但是依赖curl命令；
  # openssl版一般Linux和Windows甚至Mac系统都有集成 兼容性更好
  ```

**golang版**

* 用法说明
  ```
  Usage of \main.exe:           
    -c int           # 并发数 默认42                                                                           
          Maximum number of hosts to check at once. (default 42)                                
    -d int           # 查询多少天内要过期的证书 默认30                                                                           
          Warn if the certificate will expire within this many days. (default 30)               
    -h string        # 域名列表文件 用于批量查询 *必选项*                                                                           
          The path to the file containing a list of hosts to check.                             
    -m int           # 查询几个月内要过期的证书                                                                           
          Warn if the certificate will expire within this many months.                          
    -s    Verify that non-root certificates are using a good signature algorithm. (default true)     # 检查非根证书算法是否在支持列表
    -t int           # 连接超时控制 主要用于处理域名不可达情况                                                                           
          Dial connect timeout seconds (default 3)                                             
    -y int           # 查询几年内要过去的证书                                                                           
          Warn if the certificate will expire within this many years.
  ```

* 示例
  ```
  # 查询当前目录下hosts.txt文件里的域名2年内到期的所有证书列表
  go run .\main.go -h .\hosts.txt -y 2
  ```

* 跨平台编译
  ```
  Windows 下编译 Mac 和 Linux 64位可执行程序
  SET CGO_ENABLED=0
  SET GOOS=darwin
  SET GOARCH=amd64
  go build main.go
  SET CGO_ENABLED=0
  SET GOOS=linux
  SET GOARCH=amd64
  go build main.go
  
  Mac 下编译 Linux 和 Windows 64位可执行程序
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
  
  Linux 下编译 Mac 和 Windows 64位可执行程序
  CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
  
  ```

  # https_cert_expired_check
