package gen

import (
	"pgregory.net/rand"

	"github.com/Fantom-foundation/Tosca/go/ct/common"
	"github.com/Fantom-foundation/Tosca/go/ct/st"
)

type BlockContextGenerator struct {
}

func NewBlockContextGenerator() *BlockContextGenerator {
	return &BlockContextGenerator{}
}

func (*BlockContextGenerator) Generate(rnd *rand.Rand, revision common.Revision) (st.BlockContext, error) {
	baseFee := common.RandU256(rnd)

	revisionNumber, err := common.GetForkBlock(revision)
	if err != nil {
		return st.BlockContext{}, err
	}
	revisionNumberRange, err := common.GetBlockRangeLengthFor(revision)
	if err != nil {
		return st.BlockContext{}, err
	}
	var randomOffset uint64
	if revisionNumberRange != 0 {
		randomOffset = rnd.Uint64n(revisionNumberRange)
	} else {
		randomOffset = rnd.Uint64()
	}
	blockNumber := revisionNumber + randomOffset

	chainId := common.RandU256(rnd)
	coinbase, err := common.RandAddress(rnd)
	if err != nil {
		return st.BlockContext{}, err
	}
	gasLimit := rnd.Uint64()
	gasPrice := common.RandU256(rnd)

	difficulty := common.RandU256(rnd)
	timestamp := rnd.Uint64()

	return st.BlockContext{
		BaseFee:     baseFee,
		BlockNumber: blockNumber,
		ChainID:     chainId,
		CoinBase:    coinbase,
		GasLimit:    gasLimit,
		GasPrice:    gasPrice,
		Difficulty:  difficulty,
		TimeStamp:   timestamp,
	}, nil
}

func (*BlockContextGenerator) Clone() *BlockContextGenerator {
	return &BlockContextGenerator{}
}

func (*BlockContextGenerator) Restore(*BlockContextGenerator) {
}

func (*BlockContextGenerator) String() string {
	return "{}"
}