// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"sync"
	"time"
)

type Minter struct {
	sync.Mutex
	wallet  *Wallet
	started bool
	wg      sync.WaitGroup
	quit    chan struct{}
}

// Start begins the minting process.  Calling this function when the minter has
// already been started will have no effect.
//
// This function is safe for concurrent access.
func (m *Minter) Start() {
	m.Lock()
	defer m.Unlock()

	// Nothing to do if the miner is already running.
	if m.started {
		return
	}

	m.quit = make(chan struct{})
	m.wg.Add(1)

	go m.mintBlocks()

	m.started = true
	log.Infof("Minter started")
}

// Stop gracefully stops the mining process by signalling all workers, and the
// speed monitor to quit.  Calling this function when the CPU miner has not
// already been started will have no effect.
//
// This function is safe for concurrent access.
func (m *Minter) Stop() {
	m.Lock()
	defer m.Unlock()

	// Nothing to do if the miner is not currently running.
	if !m.started {
		return
	}

	close(m.quit)
	m.wg.Wait()
	m.started = false
	log.Infof("Minter stopped")
}

// WaitForShutdown blocks until all minter goroutines have finished executing.
func (m *Minter) WaitForShutdown() {
	m.wg.Wait()
}

// mintBlocks is a worker that is controlled by the miningWorkerController.
// It is self contained in that it creates block templates and attempts to solve
// them while detecting when it is performing stale work and reacting
// accordingly by generating a new block template.  When a block is solved, it
// is submitted.
//
// It must be run as a goroutine.
func (m *Minter) mintBlocks() {

	defer m.wg.Done()

	log.Tracef("Starting minting blocks worker")

out:
	for {
		// Quit when the miner is stopped.
		select {
		case <-m.quit:
			break out
		default:
			// Non-blocking select to fall through
		}

		// No point in searching for a solution before the chain is
		// synced.  Also, grab the same lock as used for block
		// submission, since the current block will be changing and
		// this would otherwise end up building a new block template on
		// a block that is in the process of becoming stale.
		if !m.wallet.ChainSynced() {
			time.Sleep(time.Second)
			continue
		}

		searchTime := time.Now().Unix()
		m.wallet.CreateCoinStake(searchTime)

		time.Sleep(time.Millisecond * 500)
	}

	log.Tracef("Minting blocks worker done")
}

// newMinter returns a new instance of a PPC minter for the provided wallet.
// Use Start to begin the minting process.  See the documentation for Minter
// type for more details.
func newMinter(w *Wallet) *Minter {
	return &Minter{
		wallet:            w,
	}
}