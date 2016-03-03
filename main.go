package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/aristanetworks/goeapi"
)

type Node struct {
	Hostname     string
	Username     string
	Password     string
	AuthValid    bool
	HttpEnabled  bool
	HttpsEnabled bool
	Version      string
	Model        string
}

func connect(host string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://bleaf1/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(body)

}

func stringToInterface(s []string) []interface{} {
	commands := make([]interface{}, len(s))
	for i, v := range s {
		commands[i] = v
	}
	return commands
}
func tryHttp(cmds []string, host string) (*goeapi.JSONRPCResponse, error) {
	e := goeapi.NewHTTPEapiConnection("http", host, "admin", "admin", 80)
	commands := stringToInterface(cmds)
	return e.Execute(commands, "json")
}

func tryHttps(cmds []string, host string) (*goeapi.JSONRPCResponse, error) {
	e := goeapi.NewHTTPSEapiConnection("https", host, "admin", "admin", 443)
	commands := stringToInterface(cmds)
	return e.Execute(commands, "json")
}

func tryHost(host string) Node {
	// First check if the host exists via tcp connection, avoid hard failure
	n := Node{Hostname: host}
	_, err := net.LookupHost(host)
	if err != nil {
		log.Printf("failed lookup for %s\n", host)
		return n
	}
	shver := []string{"show version"}
	h, err := tryHttp(shver, host)
	if err != nil {
		log.Println(err)
		n.HttpEnabled = false
	} else {
		n.HttpEnabled = true
		if h.Error != nil {
			log.Println(h.Error)
		}
	}
	hs, err := tryHttps(shver, host)
	if err != nil {
		log.Println(err)
		n.HttpsEnabled = false
	} else {
		n.HttpsEnabled = true
		if hs.Error != nil {
			log.Println(hs.Error)
			n.AuthValid = false
		} else {
			n.Version = hs.Result[0]["version"].(string)
			n.Model = hs.Result[0]["modelName"].(string)
			n.AuthValid = true
		}
	}
	fmt.Printf("%+v\n", n)
	return n
}
func main() {
	sws := []string{}
	for i := 1; i < 16; i++ {
		sws = append(sws, fmt.Sprintf("bleaf%d", i))
	}
	fmt.Println(sws)
	for _, h := range sws {
		go tryHost(h)
	}
	var input string
	fmt.Scanln(&input)
}
