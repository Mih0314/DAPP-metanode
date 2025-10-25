package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/Mih0314/DAPP-metanode/task2/counter"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/e2e910c5ab1047daba077491d3c52ec6")
	if err != nil {
		log.Fatal(err)
	}

	// privateKey, err := crypto.GenerateKey()
	// privateKeyBytes := crypto.FromECDSA(privateKey)
	// privateKeyHex := hex.EncodeToString(privateKeyBytes)
	// fmt.Println("Private Key:", privateKeyHex)
	privateKey, err := crypto.HexToECDSA("6fba365d0eb5b9d631a83f84efb087381bcaa791f4be5e105ab8906d7cba164f")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := big.NewInt(3)
	address, tx, instance, err := counter.DeployCounter(auth, client, input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(address.Hex())
	fmt.Println(tx.Hash().Hex())

	// 关键修复：等待部署交易确认（必须！）
	fmt.Println("等待合约部署交易确认...（约10-30秒）")
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("部署交易确认失败：%v", err)
	}

	auth.Nonce.Add(auth.Nonce, big.NewInt(1))

	// 检查交易是否成功
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Fatal("合约部署失败（交易被 revert）")
	}

	// 从收据中获取最终的合约地址（与 address 一致，但更可靠）
	contractAddr := receipt.ContractAddress
	fmt.Println("合约部署成功，地址：", contractAddr.Hex())

	// _ = instance

	fmt.Println("=== 调用 getCount() 初始值 ===")
	count, err := instance.GetCount(&bind.CallOpts{}) // 读方法用 CallOpts
	if err != nil {
		log.Fatalf("调用 getCount 失败：%v", err)
	}
	fmt.Printf("当前 count 值：%d\n", count)

	// 6. 调用写方法 incr()（需要签名，消耗 gas，会生成交易）
	fmt.Println("\n=== 调用 incr() 方法 ===")
	tx1, err := instance.Incr(auth) // 写方法用前面创建的 auth
	if err != nil {
		log.Fatalf("调用 incr 失败：%v", err)
	}
	fmt.Printf("incr 交易哈希：%s\n", tx1.Hash().Hex())
	fmt.Println("等待 incr 交易确认...（约10-30秒）")
	// 等待交易确认（确保状态已更新）
	_, err = bind.WaitMined(context.Background(), client, tx1)
	if err != nil {
		log.Fatalf("incr 交易确认失败：%v", err)
	}

	// 确认 incr 后的值
	countAfterIncr, err := instance.GetCount(&bind.CallOpts{})
	fmt.Printf("incr 后 count 值：%d\n", countAfterIncr)

	auth.Nonce.Add(auth.Nonce, big.NewInt(1))

	// 7. 调用写方法 add(uint256)（带参数的写方法）
	fmt.Println("\n=== 调用 add(5) 方法 ===")
	addNum := big.NewInt(5)                // 要添加的数值
	tx2, err := instance.Add(auth, addNum) // 传入 auth 和参数
	if err != nil {
		log.Fatalf("调用 add 失败：%v", err)
	}
	fmt.Printf("add 交易哈希：%s\n", tx2.Hash().Hex())
	fmt.Println("等待 add 交易确认...")
	_, err = bind.WaitMined(context.Background(), client, tx2)
	if err != nil {
		log.Fatalf("add 交易确认失败：%v", err)
	}

	// 确认 add 后的值
	countAfterAdd, err := instance.GetCount(&bind.CallOpts{})
	fmt.Printf("add(5) 后 count 值：%d\n", countAfterAdd)
}
