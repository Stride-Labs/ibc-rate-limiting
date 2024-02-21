package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the ratelimit MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Adds a new rate limit. Fails if the rate limit already exists or the channel value is 0
func (k msgServer) AddRateLimit(goCtx context.Context, msg *types.MsgAddRateLimit) (*types.MsgAddRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	// Confirm the channel value is not zero
	channelValue := k.Keeper.GetChannelValue(ctx, msg.Denom)
	if channelValue.IsZero() {
		return nil, types.ErrZeroChannelValue
	}

	// Confirm the rate limit does not already exist
	_, found := k.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if found {
		return nil, types.ErrRateLimitAlreadyExists
	}

	// Confirm the channel exists
	_, found = k.Keeper.channelKeeper.GetChannel(ctx, transfertypes.PortID, msg.ChannelId)
	if !found {
		return nil, types.ErrChannelNotFound
	}

	// Create and store the rate limit object
	path := types.Path{
		Denom:     msg.Denom,
		ChannelId: msg.ChannelId,
	}
	quota := types.Quota{
		MaxPercentSend: msg.MaxPercentSend,
		MaxPercentRecv: msg.MaxPercentRecv,
		DurationHours:  msg.DurationHours,
	}
	flow := types.Flow{
		Inflow:       sdkmath.ZeroInt(),
		Outflow:      sdkmath.ZeroInt(),
		ChannelValue: channelValue,
	}

	k.Keeper.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return &types.MsgAddRateLimitResponse{}, nil
}

// Updates an existing rate limit. Fails if the rate limit doesn't exist
func (k msgServer) UpdateRateLimit(goCtx context.Context, msg *types.MsgUpdateRateLimit) (*types.MsgUpdateRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	// Confirm the rate limit exists
	_, found := k.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitNotFound
	}

	// Update the rate limit object with the new quota information
	// The flow should also get reset to 0
	path := types.Path{
		Denom:     msg.Denom,
		ChannelId: msg.ChannelId,
	}
	quota := types.Quota{
		MaxPercentSend: msg.MaxPercentSend,
		MaxPercentRecv: msg.MaxPercentRecv,
		DurationHours:  msg.DurationHours,
	}
	flow := types.Flow{
		Inflow:       sdkmath.ZeroInt(),
		Outflow:      sdkmath.ZeroInt(),
		ChannelValue: k.GetChannelValue(ctx, msg.Denom),
	}

	k.Keeper.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return &types.MsgUpdateRateLimitResponse{}, nil
}

// Removes a rate limit. Fails if the rate limit doesn't exist
func (k msgServer) RemoveRateLimit(goCtx context.Context, msg *types.MsgRemoveRateLimit) (*types.MsgRemoveRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	_, found := k.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitNotFound
	}

	k.Keeper.RemoveRateLimit(ctx, msg.Denom, msg.ChannelId)
	return &types.MsgRemoveRateLimitResponse{}, nil
}

// Resets the flow on a rate limit. Fails if the rate limit doesn't exist
func (k msgServer) ResetRateLimit(goCtx context.Context, msg *types.MsgResetRateLimit) (*types.MsgResetRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := k.Keeper.ResetRateLimit(ctx, msg.Denom, msg.ChannelId); err != nil {
		return nil, err
	}

	return &types.MsgResetRateLimitResponse{}, nil
}
