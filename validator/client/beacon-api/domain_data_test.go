//go:build use_beacon_api
// +build use_beacon_api

package beacon_api

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	rpcmiddleware "github.com/prysmaticlabs/prysm/v3/beacon-chain/rpc/apimiddleware"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/testing/assert"
	"github.com/prysmaticlabs/prysm/v3/testing/require"
	"github.com/prysmaticlabs/prysm/v3/validator/client/beacon-api/mock"
)

func TestGetDomainData_ValidDomainData(t *testing.T) {
	const genesisValidatorRoot = "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"
	forkVersion := params.BeaconConfig().AltairForkVersion
	epoch := params.BeaconConfig().AltairForkEpoch
	domainType := params.BeaconConfig().DomainBeaconProposer

	genesisValidatorRootBytes, err := hexutil.Decode(genesisValidatorRoot)
	require.NoError(t, err)

	expectedForkDataRoot, err := (&ethpb.ForkData{
		CurrentVersion:        forkVersion,
		GenesisValidatorsRoot: genesisValidatorRootBytes,
	}).HashTreeRoot()
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Make sure that GetGenesis() is called exactly once
	genesisProvider := mock.NewMockgenesisProvider(ctrl)
	genesisProvider.EXPECT().GetGenesis().Return(
		&rpcmiddleware.GenesisResponse_GenesisJson{GenesisValidatorsRoot: genesisValidatorRoot},
		nil,
		nil,
	).Times(1)

	validatorClient := &beaconApiValidatorClient{genesisProvider: genesisProvider}
	resp, err := validatorClient.getDomainData(epoch, domainType)
	assert.NoError(t, err)
	require.NotNil(t, resp)

	var expectedSignatureDomain []byte
	expectedSignatureDomain = append(expectedSignatureDomain, domainType[:]...)
	expectedSignatureDomain = append(expectedSignatureDomain, expectedForkDataRoot[:28]...)

	assert.Equal(t, len(expectedSignatureDomain), len(resp.SignatureDomain))
	assert.DeepEqual(t, expectedSignatureDomain, resp.SignatureDomain)
}

func TestGetDomainData_GenesisError(t *testing.T) {
	const genesisValidatorRoot = "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"
	epoch := params.BeaconConfig().AltairForkEpoch
	domainType := params.BeaconConfig().DomainBeaconProposer

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Make sure that GetGenesis() is called exactly once
	genesisProvider := mock.NewMockgenesisProvider(ctrl)
	genesisProvider.EXPECT().GetGenesis().Return(nil, nil, errors.New("")).Times(1)

	validatorClient := &beaconApiValidatorClient{genesisProvider: genesisProvider}
	_, err := validatorClient.getDomainData(epoch, domainType)
	assert.ErrorContains(t, "failed to get genesis info", err)
}

func TestGetDomainData_InvalidGenesisRoot(t *testing.T) {
	const genesisValidatorRoot = "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"
	epoch := params.BeaconConfig().AltairForkEpoch
	domainType := params.BeaconConfig().DomainBeaconProposer

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Make sure that GetGenesis() is called exactly once
	genesisProvider := mock.NewMockgenesisProvider(ctrl)
	genesisProvider.EXPECT().GetGenesis().Return(
		&rpcmiddleware.GenesisResponse_GenesisJson{GenesisValidatorsRoot: "foo"},
		nil,
		nil,
	).Times(1)

	validatorClient := &beaconApiValidatorClient{genesisProvider: genesisProvider}
	_, err := validatorClient.getDomainData(epoch, domainType)
	assert.ErrorContains(t, "invalid genesis validators root: foo", err)
}
