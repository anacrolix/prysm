package balances

import (
	"reflect"
	"testing"

	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
)

func TestBaseRewardQuotient(t *testing.T) {
	if params.BeaconConfig().BaseRewardQuotient != 1<<10 {
		t.Errorf("BaseRewardQuotient should be 1024 for these tests to pass")
	}
	if params.BeaconConfig().Gwei != 1e9 {
		t.Errorf("BaseRewardQuotient should be 1e9 for these tests to pass")
	}

	tests := []struct {
		a uint64
		b uint64
	}{
		{0, 0},
		{1e6 * params.BeaconConfig().Gwei, 1024000},  //1M ETH staked, 9.76% interest.
		{2e6 * params.BeaconConfig().Gwei, 1447936},  //2M ETH staked, 6.91% interest.
		{5e6 * params.BeaconConfig().Gwei, 2289664},  //5M ETH staked, 4.36% interest.
		{10e6 * params.BeaconConfig().Gwei, 3237888}, // 10M ETH staked, 3.08% interest.
		{20e6 * params.BeaconConfig().Gwei, 4579328}, // 20M ETH staked, 2.18% interest.
	}
	for _, tt := range tests {
		b := baseRewardQuotient(tt.a)
		if b != tt.b {
			t.Errorf("BaseRewardQuotient(%d) = %d, want = %d",
				tt.a, b, tt.b)
		}
	}
}

func TestBaseReward(t *testing.T) {

	tests := []struct {
		a uint64
		b uint64
	}{
		{0, 0},
		{params.BeaconConfig().MinOnlineDepositSize * params.BeaconConfig().Gwei, 988},
		{30 * 1e9, 1853},
		{params.BeaconConfig().MaxDepositInGwei, 1976},
		{40 * 1e9, 1976},
	}
	for _, tt := range tests {
		state := &pb.BeaconState{
			ValidatorBalances: []uint64{tt.a},
		}
		// Assume 10M Eth staked (base reward quotient: 3237888).
		b := baseReward(state, 0, 3237888)
		if b != tt.b {
			t.Errorf("BaseReward(%d) = %d, want = %d",
				tt.a, b, tt.b)
		}
	}
}

func TestInactivityPenalty(t *testing.T) {

	tests := []struct {
		a uint64
		b uint64
	}{
		{1, 2929},
		{2, 3883},
		{5, 6744},
		{10, 11512},
		{50, 49659},
	}
	for _, tt := range tests {
		state := &pb.BeaconState{
			ValidatorBalances: []uint64{params.BeaconConfig().MaxDepositInGwei},
		}
		// Assume 10 ETH staked (base reward quotient: 3237888).
		b := inactivityPenalty(state, 0, 3237888, tt.a)
		if b != tt.b {
			t.Errorf("InactivityPenalty(%d) = %d, want = %d",
				tt.a, b, tt.b)
		}
	}
}

func TestFFGSrcRewardsPenalties(t *testing.T) {

	tests := []struct {
		voted                          []uint32
		balanceAfterSrcRewardPenalties []uint64
	}{
		// voted represents the validator indices that voted for FFG source,
		// balanceAfterSrcRewardPenalties represents their final balances,
		// validators who voted should get an increase, who didn't should get a decrease.
		{[]uint32{}, []uint64{31999431819, 31999431819, 31999431819, 31999431819}},
		{[]uint32{0, 1}, []uint64{32000284090, 32000284090, 31999431819, 31999431819}},
		{[]uint32{0, 1, 2, 3}, []uint64{32000568181, 32000568181, 32000568181, 32000568181}},
	}
	for _, tt := range tests {
		validatorBalances := make([]uint64, 4)
		for i := 0; i < len(validatorBalances); i++ {
			validatorBalances[i] = params.BeaconConfig().MaxDepositInGwei
		}
		state := &pb.BeaconState{
			ValidatorBalances: validatorBalances,
		}
		state = FFGSrcRewardsPenalties(
			state,
			tt.voted,
			uint64(len(tt.voted))*params.BeaconConfig().MaxDepositInGwei,
			uint64(len(validatorBalances))*params.BeaconConfig().MaxDepositInGwei)

		if !reflect.DeepEqual(state.ValidatorBalances, tt.balanceAfterSrcRewardPenalties) {
			t.Errorf("FFGSrcRewardsPenalties(%v) = %v, wanted: %v",
				tt.voted, state.ValidatorBalances, tt.balanceAfterSrcRewardPenalties)
		}
	}
}

