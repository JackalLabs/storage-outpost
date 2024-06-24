package main

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	types1 "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/stretchr/testify/require"
)

// WARNING: IBC transfers require more than one block to be confirmed as complete
// May not be able to bundle IBC transfer, with 'buy storage' in the same Tx because
// 'buy storage' requires transfer to be complete
// TODO: Attempt it in the future anyways when less time constrained

// only using for IBC transfer atm
func Broadcaster(ctx context.Context, t *testing.T, sourceChain *cosmos.CosmosChain, destChain *cosmos.CosmosChain, sender ibc.Wallet, receiver string, transferCoin sdk.Coin) {

	// Create the IBC transfer message
	transferMsg := &types.MsgTransfer{
		SourcePort:    "transfer",
		SourceChannel: "channel-1",
		Token: sdk.Coin{
			Denom:  transferCoin.Denom,
			Amount: transferCoin.Amount,
		},
		Sender:   sender.FormattedAddress(),
		Receiver: receiver,
		TimeoutHeight: types1.Height{
			RevisionNumber: 0,
			RevisionHeight: 100000000,
		},
	}
	b := cosmos.NewBroadcaster(t, sourceChain)

	txResp, err := cosmos.BroadcastTx(
		ctx,
		b,
		sender,
		transferMsg,
		// We can put a CosmWasm execute here
	)
	// WARNING: We get: 'invalid Bech32 prefix; expected cosmos, got wasm' error
	// Perhaps these functions are hard coded to require the broadcasting sender to have a cosmos prefix only
	require.NoError(t, err)
	require.NotEmpty(t, txResp.TxHash)
	fmt.Printf("IBC Transfer txResp: %+v\n", txResp)

	// Check the balance of the destination address after the transfer
	updatedBal, err := destChain.GetBalance(ctx, receiver, "ujkl")
	require.NoError(t, err)
	fmt.Println(updatedBal)

}
