package option

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var mindCmt = &cobra.Command{
	Use:   "mint",
	Short: "Mint a new option",
	Run: func(cmd *cobra.Command, args []string) {
		mint()
	},
}

func mint() {
	// // create a PanopticPool instance

	// // instantiate token contracts
	// tokenAddr0 := ""
	// tokenAddr1 := ""

	// // approve tokens to the token collateral tracker
	// // determine the amount

	// // deposit provided amount to the collateral tracker

	// // receive inputs in order to create legs

	// // mint the option
	panopticFactoryAddress := common.HexToAddress("0xYourPanopticFactoryAddress")
	token0Address := common.HexToAddress("0x29f2D40B0605204364af54EC677bD022dA425d03")
	token1Address := common.HexToAddress("0xFF34B3d4Aee8ddCd6F9AFFFB6Fe49bD371b8a357")

	panopticFactoryContract, err := ethclient.ContractAt(panopticFactoryABI, panopticFactoryAddress, client)
	if err != nil {
		log.Fatalf("Failed to create PanopticFactory contract instance: %v", err)
	}
	token0Contract, err := ethclient.ContractAt(ierc20PartialABI, token0Address, client)
	if err != nil {
		log.Fatalf("Failed to create token0 contract instance: %v", err)
	}
	token1Contract, err := ethclient.ContractAt(ierc20PartialABI, token1Address, client)
	if err != nil {
		log.Fatalf("Failed to create token1 contract instance: %v", err)
	}

	tx, err := panopticFactoryContract.Transact(
		"deployNewPool",
		token0Address,
		token1Address,
		big.NewInt(500),
		big.NewInt(1337),
	)
	if err != nil {
		log.Fatalf("Failed to deploy new pool: %v", err)
	}

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Fatalf("Failed to get transaction receipt: %v", err)
	}
	panopticPoolAddress := common.HexToAddress(receipt.ContractAddress.Hex())

	panopticPoolContract, err := ethclient.ContractAt(panopticPoolABI, panopticPoolAddress, client)
	if err != nil {
		log.Fatalf("Failed to create PanopticPool contract instance: %v", err)
	}

	approveAmount := new(big.Int).SetUint64(1<<64 - 1)
	_, err = token0Contract.Transact("approve", panopticPoolAddress, approveAmount)
	if err != nil {
		log.Fatalf("Failed to approve token0: %v", err)
	}

	_, err = token1Contract.Transact("approve", panopticPoolAddress, approveAmount)
	if err != nil {
		log.Fatalf("Failed to approve token1: %v", err)
	}

	depositAmounttoken0 := new(big.Int).SetUint64(10*10 ^ 8)
	depositAmounttoken1 := new(big.Int).SetUint64(100*10 ^ 18)

	_, err = panopticPoolContract.Transact("collateralToken0Deposit", depositAmounttoken0, panopticPoolAddress)
	if err != nil {
		log.Fatalf("Failed to deposit token0: %v", err)
	}

	_, err = panopticPoolContract.Transact("collateralToken1Deposit", depositAmounttoken1, panopticPoolAddress)
	if err != nil {
		log.Fatalf("Failed to deposit token1: %v", err)
	}

	legIndex := uint8(0)
	optionRatio := big.NewInt(1)
	asset := big.NewInt(1)
	isLong := uint8(0)
	tokenType := uint8(1)
	riskPartner := uint8(0)
	strike := big.NewInt(-5000)
	width := big.NewInt(2)

	_, err = panopticPoolContract.Transact(
		"addLeg",
		legIndex,
		optionRatio,
		asset,
		isLong,
		tokenType,
		riskPartner,
		strike,
		width,
	)
	if err != nil {
		log.Fatalf("Failed to add leg: %v", err)
	}

	positionIdList := []*big.Int{big.NewInt(0)}
	positionSize := toBigInt("10000000000000000000")
	effectiveLiquidityLimitX32 := big.NewInt(0)
	tickLimitLow := big.NewInt(0)
	tickLimitHigh := big.NewInt(0)

	_, err = panopticPoolContract.Transact(
		"mintOptions",
		positionIdList,
		positionSize,
		effectiveLiquidityLimitX32,
		tickLimitLow,
		tickLimitHigh,
	)
	if err != nil {
		log.Fatalf("Failed to mint options: %v", err)
	}

	fmt.Println("Options minted successfully.")
}

func toBigInt(value string) *big.Int {
	result := new(big.Int)
	result.SetString(value, 10)
	return result
}
