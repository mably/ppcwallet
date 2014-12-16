// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/kac-/umint"
	"github.com/mably/btcjson"
	"github.com/mably/btcutil"
	"github.com/mably/btcwire"
	"github.com/mably/btcws"
	"github.com/mably/ppcwallet/chain"
	"github.com/mably/ppcwallet/txstore"
)

func (w *Wallet) CreateCoinStake(fromTime int64) (err error) {

	// Get current block's height and hash.
	bs, err := w.chainSvr.BlockStamp()
	if err != nil {
		return
	}

	bits, err := w.chainSvr.CurrentTarget()
	if err != nil {
		return
	}

	params, err := w.chainSvr.Params()
	if err != nil {
		return
	}

	eligibles, err := w.findEligibleOutputs(6, bs)

	if err != nil || len(eligibles) == 0 {
		return
	}

	txNew := btcwire.NewMsgTx()

	nBalance, err := w.CalculateBalance(6)

	nCredit := btcutil.Amount(0)
	fKernelFound := false

	nStakeMinAge := params.StakeMinAge
	nMaxStakeSearchInterval := int64(60)

	for _, eligible := range eligibles {
		if w.ShuttingDown() {
			return
		}
		var block *txstore.Block
		block, err = eligible.Block()
		if err != nil {
			return
		}
		if block.Time.Unix()+nStakeMinAge > txNew.Time.Unix()-nMaxStakeSearchInterval {
			continue // only count coins meeting min age requirement
		}
		// Verify that block.KernelStakeModifier is defined
		if block.KernelStakeModifier == btcutil.KernelStakeModifierUnknown {
			var ksm uint64
			ksm, err = w.chainSvr.GetKernelStakeModifier(&block.Hash)
			if err != nil {
				log.Errorf("Error getting kernel stake modifier for block %v", &block.Hash)
				return
			} else {
				log.Infof("Found kernel stake modifier for block %v: %v", &block.Hash, ksm)
				block.KernelStakeModifier = ksm
				w.TxStore.MarkDirty()
			}
		}
		tx := eligible.Tx()
		for n := int64(0); n < 60 && !fKernelFound; n++ {
			if w.ShuttingDown() {
				return
			}
			stpl := umint.StakeKernelTemplate{
				//BlockFromTime:  int64(utx.BlockTime),
				BlockFromTime: block.Time.Unix(),
				//StakeModifier:  utx.StakeModifier,
				StakeModifier: block.KernelStakeModifier,
				//PrevTxOffset:   utx.OffsetInBlock,
				PrevTxOffset: tx.Offset(),
				//PrevTxTime:     int64(utx.Time),
				PrevTxTime: tx.MsgTx().Time.Unix(),
				//PrevTxOutIndex: outPoint.Index,
				PrevTxOutIndex: eligible.OutputIndex,
				//PrevTxOutValue: int64(utx.Value),
				PrevTxOutValue: int64(eligible.Amount()),
				IsProtocolV03:  true,
				StakeMinAge:    nStakeMinAge,
				Bits:           bits,
				TxTime:         fromTime - n,
			}
			var success bool
			_, success, err, _ = umint.CheckStakeKernelHash(&stpl)
			if err != nil {
				log.Errorf("Check kernel hash error: %v", err)
				return
			}
			if success {
				log.Infof("Valid kernel hash found!")
				// TODO create coinstake tx
				nCredit += eligible.Amount()
				fKernelFound = true
				break
			}
		}
		if fKernelFound {
			break
		}
	}

	//log.Infof("Credit available: %v / %v", nCredit, nBalance)

	if nCredit <= 0 || nCredit > nBalance {
		return
	}

	// TODO to be continued...

	return
}

type FoundStake struct {
	difficulty float32
	time       int64
}

