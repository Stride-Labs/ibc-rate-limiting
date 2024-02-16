package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
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

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params:     types.Params{},
		RateLimits: createRateLimits(),
		WhitelistedAddressPairs: []types.WhitelistedAddressPair{
			{Sender: "sender", Receiver: "receiver"},
		},
		BlacklistedDenoms:                []string{"denomA", "denomB"},
		PendingSendPacketSequenceNumbers: []string{"channel-0/1", "channel-2/3"},
	}

	s.App.RatelimitKeeper.InitGenesis(s.Ctx, genesisState)
	got := s.App.RatelimitKeeper.ExportGenesis(s.Ctx)

	s.Require().Equal(genesisState.RateLimits, got.RateLimits)
}
