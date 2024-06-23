package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloud-org/msgpush"
	"github.com/imroc/req"
)

var (
	//go:embed token
	tokenList string
)

func main() {
	dingToken := os.Getenv("DINGTALK_TOKEN")
	if len(dingToken) == 0 {
		panic("no ding talk token")
	}
	d := msgpush.NewDingTalk(dingToken)

	tokens := strings.Split(strings.Trim(tokenList, "\n"), "\n")
	networkAddrs := make(map[string]string)
	tokenName := make(map[string]string)
	for i := 0; i < len(tokens); i++ {
		token := strings.Split(tokens[i], ",")
		_, ok := networkAddrs[token[0]]
		if !ok {
			networkAddrs[token[0]] = token[1]
		} else {
			networkAddrs[token[0]] += "," + token[1]
		}
		tokenName[token[1]] = token[2]
	}

	//curl --location --request GET 'https://api.geckoterminal.com/api/v2/simple/networks/solana/token_price/F6fw97fXctQkkZDzmXrdsqm2Um2vtGnkdSnUQ6V2g9Q2' \
	//--header 'Accept: application/json;version=20230302'
	//
	//{"data":{"id":"9ac7da31-ed03-4c48-8504-747d949ffdfe","type":"simple_token_price","attributes":{"token_prices":{"F6fw97fXctQkkZDzmXrdsqm2Um2vtGnkdSnUQ6V2g9Q2":"0.0000916249278837784"}}}

	addrPrice := make(map[string]string)

	for network, addrs := range networkAddrs {
		resp, err := req.Get(fmt.Sprintf("https://api.geckoterminal.com/api/v2/simple/networks/%s/token_price/%s", network, addrs))
		if err != nil {
			panic(err)
		}
		var r GetTokenPriceResp
		if err = resp.ToJSON(&r); err != nil {
			panic(err)
		}
		for addr, price := range r.Data.Attributes.TokenPrices {
			addrPrice[addr] = price
		}
	}

	var sendText bytes.Buffer
	sendText.WriteString(fmt.Sprintf("token price, time: %s\n", time.Now().Format(time.RFC3339)))
	for addr, price := range addrPrice {
		sendText.WriteString(fmt.Sprintf("name: %s, addr: %s, price: %s\n", tokenName[addr], addr, price))
	}

	_ = d.SendText(sendText.String())
}

type GetTokenPriceResp struct {
	Data struct {
		Id         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			TokenPrices map[string]string `json:"token_prices"`
		} `json:"attributes"`
	} `json:"data"`
}
