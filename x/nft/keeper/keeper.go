package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/x/nft"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the nft store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	bk           nft.BankKeeper
	ac           address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(storeService store.KVStoreService,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		bk:           bk,
		ac:           ak.AddressCodec(),
	}
}

// x/nft/keeper/keeper.go (example)
func (k Keeper) generateClassID(ctx context.Context) string {
	// 1. Open store
	store := k.storeService.OpenKVStore(ctx)

	// 2. Get the current counter
	bz, err := store.Get([]byte("class_id_counter"))
	if err != nil {
		panic(err) // or handle error gracefully
	}

	var counter uint64
	if len(bz) == 0 {
		counter = 0
	} else {
		counter = sdk.BigEndianToUint64(bz)
	}

	// 3. Increment
	counter++

	// 4. Save back
	err = store.Set([]byte("class_id_counter"), sdk.Uint64ToBigEndian(counter))
	if err != nil {
		panic(err)
	}

	// 5. Produce the ID string
	return fmt.Sprintf("class-%d", counter)
}
