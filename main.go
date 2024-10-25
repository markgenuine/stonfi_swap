package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/nft"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	phrase := os.Getenv("SEED_PHRASE")
	if len(phrase) == 0 {
		log.Fatal("SEED_PHRASE environment variable not set")
	}

	dex := NewStonFi(phrase)

	var (
		tokenIn  *address.Address
		tokenOut *address.Address
	)

	// ton(0.5) -> jetton
	tokenIn = TonNative()
	tokenOut = address.MustParseAddr("EQAQXlWJvGbbFfE8F3oS8s87lIgdovS455IsWFaRdmJetTon") //jetton
	if err := dex.commonSwap(tokenIn, tokenOut, tlb.MustFromTON("0.5")); err != nil {
		log.Fatal("Error in commonSwap: ", err)
	}
	// ---------------

	// jetton(all balance) -> ton
	tokenIn = address.MustParseAddr("EQAQXlWJvGbbFfE8F3oS8s87lIgdovS455IsWFaRdmJetTon") //jetton
	tokenOut = TonNative()

	// get all balance jetton
	clMasterToken := jetton.NewJettonMasterClient(dex.Api, tokenIn)

	content, err := clMasterToken.GetJettonData(dex.Ctx)
	if err != nil {
		log.Fatal(err)
	}

	decimalsJetton, err := strconv.Atoi(content.Content.(*nft.ContentOnchain).GetAttribute("decimals"))
	if err != nil {
		log.Fatal(err)
	}

	//get balance JetTon and send to swap common
	jetWallet, err := clMasterToken.GetJettonWallet(dex.Ctx, dex.Wallet.WalletAddress())
	if err != nil {
		log.Fatal(err)
	}

	jettonBalance, err := jetWallet.GetBalance(dex.Ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := dex.commonSwap(tokenIn, tokenOut, tlb.MustFromNano(jettonBalance, decimalsJetton)); err != nil {
		log.Fatal("Error in commonSwap: ", err)
	}
	//---------------
}
