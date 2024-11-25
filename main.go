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
	dryRun    bool
)

func init() {
	// Check environment variable to set dryRun
	dryRun = os.Getenv("DRY_RUN") == "true"
}

func main() {
	dingToken := os.Getenv("DINGTALK_TOKEN")
	if len(dingToken) == 0 && !dryRun {
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

	addrPrice := make(map[string]string)

	// Get fear and greed index
	fearGreedResp, err := req.Get("https://api.alternative.me/fng/?limit=1")
	if err != nil {
		fmt.Printf("Failed to get fear & greed index: %v\n", err)
	}

	var fgResp FearGreedResp
	if err = fearGreedResp.ToJSON(&fgResp); err != nil {
		fmt.Printf("Failed to parse fear & greed index: %v\n", err)
	}

	// Get token prices
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

	// Add fear & greed index info
	if len(fgResp.Data) > 0 {
		sendText.WriteString(fmt.Sprintf("Fear & Greed Index: %s (%s)\n\n",
			fgResp.Data[0].Value,
			fgResp.Data[0].ValueClassification))
	}

	for addr, price := range addrPrice {
		sendText.WriteString(fmt.Sprintf("name: %s, addr: %s, price: %s\n", tokenName[addr], addr, price))
	}

	if dryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Println(sendText.String())
		fmt.Println("=== END DRY RUN ===")
	} else {
		if err := d.SendText(sendText.String()); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	}
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

type FearGreedResp struct {
	Data []struct {
		Value               string `json:"value"`
		ValueClassification string `json:"value_classification"`
		Timestamp           string `json:"timestamp"`
	} `json:"data"`
}
