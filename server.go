package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

type DnsRecord struct {
	Zone  string `json:"zone"`
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Ttl   int    `json:"ttl,omitempty"`
}

func processReq(b []byte, m string) (int, string) {
	msg := ""
	msg1 := ""
	status := http.StatusOK
	var dnsrec DnsRecord
	if len(b) > 0 {
		err := json.Unmarshal(b, &dnsrec)
		if err != nil {
			return http.StatusBadRequest, err.Error()
		}
	}
	cmd := exec.Command("dnscmd", "/recorddelete", dnsrec.Zone, dnsrec.Key, dnsrec.Type, "/f")
	out, err := cmd.Output()
	if err != nil {
		msg1 = "could not run recorddelete: " + err.Error()
	} else {
		msg = string(out)
	}
	if m == "POST" {
		cmd = exec.Command("dnscmd", "/recordadd", dnsrec.Zone, dnsrec.Key, dnsrec.Type, dnsrec.Value)
		out, err := cmd.Output()
		if err != nil {
			if len(msg1) > 0 {
				msg1 = msg1 + "\n" + "could not run recordadd: " + err.Error()
			} else {
				msg1 = "could not run command: " + err.Error()
			}
		} else {
			msg1 = ""
			if len(msg) > 0 {
				msg = msg + "\n" + string(out)
			} else {
				msg = string(out)
			}
		}
	}
	if len(msg1) > 0 {
		msg = msg1
		status = http.StatusInternalServerError
	}
	return status, msg
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var request []string
		// Add the request string
		url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
		request = append(request, url)                             // Add the host
		request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
		for name, headers := range r.Header {
			name = strings.ToLower(name)
			for _, h := range headers {
				request = append(request, fmt.Sprintf("%v: %v", name, h))
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Join(request, "\n")))
	case "POST", "DELETE":
		bodyBytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		status, msg := processReq(bodyBytes, r.Method)
		w.WriteHeader(status)
		w.Write([]byte(msg))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(r.Method + " not allowed"))
	}
}

func main() {
	http.HandleFunc("/", apiHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
