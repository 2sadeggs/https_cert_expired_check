package main

import (
	"crypto/tls"
	"fmt"
	"time"
)

//https://www.freecodecamp.org/news/how-to-validate-ssl-certificates-in-go/

func main() {
	//Step 1: Check if your website has an SSL certificate
	conn, err := tls.Dial("tcp", "www.baidu.com:443", nil)
	if err != nil {
		panic("Server doesn't support SSL certificate err: " + err.Error())
	}
	defer conn.Close()
	//Step 2: Check whether the SSL certificate and website hostname match
	err = conn.VerifyHostname("www.baidu.com")
	if err != nil {
		panic("Hostname doesn't match with certificate: " + err.Error())
	}
	//Step 3: Verify the expiration date of the server's SSL certificate
	issuer := conn.ConnectionState().PeerCertificates[0].Issuer
	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	//fmt.Printf("Issuer: %s\nExpiry: %v\n", issuer, expiry)
	fmt.Printf("Issuer: %s\nExpiry: %v\n", issuer, expiry.Format(time.RFC3339))
}