func TestFFGTargetRewardsPenalties(t *testing.T) {

	tests := []struct {
		voted                          []uint32
		balanceAfterTgtRewardPenalties []uint64
	}{
		// voted represents the validator indices that voted for FFG target,
		// balanceAfterTgtRewardPenalties represents their final balances,
		// validators who voted should get an increase, who didn't should get a decrease.
		{[]uint32{}, []uint64{31999431819, 31999431819, 31999431819, 31999431819}},
		{[]uint32{0, 1}, []uint64{32000284090, 32000284090, 31999431819, 31999431819}},
		{[]uint32{0, 1, 2, 3}, []uint64{32000568181, 32000568181, 32000568181, 32000568181}},
	}
	for _, tt := range tests {
		validatorBalances := make([]uint64, 4)
		for i := 0; i < len(validatorBalances); i++ {
			validatorBalances[i] = params.BeaconConfig().MaxDepositInGwei
		}
		state := &pb.BeaconState{
			ValidatorBalances: validatorBalances,
		}
		state = FFGTargetRewardsPenalties(
			state,
			tt.voted,
			uint64(len(tt.voted))*params.BeaconConfig().MaxDepositInGwei,
			uint64(len(validatorBalances))*params.BeaconConfig().MaxDepositInGwei)

		if !reflect.DeepEqual(state.ValidatorBalances, tt.balanceAfterTgtRewardPenalties) {
			t.Errorf("FFGTargetRewardsPenalties(%v) = %v, wanted: %v",
				tt.voted, state.ValidatorBalances, tt.balanceAfterTgtRewardPenalties)
		}
	}
}

func TestChainHeadRewardsPenalties(t *testing.T) {

	tests := []struct {
		voted                           []uint32
		balanceAfterHeadRewardPenalties []uint64
	}{
		// voted represents the validator indices that voted for canonical chain,
		// balanceAfterHeadRewardPenalties represents their final balances,
		// validators who voted should get an increase, who didn't should get a decrease.
		{[]uint32{}, []uint64{31999431819, 31999431819, 31999431819, 31999431819}},
		{[]uint32{0, 1}, []uint64{32000284090, 32000284090, 31999431819, 31999431819}},
		{[]uint32{0, 1, 2, 3}, []uint64{32000568181, 32000568181, 32000568181, 32000568181}},
	}
	for _, tt := range tests {
		validatorBalances := make([]uint64, 4)
		for i := 0; i < len(validatorBalances); i++ {
			validatorBalances[i] = params.BeaconConfig().MaxDepositInGwei
		}
		state := &pb.BeaconState{
			ValidatorBalances: validatorBalances,
		}
		state = ChainHeadRewardsPenalties(
			state,
			tt.voted,
			uint64(len(tt.voted))*params.BeaconConfig().MaxDepositInGwei,
			uint64(len(validatorBalances))*params.BeaconConfig().MaxDepositInGwei)

		if !reflect.DeepEqual(state.ValidatorBalances, tt.balanceAfterHeadRewardPenalties) {
			t.Errorf("ChainHeadRewardsPenalties(%v) = %v, wanted: %v",
				tt.voted, state.ValidatorBalances, tt.balanceAfterHeadRewardPenalties)
		}
	}
}

func TestInclusionDistRewards(t *testing.T) {

	shardAndCommittees := []*pb.ShardAndCommitteeArray{
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 1, Committee: []uint32{0, 1, 2, 3, 4, 5, 6, 7}},
		}}}
	attestation := []*pb.PendingAttestationRecord{
		{Data: &pb.AttestationData{Shard: 1, Slot: 0},
			ParticipationBitfield: []byte{0xff},
			SlotIncluded:          5},
	}

	tests := []struct {
		voted                        []uint32
		balanceAfterInclusionRewards []uint64
	}{
		// voted represents the validator indices that voted this epoch,
		// balanceAfterInclusionRewards represents their final balances after
		// applying rewards with inclusion.
		//
		// Validators shouldn't get penalized.
		{[]uint32{}, []uint64{32000000000, 32000000000, 32000000000, 32000000000}},
		// Validators inclusion rewards are constant.
		{[]uint32{0, 1}, []uint64{32000454544, 32000454544, 32000000000, 32000000000}},
		{[]uint32{0, 1, 2, 3}, []uint64{32000454544, 32000454544, 32000454544, 32000454544}},
	}
	for _, tt := range tests {
		validatorBalances := make([]uint64, 4)
		for i := 0; i < len(validatorBalances); i++ {
			validatorBalances[i] = params.BeaconConfig().MaxDepositInGwei
		}
		state := &pb.BeaconState{
			ShardAndCommitteesAtSlots: shardAndCommittees,
			ValidatorBalances:         validatorBalances,
			LatestAttestations:        attestation,
		}
		state, err := InclusionDistRewards(
			state,
			tt.voted,
			uint64(len(validatorBalances))*params.BeaconConfig().MaxDepositInGwei)
		if err != nil {
			t.Fatalf("could not execute InclusionDistRewards:%v", err)
		}
		if !reflect.DeepEqual(state.ValidatorBalances, tt.balanceAfterInclusionRewards) {
			t.Errorf("InclusionDistRewards(%v) = %v, wanted: %v",
				tt.voted, state.ValidatorBalances, tt.balanceAfterInclusionRewards)
		}
	}
}