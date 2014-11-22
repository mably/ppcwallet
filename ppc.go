// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	_ "github.com/conformal/btcec"
	"github.com/mably/btcnet"
	"github.com/mably/btcscript"
	"github.com/mably/btcutil"
	"github.com/mably/btcwire"
	"github.com/mably/ppcwallet/keystore"
)

// SignBlock TODO(?)
// ppc: sign block
func SignBlock(
	msgBlock *btcwire.MsgBlock, params *btcnet.Params, store *keystore.Store) error {

	var txOut *btcwire.TxOut

	if msgBlock.IsProofOfStake() {
		txOut = msgBlock.Transactions[1].TxOut[1]
	} else {
		txOut = msgBlock.Transactions[0].TxOut[0]
	}

	scriptClass, addresses, _, err :=
		btcscript.ExtractPkScriptAddrs(txOut.PkScript, params)
	if err != nil {
		return fmt.Errorf("cannot extract addresses: %v", err)
	}

	if scriptClass != btcscript.PubKeyTy {
		return UnsupportedTransactionType
	}

	apk, ok := addresses[0].(*btcutil.AddressPubKey)
	if !ok {
		return UnsupportedTransactionType
	}

	ai, err := store.Address(apk)
	if err != nil {
		return fmt.Errorf("cannot get address info: %v", err)
	}

	pka := ai.(keystore.PubKeyAddress)
	key, err := pka.PrivKey()
	if err != nil {
		return fmt.Errorf("cannot get private key: %v", err)
	}

	sha, err := msgBlock.BlockSha()
	if err != nil {
		return fmt.Errorf("cannot get block hash: %v", err)
	}

	signature, err := key.Sign(sha.Bytes())
	if err != nil {
		return fmt.Errorf("cannot sign block: %v", err)
	}

	msgBlock.Signature = signature.Serialize()

	return nil
}
