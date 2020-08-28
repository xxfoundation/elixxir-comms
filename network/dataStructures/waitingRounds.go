///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////
package dataStructures

import (
	"container/list"
	"github.com/pkg/errors"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/primitives/current"
	"gitlab.com/elixxir/primitives/states"
	"sync"
	"time"
)

var timeOutError = errors.New("Timed out getting round furthest in the future.")

// WaitingRounds contains a list of all queued rounds ordered by which occurs
// furthest in the future with the furthest in the the back.
type WaitingRounds struct {
	rounds *list.List
	c      *sync.Cond
	mux    sync.RWMutex
}

// NewWaitingRounds generates a new WaitingRounds with an empty round list.
func NewWaitingRounds() *WaitingRounds {
	wr := WaitingRounds{
		rounds: list.New(),
	}

	m := sync.Mutex{}
	wr.c = sync.NewCond(&m)

	return &wr
}

// Len returns the number of rounds in the list.
func (wr *WaitingRounds) Len() int {
	return wr.rounds.Len()
}

// Insert inserts a queued round into the list in order of its timestamp, from
// smallest to greatest. If the new round is not in a QUEUED state, then it is
// not inserted. If the new round already exists in the list but is no longer
// queued, then it is removed.
func (wr *WaitingRounds) Insert(newRound *pb.RoundInfo) {
	wr.mux.Lock()
	defer wr.mux.Unlock()

	// If the round is queued, then add it to the list; otherwise, remove it
	if newRound.GetState() == uint32(states.QUEUED) {

		// Loop through every round, starting with the furthest in the future
		for e := wr.rounds.Back(); e != nil; e = e.Prev() {
			// If the new round is larger, than add it before
			if getTime(newRound) > getTime(e.Value.(*pb.RoundInfo)) {
				wr.rounds.InsertAfter(newRound, e)

				// Broadcast change to GetUpcomingRealtime()
				wr.c.L.Lock()
				wr.c.Broadcast()
				wr.c.L.Unlock()

				return
			}
		}

		// If the round's realtime is the sooner than all other rounds, then add
		// it to the beginning  of the list
		wr.rounds.PushFront(newRound)

		// Broadcast change to GetUpcomingRealtime()
		wr.c.L.Lock()
		wr.c.Broadcast()
		wr.c.L.Unlock()

	} else {
		wr.remove(newRound)
	}
}

// getTime returns the timestamp for the round's realtime.
func getTime(round *pb.RoundInfo) uint64 {
	return round.Timestamps[current.REALTIME]
}

// remove deletes the round from the list if it exists.
func (wr *WaitingRounds) remove(newRound *pb.RoundInfo) {
	// Look for a node with a matching ID from the list
	for e := wr.rounds.Back(); e != nil; e = e.Prev() {
		if e.Value.(*pb.RoundInfo).ID == newRound.ID {
			wr.rounds.Remove(e)
			return
		}
	}
}

// getFurthest returns the round that will occur furthest in the future. If the
// list is empty, then nil is returned.
func (wr *WaitingRounds) getFurthest() *pb.RoundInfo {
	wr.mux.RLock()
	defer wr.mux.RUnlock()

	if wr.Len() == 0 {
		return nil
	}

	// Return the last round in the list, which is the furthest in the future
	return wr.rounds.Back().Value.(*pb.RoundInfo)
}

// GetSlice returns a slice of all round infos in the list
func (wr *WaitingRounds) GetSlice() []*pb.RoundInfo {
	wr.mux.RLock()
	defer wr.mux.RUnlock()

	roundInfos := make([]*pb.RoundInfo, wr.Len())

	for e, i := wr.rounds.Front(), 0; e != nil; e, i = e.Next(), i+1 {
		roundInfos[i] = e.Value.(*pb.RoundInfo)
	}

	// Return the last round in the list, which is the furthest in the future
	return roundInfos
}

// GetUpcomingRealtime returns the round that will occur furthest in the future.
// If the list is empty, then it waits waits for a round to be added for the
// specified duration. If no round is added, then an error is returned.
func (wr *WaitingRounds) GetUpcomingRealtime(timeout time.Duration) (*pb.RoundInfo, error) {

	// Start timeout timer
	timer := time.NewTimer(timeout)

	// Start waiting for rounds to be added
	sig := make(chan struct{}, 1)
	go func() {
		wr.c.L.Lock()
		wr.c.Wait()
		wr.c.L.Unlock()
		sig <- struct{}{}
	}()

	// If rounds already exist in the list, then return the the correct round
	// without waiting
	round := wr.getFurthest()
	if round != nil {
		return round, nil
	}

	// If the list is empty, then start waiting for rounds to be added.
	for {
		select {
		case <-timer.C:
			return nil, timeOutError
		case <-sig:
			round := wr.getFurthest()
			if round != nil {
				return round, nil
			}
		}
	}
}