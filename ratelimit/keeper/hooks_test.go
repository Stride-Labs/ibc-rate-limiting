package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	epochstypes "github.com/Stride-Labs/stride/v17/x/epochs/types"
)

// Store a rate limit with a non-zero flow for each duration
func (s *KeeperTestSuite) resetRateLimits(denom string, durations []uint64, nonZeroFlow int64) {
	// Add/reset rate limit with a quota duration hours for each duration in the list
	for i, duration := range durations {
		channelId := fmt.Sprintf("channel-%d", i)

		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
			Path: &types.Path{
				Denom:     denom,
				ChannelId: channelId,
			},
			Quota: &types.Quota{
				DurationHours: duration,
			},
			Flow: &types.Flow{
				Inflow:       sdkmath.NewInt(nonZeroFlow),
				Outflow:      sdkmath.NewInt(nonZeroFlow),
				ChannelValue: sdkmath.NewInt(100),
			},
		})
	}
}

func (s *KeeperTestSuite) TestBeforeEpochStart() {
	// We'll create three rate limits with different durations
	// And then pass in epoch ids that will cause each to trigger a reset in order
	// i.e. epochId 2   will only cause duration 2 to trigger (2 % 2 == 0; and 9 % 2 != 0; 25 % 2 != 0),
	//      epochId 9,  will only cause duration 3 to trigger (9 % 2 != 0; and 9 % 3 == 0; 25 % 3 != 0)
	//      epochId 25, will only cause duration 5 to trigger (9 % 5 != 0; and 9 % 5 != 0; 25 % 5 == 0)
	durations := []uint64{2, 3, 5}
	epochIds := []int64{2, 9, 25}
	nonZeroFlow := int64(10)

	for i, epochId := range epochIds {
		// First reset the  rate limits to they have a non-zero flow
		s.resetRateLimits(denom, durations, nonZeroFlow)

		duration := durations[i]
		channelIdFromResetRateLimit := fmt.Sprintf("channel-%d", i)

		// Then trigger the epoch hook
		epoch := epochstypes.EpochInfo{
			Identifier:   epochstypes.HOUR_EPOCH,
			CurrentEpoch: epochId,
		}
		s.App.RatelimitKeeper.BeginBlocker(s.Ctx, epoch)

		// Check rate limits (only one rate limit should reset for each hook trigger)
		rateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
		for _, rateLimit := range rateLimits {
			context := fmt.Sprintf("duration: %d, epoch: %d", duration, epochId)

			if rateLimit.Path.ChannelId == channelIdFromResetRateLimit {
				s.Require().Equal(int64(0), rateLimit.Flow.Inflow.Int64(), "inflow was not reset to 0 - %s", context)
				s.Require().Equal(int64(0), rateLimit.Flow.Outflow.Int64(), "outflow was not reset to 0 - %s", context)
			} else {
				s.Require().Equal(nonZeroFlow, rateLimit.Flow.Inflow.Int64(), "inflow should have been left unchanged - %s", context)
				s.Require().Equal(nonZeroFlow, rateLimit.Flow.Outflow.Int64(), "outflow should have been left unchanged - %s", context)
			}
		}
	}
}

func (s *KeeperTestSuite) TestCheckHourEpochStarting() {
	epochStartTime := time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC)
	blockHeight := int64(10)
	duration := time.Minute

	initialEpoch := types.HourEpoch{
		EpochNumber:    10,
		EpochStartTime: epochStartTime,
		Duration:       duration,
	}
	nextEpoch := types.HourEpoch{
		EpochNumber:      initialEpoch.EpochNumber + 1, // epoch number increments
		EpochStartTime:   epochStartTime.Add(duration), // start time increments by duration
		EpochStartHeight: blockHeight,                  // height gets current block height
		Duration:         duration,
	}

	testCases := []struct {
		name                  string
		blockTime             time.Time
		expectedEpochStarting bool
	}{
		{
			name:                  "in middle of epoch",
			blockTime:             epochStartTime.Add(duration / 2), // halfway through epoch
			expectedEpochStarting: false,
		},
		{
			name:                  "right before epoch boundary",
			blockTime:             epochStartTime.Add(duration).Add(-1 * time.Second), // 1 second before epoch
			expectedEpochStarting: false,
		},
		{
			name:                  "at epoch boundary",
			blockTime:             epochStartTime.Add(duration), // at epoch boundary
			expectedEpochStarting: false,
		},
		{
			name:                  "right after epoch boundary",
			blockTime:             epochStartTime.Add(duration).Add(time.Second), // one second after epoch boundary
			expectedEpochStarting: true,
		},
		{
			name:                  "in middle of next epoch",
			blockTime:             epochStartTime.Add(duration).Add(duration / 2), // halfway through next epoch
			expectedEpochStarting: true,
		},
		{
			name:                  "next epoch skipped",
			blockTime:             epochStartTime.Add(duration * 10), // way after next epoch (still increments only once)
			expectedEpochStarting: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Ctx = s.Ctx.WithBlockTime(tc.blockTime)
			s.Ctx = s.Ctx.WithBlockHeight(blockHeight)

			s.App.RatelimitKeeper.SetHourEpoch(s.Ctx, initialEpoch)

			actualStarting, actualEpochNumber := s.App.RatelimitKeeper.CheckHourEpochStarting(s.Ctx)
			s.Require().Equal(tc.expectedEpochStarting, actualStarting, "epoch starting")

			expectedEpoch := initialEpoch
			if tc.expectedEpochStarting {
				expectedEpoch = nextEpoch
				s.Require().Equal(expectedEpoch.EpochNumber, actualEpochNumber, "epoch number")
			}

			actualHourEpoch := s.App.RatelimitKeeper.GetHourEpoch(s.Ctx)
			s.Require().Equal(expectedEpoch, actualHourEpoch, "hour epoch")
		})
	}
}
