package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"strings"

	"github.com/markgenuine/stonfi_swap/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type Router struct {
	Api ton.APIClientWrapped
	Ctx context.Context
}

type Fee struct {
	TxTon      *big.Int
	ForwardTon *big.Int
}

type Fees struct {
	TonToJetton    Fee
	JettonToJetton Fee
	JettonToTon    Fee
}

type StonFi struct {
	Router       *Router
	Wallet       *wallet.Wallet
	Ctx          context.Context
	Api          ton.APIClientWrapped
	StonFiRouter *address.Address
	*Fees
}

func GetRouterAddress() *address.Address {
	return address.MustParseAddr("EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt")
}

func TonNative() *address.Address {
	return address.MustParseAddr("EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c")
}

func GetPtonWalletAddress() *address.Address {
	return address.MustParseAddr("EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC")
}

func NewRouter(api ton.APIClientWrapped, ctx context.Context) *Router {
	return &Router{Api: api, Ctx: ctx}
}

func NewStonFi(phrase string) *StonFi {
	client := liteclient.NewConnectionPool()

	cfg, err := liteclient.GetConfigFromUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		log.Fatalln("get config err: ", err.Error())
	}

	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalln("connection err: ", err.Error())

	}

	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	ctx := client.StickyContext(context.Background())
	w, err := wallet.FromSeed(api, strings.Split(phrase, " "), wallet.V4R2)
	if err != nil {
		log.Fatalln("FromSeed err:", err.Error())
	}

	fees := &Fees{
		TonToJetton: Fee{
			TxTon:      tlb.MustFromTON("0.215").Nano(),
			ForwardTon: tlb.MustFromTON("0.215").Nano(),
		},
		JettonToJetton: Fee{
			TxTon:      tlb.MustFromTON("0.265").Nano(),
			ForwardTon: tlb.MustFromTON("0.205").Nano(),
		}, JettonToTon: Fee{
			TxTon:      tlb.MustFromTON("0.185").Nano(),
			ForwardTon: tlb.MustFromTON("0.125").Nano(),
		},
	}

	return &StonFi{
		Router:       NewRouter(api, ctx),
		Wallet:       w,
		Ctx:          ctx,
		Api:          api,
		StonFiRouter: GetRouterAddress(),
		Fees:         fees,
	}
}

func (sF *StonFi) commonSwap(tokenIn *address.Address, tokenOut *address.Address, amount tlb.Coins) error {
	err := errors.New("invalid swap strategy")

	nativeTon := TonNative().String()
	if tokenIn.String() == nativeTon && tokenOut.String() != nativeTon {
		err = sF.swapTonToJetton(tokenOut, amount)
	} else if tokenIn.String() != nativeTon && tokenOut.String() == nativeTon {
		err = sF.swapJettonToTon(tokenIn, amount)
	}

	return err
}

func (sF *StonFi) swapTonToJetton(tokenOut *address.Address, amount tlb.Coins) error {
	masterTokenOut := jetton.NewJettonMasterClient(sF.Api, tokenOut)
	tokeOutWallet, err := masterTokenOut.GetJettonWallet(sF.Ctx, sF.StonFiRouter)
	if err != nil {
		return err
	}

	stonfiSwapBodyCell, err := utils.GetSwapBody(
		tokeOutWallet.Address(),
		tlb.MustFromTON("0"), //TODO minOut calculate with slippage ??
		sF.Wallet.WalletAddress(),
		false, // TODO check it
		nil,   // TODO check it
	)
	if err != nil {
		return err
	}

	transferRequestCell, err := utils.GetCellTransferRequest(
		rand.Uint64(), amount, sF.StonFiRouter, sF.StonFiRouter, nil,
		tlb.FromNanoTON(sF.TonToJetton.ForwardTon), stonfiSwapBodyCell)

	if err != nil {
		return err
	}

	// TODO very slow and not information about txhash for check
	if err := sF.Wallet.Send(
		sF.Ctx,
		wallet.SimpleMessage(
			GetPtonWalletAddress(),
			tlb.FromNanoTON(new(big.Int).Add(amount.Nano(), sF.TonToJetton.ForwardTon)),
			transferRequestCell,
		),
		true,
	); err != nil {
		return err
	}

	return nil
}

func (sF *StonFi) swapJettonToTon(tokenIn *address.Address, amount tlb.Coins) error {
	tokenInMaster := jetton.NewJettonMasterClient(sF.Api, tokenIn)
	tokenInSpender, err := tokenInMaster.GetJettonWallet(sF.Ctx, sF.Wallet.WalletAddress())
	if err != nil {
		return err
	}

	stonfiSwapBodyCell, err := utils.GetSwapBody(
		GetPtonWalletAddress(),
		tlb.MustFromTON("0"), //TODO add slippage
		sF.Wallet.WalletAddress(),
		false, // TODO ref?
		nil,   // TODO ref?
	)
	if err != nil {
		return err
	}

	content, err := tokenInMaster.GetJettonData(sF.Ctx)
	if err != nil {
		return err
	}

	decimalsJetton, err := strconv.Atoi(content.Content.(*nft.ContentOnchain).GetAttribute("decimals"))
	if err != nil {
		return err
	}

	transferRequestCell, err := utils.GetCellTransferRequest(
		rand.Uint64(),
		tlb.MustFromNano(amount.Nano(), decimalsJetton),
		sF.StonFiRouter,
		sF.Wallet.WalletAddress(),
		nil,
		tlb.FromNanoTON(sF.JettonToTon.ForwardTon),
		stonfiSwapBodyCell,
	)
	if err != nil {
		return err
	}

	if err := sF.Wallet.Send(
		sF.Ctx,
		wallet.SimpleMessage(
			tokenInSpender.Address(),
			tlb.FromNanoTON(sF.JettonToTon.TxTon),
			transferRequestCell,
		),
		true,
	); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
