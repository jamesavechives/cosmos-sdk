package keeper

import (
	"bytes"
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/nft"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ nft.MsgServer = Keeper{}

// Send implements Send method of the types.MsgServer.
func (k Keeper) Send(goCtx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	if len(msg.ClassId) == 0 {
		return nil, nft.ErrEmptyClassID
	}

	if len(msg.Id) == 0 {
		return nil, nft.ErrEmptyNFTID
	}

	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	receiver, err := k.ac.StringToBytes(msg.Receiver)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid receiver address (%s)", msg.Receiver)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	owner := k.GetOwner(ctx, msg.ClassId, msg.Id)
	if !bytes.Equal(owner, sender) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the owner of nft %s", msg.Sender, msg.Id)
	}

	if err := k.Transfer(ctx, msg.ClassId, msg.Id, receiver); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitTypedEvent(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	})
	return &nft.MsgSendResponse{}, nil
}

// / CreateClass implements the MsgServer interface for creating a new Class.
func (k Keeper) CreateClass(ctx context.Context, msg *nft.MsgCreateclass) (*nft.MsgCreateclassResponse, error) {
	// 1. Validate basic fields
	if len(msg.Name) == 0 {
		return nil, nft.ErrEmptyClassName
	}
	if len(msg.Symbol) == 0 {
		return nil, nft.ErrEmptyClassSymbol
	}
	// ... any other checks you find important

	// 2. Convert sender from Bech32 -> bytes
	_, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", msg.Sender)
	}

	// (Optional) If you only allow certain addresses to create classes, do your checks here.
	// e.g. check if senderAddr is in some whitelist, or if there's a global "class creation" fee, etc.

	// 3. Determine final class ID (either use msg.Id or generate if blank)
	classID := msg.Id
	if len(classID) == 0 {
		// If you want to auto-generate an ID, define a helper in keeper, e.g.:
		classID = k.generateClassID(ctx)
	}

	// 4. Build the Class struct from the message’s fields
	classObj := nft.Class{
		Id:          classID,
		Name:        msg.Name,
		Symbol:      msg.Symbol,
		Description: msg.Description,
		Uri:         msg.Uri,
		UriHash:     msg.UriHash,
		Data:        msg.Data, // google.protobuf.Any
	}

	// 5. Save the class to the store using keeper’s `SaveClass`
	if err := k.SaveClass(ctx, classObj); err != nil {
		return nil, err
	}

	// 7. Return a response
	return &nft.MsgCreateclassResponse{ClassId: classID}, nil
}
