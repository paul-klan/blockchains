package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

var (
	url        = "https://mainnet.infura.io/v3/f3ce4332806149acaf367c9a25ac1bc1"
	address721 = "0x306b1ea3ecdf94ab739f1910bbda052ed4a9f949"
	tokenId721 = "17412"

	address1155 = "0x9ca3a9a3aa59c7ddd61c29f6b0540ad9988aede6"
	tokenId1155 = "5"
)

func getUrl(input []byte) string {
	if len(input) <= 2 {
		return ""
	}

	if input[0] == '0' && (input[1] == 'x' || input[1] == 'X') {
		input = input[2:]
	}

	code := make([]byte, 0, len(input))
	for _, in := range input {
		if in == 0 {
			continue
		}

		if in == 32 {
			continue
		}
		code = append(code, in)
	}

	if len(code) == 0 {
		return ""
	}

	cut := int(code[0]) + 1
	log.Infof("len %v, cut %s", cut, string(code[1:cut]))

	out := string(code[1:cut])
	return out
}

func getContractTokenUri(ctx context.Context, client *ethclient.Client, address, tokenId string, is721 bool) (string, error) {
	to := common.HexToAddress(address)
	tokenCode, _ := new(big.Int).SetString(tokenId, 10)

	var arg []byte
	var err error
	if is721 {
		arg, err = hexutil.Decode("0xc87b56dd" + fmt.Sprintf("%064X", tokenCode.Int64()))
	} else {
		arg, err = hexutil.Decode("0x0e89341c" + fmt.Sprintf("%064X", tokenCode.Int64()))
	}

	if err != nil {
		log.Errorf("hex decode parameter error, %v", err)
		return "", err
	}

	msg := ethereum.CallMsg{
		To: &to,
		//Data: "0xc87b56dd" + "0000000000000000000000000000000000000000000000000000000000000002"
		Data: arg,
	}
	log.Infof("call data %x", msg.Data)

	outUrl, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Fatalf("call contract error, %v", err)
	}
	return getUrl(outUrl), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		log.Errorf("dail chain error, %s, %v", url, err)
		panic(err)
	}

	// erc721
	outUrl1, err := getContractTokenUri(ctx, client, address721, tokenId721, true)
	if err != nil {
		log.Errorf("get 721 url error, %s, %v", outUrl1, err)
		panic(err)
	}
	log.Infof("erc721 >> %s", outUrl1)

	// erc721
	outUrl1155, err := getContractTokenUri(ctx, client, address1155, tokenId1155, false)
	if err != nil {
		log.Errorf("get 721 url error, %s, %v", outUrl1, err)
		panic(err)
	}
	log.Infof("erc1155 >> %s", outUrl1155)

	log.Info("test quit")
}
