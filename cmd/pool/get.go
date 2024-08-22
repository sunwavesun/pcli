package pool

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"pcli/contracts/IUniswapV3Factory"
	"pcli/contracts/PanopticFactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Panoptic pool",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Please provide addresses of the token pair.")
			return
		}
		Get(args[0], args[1])
	},
}

func Get(tokenAddr0, tokenAddr1 string) {
	client, err := ConnectToClient()
	if err != nil {
		log.Fatalf("%v", err)
	}

	uniswapV3FactoryAddress := common.HexToAddress(viper.GetString("UniswapV3Factory"))
	uniswapV3Factory, err := IUniswapV3Factory.NewIUniswapV3Factory(uniswapV3FactoryAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate Uniswap V3 Factory contract: %v", err)
	}

	token0 := common.HexToAddress(tokenAddr0)
	token1 := common.HexToAddress(tokenAddr1)
	fee := big.NewInt(500) // 0.5% fee

	uniswapPoolAddress, err := GetUniswapPool(uniswapV3Factory, token0, token1, fee)
	if err != nil {
		log.Fatalf("%v", err)
	}

	panopticFactoryAddress := common.HexToAddress(viper.GetString("PanopticFactory"))
	panopticFactory, err := PanopticFactory.NewPanopticFactory(panopticFactoryAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate Panoptic Factory contract: %v", err)
	}

	panopticPoolAddress, err := GetPanopticPool(panopticFactory, uniswapPoolAddress)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println(panopticPoolAddress.Hex())
}

// func get(tokenAddr0, tokenAddr1 string) {
// 	// load private key
// 	privateKeyStr := viper.GetString("private_key")
// 	if privateKeyStr == "" {
// 		log.Fatalf("Private key not defined.")
// 	}

// 	privateKey, err := crypto.HexToECDSA(privateKeyStr)
// 	if err != nil {
// 		log.Fatalf("Failed to load private key: %v", err)
// 	}

// 	// connect to the network client
// 	rpcUrl := viper.GetString("RPC_URL")
// 	client, err := ethclient.Dial(rpcUrl)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
// 	}

// 	// create a new transactor
// 	chainID, err := client.NetworkID(context.Background())
// 	if err != nil {
// 		log.Fatalf("Failed to get network ID: %v", err)
// 	}

// 	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
// 	// create the uniswap pool since it does not exist
// 	// TODO: find the appropriate gas limit
// 	auth.GasLimit = 500000 // NOTE: aribtrarily set

// 	if err != nil {
// 		log.Fatalf("Failed to create authorized transactor: %v", err)
// 	}

// 	// instantiate uniswap v3 factory
// 	uniswapV3FactoryAddress := common.HexToAddress(viper.GetString("UniswapV3Factory"))

// 	uniswapV3Factory, err := IUniswapV3Factory.NewIUniswapV3Factory(uniswapV3FactoryAddress, client)
// 	if err != nil {
// 		log.Fatalf("Failed to instantiate Uniswap V3 Factory contract: %v", err)
// 	}

// 	// check if the pool already exists
// 	token0 := common.HexToAddress(tokenAddr0)
// 	token1 := common.HexToAddress(tokenAddr1)
// 	fee := big.NewInt(500) // TODO: currently defaulted to 500; make it configurable

// 	uniswapPoolAddress, err := uniswapV3Factory.GetPool(nil, token0, token1, fee)
// 	if err != nil {
// 		log.Fatalf("Failed to get Uniswap V3 pool: %v", err)
// 	}

// 	if uniswapPoolAddress == (common.Address{}) {
// 		log.Fatalf("Uniswap V3 pool for the token pair does not exist.")
// 	}

// 	panopticFactoryAddress := common.HexToAddress(viper.GetString("PanopticFactory"))
// 	panopticFactory, err := PanopticFactory.NewPanopticFactory(panopticFactoryAddress, client)
// 	if err != nil {
// 		log.Fatalf("Failed to instantiate Panoptic Factory contract: %v", err)
// 	}

// 	panopticPoolAddress, err := panopticFactory.GetPanopticPool(nil, uniswapPoolAddress)
// 	if err != nil {
// 		log.Fatalf("Failed to get Panoptic pool: %v", err)
// 	}

// 	if panopticPoolAddress == (common.Address{}) {
// 		fmt.Println("Panoptic pool does not exist.")
// 	}

// 	fmt.Println(panopticPoolAddress)
// }

func LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	privateKeyStr := viper.GetString("private_key")
	if privateKeyStr == "" {
		return nil, fmt.Errorf("Private key not defined")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to load private key: %v", err)
	}

	return privateKey, nil
}

func ConnectToClient() (*ethclient.Client, error) {
	rpcUrl := viper.GetString("RPC_URL")
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the Ethereum client: %v", err)
	}

	return client, nil
}

func CreateTransactor(client *ethclient.Client, privateKey *ecdsa.PrivateKey, gasLimit uint64) (*bind.TransactOpts, error) {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to get network ID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("Failed to create authorized transactor: %v", err)
	}

	auth.GasLimit = gasLimit
	return auth, nil
}

func GetUniswapPool(factory *IUniswapV3Factory.IUniswapV3Factory, token0, token1 common.Address, fee *big.Int) (common.Address, error) {
	uniswapPoolAddress, err := factory.GetPool(nil, token0, token1, fee)
	if err != nil {
		return common.Address{}, fmt.Errorf("Failed to get Uniswap V3 pool: %v", err)
	}

	if uniswapPoolAddress == (common.Address{}) {
		return common.Address{}, fmt.Errorf("Uniswap V3 pool for the token pair does not exist")
	}

	return uniswapPoolAddress, nil
}

func GetPanopticPool(factory *PanopticFactory.PanopticFactory, uniswapPoolAddress common.Address) (common.Address, error) {
	panopticPoolAddress, err := factory.GetPanopticPool(nil, uniswapPoolAddress)
	if err != nil {
		return common.Address{}, fmt.Errorf("Failed to get Panoptic pool: %v", err)
	}

	if panopticPoolAddress == (common.Address{}) {
		fmt.Println("Panoptic pool does not exist.")
	}

	return panopticPoolAddress, nil
}
