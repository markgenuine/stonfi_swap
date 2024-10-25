package structures

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type JettonTrasfer struct {
	_                   tlb.Magic        `tlb:"#0f8a7ea5"`
	QueryId             uint64           `tlb:"## 64"`
	Amount              tlb.Coins        `tlb:"."`
	Destination         *address.Address `tlb:"addr"`
	ResponseDestination *address.Address `tlb:"addr"`
	CustomPayload       *cell.Cell       `tlb:"maybe ^"`
	FwdTonAmount        tlb.Coins        `tlb:"."`
	FwdPayload          *cell.Cell       `tlb:"either . ^"`
}

type StonFiRequest struct {
	_            tlb.Magic        `tlb:"#25938561"`
	TokenWallet1 *address.Address `tlb:"addr"`
	MinOut       tlb.Coins        `tlb:"."`
	ToAddress    *address.Address `tlb:"addr"`
	HasRef       bool             `tlb:"bool"`
	RefAddress   *address.Address `tlb:"?HasRef addr"`
}
