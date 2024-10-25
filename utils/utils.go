package utils

import (
	"github.com/markgenuine/stonfi_swap/structures"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func GetSwapBody(
	tokenWallet1 *address.Address,
	minOut tlb.Coins,
	toAddress *address.Address,
	hasRef bool,
	refAddress *address.Address,
) (*cell.Cell, error) {
	stonFiSwapBody := structures.StonFiRequest{
		TokenWallet1: tokenWallet1,
		MinOut:       minOut, //TODO minOut calculate with slippage ??
		ToAddress:    toAddress,
		HasRef:       hasRef,     // TODO check it
		RefAddress:   refAddress, // TODO check it
	}

	stonfiSwapBodyCell, err := tlb.ToCell(&stonFiSwapBody)
	if err != nil {
		return nil, err
	}

	return stonfiSwapBodyCell, nil
}

func GetCellTransferRequest(
	queryID uint64,
	amount tlb.Coins,
	destination *address.Address,
	responseDestination *address.Address,
	customPayload *cell.Cell,
	fwdTonAmount tlb.Coins,
	fwdPayload *cell.Cell,
) (*cell.Cell, error) {
	transferRequest := structures.JettonTrasfer{
		QueryId:             queryID,
		Amount:              amount,
		Destination:         destination,
		ResponseDestination: responseDestination,
		CustomPayload:       customPayload,
		FwdTonAmount:        fwdTonAmount,
		FwdPayload:          fwdPayload,
	}

	transferRequestCell, err := tlb.ToCell(&transferRequest)
	if err != nil {
		return nil, err
	}
	return transferRequestCell, nil
}
