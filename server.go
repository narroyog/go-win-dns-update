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
	Zone  string `json:"zone"`
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Ttl   int    `json:"ttl,omitempty"`
}

// func runPS(script string) ([]byte, error) {
// 	cmd := exec.Command("powershell", "-Command", "("+script+")")
// 	return cmd.Output()
// }

// func processReq(b []byte, m string) (int, string) {
// 	msg := ""
// 	msg1 := ""
// 	status := http.StatusOK
// 	var dnsrec DnsRecord
// 	if len(b) > 0 {
// 		err := json.Unmarshal(b, &dnsrec)
// 		if err != nil {
// 			return http.StatusBadRequest, err.Error()
// 		}
// 	}
// 	cmd := exec.Command("dnscmd", "/RecordDelete", dnsrec.Zone, dnsrec.Key, dnsrec.Type, "/f")
// 	out, err := cmd.Output()
// 	if err != nil {
// 		msg1 = "could not run RecordDelete: " + err.Error()
// 	} else {
// 		msg = string(out)
// 	}
// 	if m == "POST" {
// 		cmd = exec.Command("dnscmd", "/RecordAdd", dnsrec.Zone, dnsrec.Key, dnsrec.Type, dnsrec.Value)
// 		out, err := cmd.Output()
// 		if err != nil {
// 			if len(msg1) > 0 {
// 				msg1 = msg1 + "\n" + "could not run RecordAdd: " + err.Error()
// 			} else {
// 				msg1 = "could not run command: " + err.Error()
// 			}
// 		} else {
// 			msg1 = ""
// 			if len(msg) > 0 {
// 				msg = msg + "\n" + string(out)
// 			} else {
// 				msg = string(out)
// 			}
// 		}
// 	}
// 	if len(msg1) > 0 {
// 		msg = msg1
// 		status = http.StatusInternalServerError
// 	}
// 	return status, msg
// }

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
	cmd := "Get-DnsServerResourceRecord -ComputerName " + body.Key + " -ZoneName " + body.Zone
	ps := powershell.New()
	stdOut, stdErr, err := ps.execute(cmd)
	fmt.Printf("\nEnableHyperV:\nStdOut : '%s'\nStdErr: '%s'\nErr: %s", strings.TrimSpace(stdOut), stdErr, err)
	c.JSON(http.StatusAccepted, &body)
}

// func apiHandler(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case "POST", "DELETE":
// 		bodyBytes, err := io.ReadAll(r.Body)
// 		defer r.Body.Close()
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 		}
// 		status, msg := processReq(bodyBytes, r.Method)
// 		w.WriteHeader(status)
// 		w.Write([]byte(msg))
// 	default:
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		w.Write([]byte(r.Method + " not allowed"))
// 	}
// }

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
	// http.HandleFunc("/", apiHandler)
	// err := http.ListenAndServe(port, nil)
	// if err != nil {
	// 	panic(err)
	// }
}
