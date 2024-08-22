package pool

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"pcli/contracts/ERC20"
	"pcli/contracts/IUniswapV3Factory"
	"pcli/contracts/IUniswapV3Pool"
	"pcli/contracts/PanopticFactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a new Panoptics pool",
	Long:  "Deploy a new Panoptics pool on top of Uniswap V3 pool. Deploys a Uniswap V3 pool if it does not exist.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Please provide addresses of the token pair.")
			return
		}
		deploy(args[0], args[1])
	},
}

func deploy(tokenAddr0, tokenAddr1 string) {
	// load private key
	privateKeyStr := viper.GetString("private_key")
	if privateKeyStr == "" {
		log.Fatalf("Private key not defined.")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// connect to the network client
	rpcUrl := viper.GetString("RPC_URL")
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// create a new transactor
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	// create the uniswap pool since it does not exist
	// TODO: find the appropriate gas limit
	auth.GasLimit = 500000 // NOTE: aribtrarily set

	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// instantiate uniswap v3 factory
	uniswapV3FactoryAddress := common.HexToAddress(viper.GetString("UniswapV3Factory"))

	uniswapV3Factory, err := IUniswapV3Factory.NewIUniswapV3Factory(uniswapV3FactoryAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate Uniswap V3 Factory contract: %v", err)
	}

	// check if the pool already exists
	token0 := common.HexToAddress(tokenAddr0)
	token1 := common.HexToAddress(tokenAddr1)
	fee := big.NewInt(500) // TODO: currently defaulted to 500; make it configurable

	uniswapPoolAddress, err := uniswapV3Factory.GetPool(nil, token0, token1, fee)
	if err != nil {
		log.Fatalf("Failed to get Uniswap V3 pool: %v", err)
	}

	if uniswapPoolAddress == (common.Address{}) {
		tx, err := uniswapV3Factory.CreatePool(auth, token0, token1, fee)
		if err != nil {
			log.Fatalf("Failed to create Uniswap V3 pool: %v", err)
		}
		fmt.Printf("Uniswap V3 Create Pool transaction sent: %s\n", tx.Hash().Hex())

		receipt, err := bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			log.Fatalf("Failed to get receipt: %v", err)
		}

		if receipt.Status != types.ReceiptStatusSuccessful {
			log.Fatalf("Transaction failed: %s", tx.Hash().Hex())
		}

		filterOpts := &bind.FilterOpts{
			Start:   receipt.BlockNumber.Uint64(),
			Context: context.Background(),
		}

		iterator, err := uniswapV3Factory.FilterPoolCreated(filterOpts, []common.Address{token0}, []common.Address{token1}, []*big.Int{fee})
		if err != nil {
			log.Fatalf("Failed to filter PoolCreated event: %v", err)
		}

		for iterator.Next() {
			event := iterator.Event
			poolAddress := event.Pool
			fmt.Printf("New pool created at address: %s\n", poolAddress.Hex())

			// initialize the pool with the new address
			uniswapV3Pool, err := IUniswapV3Pool.NewIUniswapV3Pool(poolAddress, client)
			if err != nil {
				log.Fatalf("Failed to create Uniswap V3 pool instance: %v", err)
			}

			// set the initial price tick
			// TODO: make this configurable
			hexValue := "0x1000000000000000000000000"
			initialPrice, _ := new(big.Int).SetString(hexValue[2:], 16)
			tx, err = uniswapV3Pool.Initialize(auth, initialPrice)
			if err != nil {
				log.Fatalf("Failed to initialize price tick: %v", err)
			}
			fmt.Printf("Price tick initialized to: %d\n", initialPrice.Int64())

			break
		}

		if err := iterator.Error(); err != nil {
			log.Fatalf("Iterator error: %v", err)
		}
	} else {
		fmt.Printf("Uniswap V3 pool already exists at address: %s\n", uniswapPoolAddress.Hex())
	}

	// deploy the panoptics pool
	panopticFactoryAddress := common.HexToAddress(viper.GetString("PanopticFactory"))

	panopticFactory, err := PanopticFactory.NewPanopticFactory(panopticFactoryAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate Panoptic Factory contract: %v", err)
	}

	panopticPoolAddress, err := panopticFactory.GetPanopticPool(nil, uniswapPoolAddress)
	if err != nil {
		log.Fatalf("Failed to get Panoptic pool: %v", err)
	}

	if panopticPoolAddress != (common.Address{}) {
		fmt.Printf("Panoptic pool already exists at address: %s\n", panopticPoolAddress.Hex())
		return
	} else {
		fmt.Println("Panoptic pool does not exist. Creating one.")

		amount := new(big.Int).SetUint64(^uint64(0)) // equivalent to type(uint256).max in Solidity

		if err := approveToken(client, auth, token0, panopticFactoryAddress, amount); err != nil {
			log.Fatalf("Failed to approve token %s: %v", tokenAddr0, err)
		}

		if err := approveToken(client, auth, token1, panopticFactoryAddress, amount); err != nil {
			log.Fatalf("Failed to approve token %s: %v", tokenAddr1, err)
		}

		salt := deriveSalt(token0, token1, fee)
		tx, err := panopticFactory.DeployNewPool(auth, token0, token1, fee, salt)

		if err != nil {
			log.Fatalf("Failed to deploy Panoptics pool: %v", err)
		}

		receipt, err := bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			log.Fatalf("Failed to get transaction receipt: %v", err)
		}

		if receipt.Status == 1 {
			fmt.Printf("Transaction was successful: %v\n", receipt.TxHash.Hex())
		} else {
			fmt.Printf("Transaction failed: %v\n", receipt.TxHash.Hex())
		}
	}
}

func approveToken(client *ethclient.Client, auth *bind.TransactOpts, tokenAddress, spender common.Address, amount *big.Int) error {
	tokenContract, err := ERC20.NewERC20(tokenAddress, client)
	if err != nil {
		return fmt.Errorf("Failed to instantiate token: %s", tokenAddress.Hex())
	}

	tx, err := tokenContract.Approve(auth, spender, amount)
	if err != nil {
		return fmt.Errorf("failed to approve token: %w", err)
	}

	fmt.Printf("Approval transaction sent: %s for token %s\n", tx.Hash().Hex(), tokenAddress)
	return nil
}

func deriveSalt(token0, token1 common.Address, fee *big.Int) *big.Int {
	hash := crypto.Keccak256Hash(token0.Bytes(), token1.Bytes(), common.BigToHash(fee).Bytes())
	return new(big.Int).SetBytes(hash.Bytes()[:12])
}
