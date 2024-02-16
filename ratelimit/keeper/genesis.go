package keeper

import (
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Set rate limits, blacklists, and whitelists
	for _, rateLimit := range genState.RateLimits {
		k.SetRateLimit(ctx, rateLimit)
	}
	for _, denom := range genState.BlacklistedDenoms {
		k.AddDenomToBlacklist(ctx, denom)
	}
	for _, addressPair := range genState.WhitelistedAddressPairs {
		k.SetWhitelistedAddressPair(ctx, addressPair)
	}

	// Set pending sequence numbers - validating that they're in right format of {channelId}/{sequenceNumber}
	for _, pendingPacketId := range genState.PendingSendPacketSequenceNumbers {
		splits := strings.Split(pendingPacketId, "/")
		if len(splits) != 2 {
			panic("Invalid pending send packet, must be of form: {channelId}/{sequenceNumber}")
		}
		channelId := splits[0]
		sequence, err := strconv.ParseUint(splits[1], 10, 64)
		if err != nil {
			panic(err)
		}
		k.SetPendingSendPacket(ctx, channelId, sequence)
	}

	// Verify the epoch hour duration is specified
	if genState.HourEpoch.Duration == 0 {
		panic("Hour epoch duration must be specified")
	}

	// If the hour epoch has been initialized already (epoch number != 0), validate and then use it
	if genState.HourEpoch.EpochNumber > 0 {
		if genState.HourEpoch.EpochStartTime.Equal(time.Time{}) {
			panic("If hour epoch number is non-empty, epoch time must be initialized")
		}
		if genState.HourEpoch.EpochStartHeight == 0 {
			panic("If hour epoch number is non-empty, epoch height must be initialized")
		}
		k.SetHourEpoch(ctx, genState.HourEpoch)
	} else {
		// If the hour epoch has not been initialized yet, set it so that the epoch number matches
		// the current hour and the start time is precisely on the hour
		genState.HourEpoch.EpochNumber = uint64(ctx.BlockTime().Hour())
		genState.HourEpoch.EpochStartTime = ctx.BlockTime().Truncate(time.Hour)
		genState.HourEpoch.EpochStartHeight = ctx.BlockHeight()
		k.SetHourEpoch(ctx, genState.HourEpoch)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.RateLimits = k.GetAllRateLimits(ctx)
	genesis.BlacklistedDenoms = k.GetAllBlacklistedDenoms(ctx)
	genesis.WhitelistedAddressPairs = k.GetAllWhitelistedAddressPairs(ctx)
	genesis.PendingSendPacketSequenceNumbers = k.GetAllPendingSendPackets(ctx)
	genesis.HourEpoch = k.GetHourEpoch(ctx)

	return genesis
}
