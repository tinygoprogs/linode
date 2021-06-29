package main

import (
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var urls = map[string]string{
	"instances": "https://api.linode.com/v4/linode/instances",
	"transfer":  "https://api.linode.com/v4/linode/instances/%d/transfer",
}

type LinodeList struct {
	Data []struct {
		Id    int
		Label string
		Ipv4  []string
		Specs struct {
			Disk   uint64
			Memory uint64
			Vcpus  uint64
			//Gpus     uint64
			Transfer uint64
		}
	}
}

type LinodeNetworkTransfer struct {
	Used  uint64
	Quota uint64
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run linodeinfo.go <path-to-token>")
	}
	tokenFile := os.Args[1]
	c := http.Client{}
	tokenB, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Fatal(err)
	}
	token := strings.Trim(string(tokenB), " \n")
	authHeader := map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	url, err := url.Parse(urls["instances"])
	if err != nil {
		log.Fatal(err)
	}
	instances, err := c.Do(&http.Request{
		Method: "GET",
		URL:    url,
		Header: authHeader,
	})
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(instances.Body)
	if err != nil {
		log.Fatal(err)
	}
	linodes := LinodeList{}
	err = json.Unmarshal(data, &linodes)
	if err != nil {
		log.Fatal(err)
	}
	//log.Printf("linodes: %v", linodes)
	for _, linode := range linodes.Data {
		id := linode.Id
		url, err = url.Parse(fmt.Sprintf(urls["transfer"], id))
		if err != nil {
			log.Fatal(err)
		}
		tranfer, err := c.Do(&http.Request{
			Method: "GET",
			URL:    url,
			Header: authHeader,
		})
		if err != nil {
			log.Fatal(err)
		}
		data, err := ioutil.ReadAll(tranfer.Body)
		if err != nil {
			log.Fatal(err)
		}
		transferinfo := LinodeNetworkTransfer{}
		err = json.Unmarshal(data, &transferinfo)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v: %s @ %s used %s out of %d GB (~ %.2f %%)\n",
			time.Now().Format("2006-01-02T15:04:05-0700"),
			linode.Label, linode.Ipv4[0], humanize.Bytes(transferinfo.Used), linode.Specs.Transfer,
			float64(transferinfo.Used)/float64(linode.Specs.Transfer*1024*1024*1024)*100)
	}
}
