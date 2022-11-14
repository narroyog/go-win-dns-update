package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/narroyog/go-win-dns-update/powershell"
)

type DnsRecord struct {
	Zone    string `json:"zone"`
	Type    string `json:"type"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Ttl     string `json:"ttl,omitempty"`
	RevZone string `json:"revzone,omitempty"`
}

func reverse(str []string) (result string) {
	for _, v := range str {
		result = string(v) + "." + result
	}
	result = result[:len(result)-1]
	return
}

func getHandler(c *gin.Context) {
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", c.Request.Method, c.Request.URL, c.Request.Proto)
	request = append(request, url)                                     // Add the host
	request = append(request, fmt.Sprintf("Host: %v", c.Request.Host)) // Loop through headers
	for name, headers := range c.Request.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	c.String(http.StatusOK, strings.Join(request, "\n"))
}

func postHandler(c *gin.Context) {
	body := DnsRecord{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fmt.Println(body)
	dnsServer, ok := os.LookupEnv("DNS_SERVER")
	if !ok {
		dnsServer = "localhost"
	}
	ip := ""
	if len(body.RevZone) > 0 {
		octets := strings.Split(body.RevZone, ".")
		ipaddr := strings.Split(body.Value, ".")
		ip = reverse(ipaddr[len(octets):])
	}
	cmd := "$oldRec = Get-DnsServerResourceRecord -ComputerName " + dnsServer + " -ZoneName " + body.Zone + " -Name " + body.Key + " -ErrorAction SilentlyContinue \n"
	cmd = cmd + "if ($oldRec) {"
	cmd = cmd + "$newRec = Get-DnsServerResourceRecord -ComputerName " + dnsServer + " -ZoneName " + body.Zone + " -Name " + body.Key + "\n"
	switch body.Type {
	case "A":
		cmd = cmd + "$newRec.RecordData.IPv4Address = [System.Net.IPAddress]::parse('" + body.Value + "')\n"
	case "CNAME":
		cmd = cmd + "$newRec.RecordData.HostNameAlias = " + body.Value + "\n"
	default:
		panic("DNS record " + body.Type + " not implemented")
	}
	cmd = cmd + "Set-DnsServerResourceRecord -NewInputObject $newRec -OldInputObject $oldRec -ZoneName " + body.Zone + " -ComputerName " + dnsServer + "\n } else {\n"
	switch body.Type {
	case "CNAME":
		cmd = cmd + "Add-DnsServerResourceRecordCName -ZoneName " + body.Zone + " -Name " + body.Key + " -HostNameAlias " + body.Value + "\n}\n"
	case "A":
		cmd = cmd + "Add-DnsServerResourceRecordA -Name " + body.Key + " -IPv4Address " + body.Value + " -ZoneName " + body.Zone
		if len(body.Ttl) > 0 {
			cmd = cmd + " -TimeToLive " + body.Ttl + "\n}\n"
		}
		if len(ip) > 0 {
			cmd = cmd + "Add-DNSServerResourceRecordPTR -ZoneName " + body.RevZone + ".in-addr.arpa -Name " + ip + " -PTRDomainName " + body.Key + "." + body.Zone + ". -ErrorAction SilentlyContinue\n"
		}
	}
	ps := powershell.New()
	// fmt.Printf("%s", cmd)
	stdOut, stdErr, err := ps.Execute(cmd)
	if err != nil {
		fmt.Printf("\npostHandler:\nStdOut : '%s'\nStdErr: '%s'\nErr: %s\n", strings.TrimSpace(stdOut), stdErr, err)
	}
	c.JSON(http.StatusAccepted, &body)
}

func main() {
	port := ":8080"
	args := os.Args[1:]
	if len(args) > 0 {
		port = ":" + args[0]
	}
	router := gin.Default()
	router.GET("/", getHandler)
	router.POST("/", postHandler)
	router.Run(port)
	// r.RunTLS(port, "./server.pem", "./server.key")
}
