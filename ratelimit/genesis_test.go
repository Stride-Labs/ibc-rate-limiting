package ratelimit_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit"
	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	"github.com/Stride-Labs/ibc-rate-limiting/testing/simapp/apptesting"
)

func createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := int64(1); i <= 3; i++ {
		suffix := strconv.Itoa(int(i))
		rateLimit := types.RateLimit{
			Path:  &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
			Quota: &types.Quota{MaxPercentSend: sdkmath.NewInt(i), MaxPercentRecv: sdkmath.NewInt(i), DurationHours: uint64(i)},
			Flow:  &types.Flow{Inflow: sdkmath.NewInt(i), Outflow: sdkmath.NewInt(i), ChannelValue: sdkmath.NewInt(i)},
		}

		rateLimits = append(rateLimits, rateLimit)
	}
	return rateLimits
}

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:     types.Params{},
		RateLimits: createRateLimits(),
		WhitelistedAddressPairs: []types.WhitelistedAddressPair{
			{Sender: "sender", Receiver: "receiver"},
		},
		BlacklistedDenoms:                []string{"denomA", "denomB"},
		PendingSendPacketSequenceNumbers: []string{"channel-0/1", "channel-2/3"},
	}

	s := apptesting.SetupSuitelessTestHelper()
	ratelimit.InitGenesis(s.Ctx, s.App.RatelimitKeeper, genesisState)
	got := ratelimit.ExportGenesis(s.Ctx, s.App.RatelimitKeeper)

	require.Equal(t, genesisState.RateLimits, got.RateLimits)
}
