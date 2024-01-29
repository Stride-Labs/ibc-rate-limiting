package client

import (
	"github.com/Stride-Labs/ibc-rate-limiting/v1/ratelimit/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	AddRateLimitProposalHandler    = govclient.NewProposalHandler(cli.CmdAddRateLimitProposal)
	UpdateRateLimitProposalHandler = govclient.NewProposalHandler(cli.CmdUpdateRateLimitProposal)
	RemoveRateLimitProposalHandler = govclient.NewProposalHandler(cli.CmdRemoveRateLimitProposal)
	ResetRateLimitProposalHandler  = govclient.NewProposalHandler(cli.CmdResetRateLimitProposal)
)
