package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Before each hour epoch, check if any of the rate limits have expired,
// and reset them if they have
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	if epochStarting, epochNumber := k.CheckHourEpochStarting(ctx); epochStarting {
		for _, rateLimit := range k.GetAllRateLimits(ctx) {
			if rateLimit.Quota.DurationHours != 0 && epochNumber%rateLimit.Quota.DurationHours == 0 {
				err := k.ResetRateLimit(ctx, rateLimit.Path.Denom, rateLimit.Path.ChannelId)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Unable to reset quota for Denom: %s, ChannelId: %s", rateLimit.Path.Denom, rateLimit.Path.ChannelId))
				}
			}
		}
	}
}

// Checks if it's time to start the new hour epoch
func (k Keeper) CheckHourEpochStarting(ctx sdk.Context) (epochStarting bool, epochNumber uint64) {
	hourEpoch := k.GetHourEpoch(ctx)

	// If the block time is later than the current epoch start time + epoch duration,
	// move onto the next epoch by incrementing the epoch number, height, and start time
	currentEpochEndTime := hourEpoch.EpochStartTime.Add(hourEpoch.Duration)
	shouldNextEpochStart := ctx.BlockTime().After(currentEpochEndTime)
	if shouldNextEpochStart {
		hourEpoch.EpochNumber++
		hourEpoch.EpochStartTime = currentEpochEndTime
		hourEpoch.EpochStartHeight = ctx.BlockHeight()

		k.SetHourEpoch(ctx, hourEpoch)
		return true, hourEpoch.EpochNumber
	}

	// Otherwise, indicate that a new epoch is not starting
	return false, 0
}