func (w *Wallet) findStake(maxTime int64, diff float32) (foundStakes []FoundStake, err error) {

	// Get NetParams
	params, err := w.chainSvr.Params()
	if err != nil {
		return
	}

	// Get current block's height and hash.
	bs, err := w.chainSvr.BlockStamp()
	if err != nil {
		return
	}

	bits, err := w.chainSvr.CurrentTarget()
	if err != nil {
		return
	}

	if diff != 0 {
		bits = umint.BigToCompact(umint.DiffToTarget(diff))
	}

	log.Infof("Required difficulty: %v (%v)", umint.CompactToDiff(bits), bits)

	eligibles, err := w.findEligibleOutputs(6, bs)

	if err != nil || len(eligibles) == 0 {
		return
	}

	fromTime := time.Now().Unix()
	if maxTime == 0 {
		maxTime = fromTime + 30*24*60*60 // 30 days
	}

	foundStakes = make([]FoundStake, 0)

	nStakeMinAge := params.StakeMinAge
	nMaxStakeSearchInterval := int64(60)

	for _, eligible := range eligibles {
		if w.ShuttingDown() {
			return
		}
		var block *txstore.Block
		block, err = eligible.Block()
		if err != nil {
			return
		}
		if block.Time.Unix()+nStakeMinAge > fromTime-nMaxStakeSearchInterval {
			continue // only count coins meeting min age requirement
		}
		// Verify that block.KernelStakeModifier is defined
		if block.KernelStakeModifier == btcutil.KernelStakeModifierUnknown {
			var ksm uint64
			ksm, err = w.chainSvr.GetKernelStakeModifier(&block.Hash)
			if err != nil {
				log.Errorf("Error getting kernel stake modifier for block %v", &block.Hash)
				return
			} else {
				log.Infof("Found kernel stake modifier for block %v: %v", &block.Hash, ksm)
				block.KernelStakeModifier = ksm
				w.TxStore.MarkDirty()
			}
		}

		scriptClass, addresses, _, _ := eligible.Addresses(params)
		log.Infof("Addresses: %v (%v)", addresses, scriptClass)

		tx := eligible.Tx()

		log.Infof("CHECK %v PPCs from %v https://bkchain.org/ppc/tx/%v#o%v",
			float64(eligible.Amount())/1000000.0,
			time.Unix(int64(tx.MsgTx().Time.Unix()), 0).Format("2006-01-02"),
			eligible.OutPoint().Hash, eligible.OutPoint().Index)

		stpl := umint.StakeKernelTemplate{
			//BlockFromTime:  int64(utx.BlockTime),
			BlockFromTime: block.Time.Unix(),
			//StakeModifier:  utx.StakeModifier,
			StakeModifier: block.KernelStakeModifier,
			//PrevTxOffset:   utx.OffsetInBlock,
			PrevTxOffset: tx.Offset(),
			//PrevTxTime:     int64(utx.Time),
			PrevTxTime: tx.MsgTx().Time.Unix(),
			//PrevTxOutIndex: outPoint.Index,
			PrevTxOutIndex: eligible.OutputIndex,
			//PrevTxOutValue: int64(utx.Value),
			PrevTxOutValue: int64(eligible.Amount()),
			IsProtocolV03:  true,
			StakeMinAge:    nStakeMinAge,
			Bits:           bits,
			TxTime:         fromTime,
		}

		for true {
			if w.ShuttingDown() {
				return
			}
			_, succ, ferr, minTarget := umint.CheckStakeKernelHash(&stpl)
			if ferr != nil {
				err = fmt.Errorf("check kernel hash error :%v", ferr)
				return
			}
			if succ {
				comp := umint.IncCompact(umint.BigToCompact(minTarget))
				maximumDiff := umint.CompactToDiff(comp)
				log.Infof("MINT %v %v", time.Unix(stpl.TxTime, 0),
					maximumDiff)
				foundStakes = append(foundStakes, FoundStake{maximumDiff, stpl.TxTime})
			}
			stpl.TxTime++
			if stpl.TxTime > maxTime {
				break
			}
		}
	}

	return
}

// FindStake
func FindStake(w *Wallet, chainSvr *chain.Client, icmd btcjson.Cmd) (interface{}, error) {
	cmd := icmd.(*btcws.FindStakeCmd)

	foundStakes, err := w.findStake(cmd.MaxTime, cmd.Difficulty)
	if err != nil {
		return nil, err
	}

	stakesResult := []btcws.FindStakeResult{}
	for _, foundStake := range foundStakes {
		jsonResult := btcws.FindStakeResult{
			Difficulty: foundStake.difficulty,
			Time:       foundStake.time,
		}
		stakesResult = append(stakesResult, jsonResult)
	}

	return stakesResult, nil
}
