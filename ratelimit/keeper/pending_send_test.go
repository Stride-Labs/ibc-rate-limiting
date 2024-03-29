package keeper_test

import "fmt"

func (s *KeeperTestSuite) TestPendingSendPacketPrefix() {
	// Store 5 packets across two channels
	sendPackets := []string{}
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, channelId, sequence)
			sendPackets = append(sendPackets, fmt.Sprintf("%s/%d", channelId, sequence))
		}
	}

	// Check that they each sequence number is found
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			found := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence)
			s.Require().True(found, "send packet should have been found - channel %s, sequence: %d", channelId, sequence)
		}
	}

	// Check lookup of all sequence numbers
	actualSendPackets := s.App.RatelimitKeeper.GetAllPendingSendPackets(s.Ctx)
	s.Require().Equal(sendPackets, actualSendPackets, "all send packets")

	// Remove 0 sequence numbers and all sequence numbers from channel-0
	s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, "channel-0", 0)
	s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, "channel-1", 0)
	s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, "channel-0")

	// Check that only the remaining sequences are found
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			expected := (channelId == "channel-1") && (sequence != 0)
			actual := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence)
			s.Require().Equal(expected, actual, "send packet after removal - channel: %s, sequence: %d", channelId, sequence)
		}
	}
}
