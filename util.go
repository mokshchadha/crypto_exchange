package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferETH(client *ethclient.Client, to common.Address, fromPvtKey *ecdsa.PrivateKey, amount *big.Int) error {
	ctx := context.Background()

	publicKey := fromPvtKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)

	if err != nil {
		log.Fatal(err)
	}

	value := amount           // in wei (1 eth)
	gasLimit := uint64(21000) //

	gasPrice, err := client.SuggestGasPrice(ctx)

	toAddress := to

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	fmt.Println(tx)

	// chainID, err := client.NetworkID(ctx)
	chainID := big.NewInt(1337)
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPvtKey)

	if err != nil {
		return err

	}
	return client.SendTransaction(ctx, signedTx)

}
