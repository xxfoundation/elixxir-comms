////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package dataStructures

import (
	"container/list"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elliotchance/orderedmap"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/comms/testutils"
	"gitlab.com/elixxir/primitives/current"
	"gitlab.com/elixxir/primitives/excludedRounds"
	"gitlab.com/elixxir/primitives/states"
	"gitlab.com/xx_network/primitives/netTime"
)

// Happy path of NewWaitingRounds().
func TestNewWaitingRounds(t *testing.T) {
	expectedWR := &WaitingRounds{
		writeRounds: orderedmap.NewOrderedMap(),
		readRounds:  &atomic.Value{},
		mux:         sync.Mutex{},
	}
	roundsList := make([]*Round, 0, 0)
	expectedWR.readRounds.Store(roundsList)
	testWR := NewWaitingRounds()

	expectedWR.signal = testWR.signal

	if !reflect.DeepEqual(expectedWR, testWR) {
		t.Errorf("NewWaitingRounds() did not return the expected object."+
			"\nexpected: %#v\nrecieved: %#v", expectedWR, testWR)
	}
}

// Happy path of WaitingRounds.Len().
func TestWaitingRounds_Len(t *testing.T) {
	expectedLen := list.New().Len()

	testWR := NewWaitingRounds()

	if expectedLen != testWR.Len() {
		t.Errorf("Len() did not return the expected length."+
			"\nexpected: %d\nrecieved: %d", expectedLen, testWR.Len())
	}
}

// Happy path of WaitingRounds.Insert().
func TestWaitingRounds_Insert(t *testing.T) {
	// Generate rounds
	expectedRounds, testRounds := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)
	t.Logf("expectedRndLen: %d", len(expectedRounds))
	t.Logf("testRndLen: %v", len(testRounds))
	// Add rounds to list
	testWR := NewWaitingRounds()
	testWR.Insert(expectedRounds, nil)

	if len(expectedRounds) != testWR.Len() {
		t.Fatalf("List does not have the expected length."+
			"\nexpected: %v\nrecieved: %v", len(expectedRounds), testWR.Len())
	}

	// Check if they are in the correct order
	for e, i := testWR.writeRounds.Front(), 0; e != nil; e, i = e.Next(), i+1 {
		if expectedRounds[i].info != e.Value.(*Round).info {
			t.Errorf("Insert() did not inser the correct element at position %d."+
				"\nexpected: %v\nrecieved: %v", i, expectedRounds[i], e.Value)
		}
	}
}

// Happy path of WaitingRounds.getFurthest().
func TestWaitingRounds_getFurthest(t *testing.T) {
	// Generate rounds
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)

	// Add rounds to list
	testWR := NewWaitingRounds()
	testWR.Insert(expectedRounds, nil)

	testWR.storeReadRounds()

	for i := len(expectedRounds) - 1; i >= 0; i-- {
		if !reflect.DeepEqual(expectedRounds[i], testWR.getFurthest(nil, 0)) {
			t.Errorf("getFurthest() did not return the expected round for %d."+
				"\nexpected: %+v\nrecieved: %+v", i,
				expectedRounds[i].info, testWR.getFurthest(nil, 0).info)
		}
		// testWR.remove(expectedRounds[i])
		testWR.Insert(nil, []*Round{expectedRounds[i]})
		testWR.storeReadRounds()
	}

	if testWR.getFurthest(nil, 0) != nil {
		t.Errorf("getFurthest() did not return nil on empty list.")
	}
}

// Happy path of WaitingRounds.getFurthest() with half the rounds excluded.
func TestWaitingRounds_getFurthest_Exclude(t *testing.T) {
	// Generate rounds
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)

	// Add rounds to list
	testWR := NewWaitingRounds()
	testWR.Insert(expectedRounds, nil)

	// Add rounds to exclusion list
	exclude := excludedRounds.NewSet()
	for i, round := range expectedRounds {
		if i%2 == 0 {
			exclude.Insert(round.info.GetRoundId())
		}
	}

	for i := len(expectedRounds) - 1; i >= 0; i-- {
		if i%2 == 1 {
			received := testWR.getFurthest(exclude, 0)
			if !reflect.DeepEqual(expectedRounds[i], received) {
				t.Errorf("getFurthest() did not return the expected round for %d."+
					"\nexpected: %v\nrecieved: %v", i, expectedRounds[i], received)
			}
			testWR.Insert(nil, []*Round{expectedRounds[i]})
			testWR.storeReadRounds()
		}
	}
	if testWR.getFurthest(exclude, 0) != nil {
		t.Errorf("getFurthest() did not return nil on empty list.")
	}
}

// Happy path.
func TestWaitingRounds_getClosest(t *testing.T) {
	// Generate rounds
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)

	// Add rounds to list
	testWR := NewWaitingRounds()
	testWR.Insert(expectedRounds, nil)

	for i := 0; i < len(expectedRounds); i++ {
		if !reflect.DeepEqual(expectedRounds[i], testWR.getClosest(nil, 0)) {
			t.Errorf("getClosest() did not return the expected round for %d."+
				"\nexpected: %+v\nrecieved: %+v", i,
				expectedRounds[i].info, testWR.getClosest(nil, 0).info)
		}
		testWR.Insert(nil, []*Round{expectedRounds[i]})
		testWR.storeReadRounds()
	}

	if testWR.getClosest(nil, 0) != nil {
		t.Errorf("getFurthest() did not return nil on empty list: %+v",
			testWR.writeRounds)
	}
}

// Happy path of WaitingRounds.GetUpcomingRealtime() when the list is not empty
// and no waiting occurs.
func TestWaitingRounds_GetUpcomingRealtime_NoWait(t *testing.T) {
	// Generate rounds and add to new list
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
	}

	testWR.Insert(expectedRounds, nil)

	for i := 0; i < len(expectedRounds); i++ {
		furthestRound, _, err := testWR.GetUpcomingRealtime(300*time.Millisecond, excludedRounds.NewSet(), 0, 0)
		if err != nil {
			t.Errorf("GetUpcomingRealtime() returned an unexpected error."+
				"\n\terror: %v", err)
		}
		if expectedRounds[i].info != furthestRound {
			t.Errorf("GetUpcomingRealtime() did not return the expected round (%d)."+
				"\nexpected: %+v\nrecieved: %+v", i, expectedRounds[i].info, furthestRound)
		}
		// testWR.remove(expectedRounds[i])
		testWR.Insert(nil, []*Round{expectedRounds[i]})
		testWR.storeReadRounds()
	}
}

// Tests that WaitingRounds.GetUpcomingRealtime() returns an error on an empty
// list after timeout.
func TestWaitingRounds_GetUpcomingRealtime_TimeoutError(t *testing.T) {
	testWR := NewWaitingRounds()
	testWR.storeReadRounds()

	furthestRound, _, err := testWR.GetUpcomingRealtime(300*time.Millisecond, excludedRounds.NewSet(), 0, 0)
	if err != timeOutError {
		t.Errorf("GetUpcomingRealtime() did not time out when expected."+
			"\nexpected: %v\n\treceived: %v", timeOutError, err)
	}
	if furthestRound != nil {
		t.Errorf("GetUpcomingRealtime() did not return nil on empty list."+
			"\nexpected: %v\nrecieved: %v", nil, furthestRound)
	}
}

// Happy path of WaitingRounds.GetUpcomingRealtime().
func TestWaitingRounds_GetUpcomingRealtime(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		go func(round *Round) {
			time.Sleep(30 * time.Millisecond)
			err := testutils.SignRoundInfoRsa(round.info, t)
			if err != nil {
				t.Errorf("Failed to sign round info #%d: %+v", i, err)
			}
			testWR.Insert([]*Round{round}, nil)
		}(round)
		testWR.storeReadRounds()

		furthestRound, _, err := testWR.GetUpcomingRealtime(5*time.Second, excludedRounds.NewSet(), 0, 0)

		if err != nil {
			t.Errorf("GetUpcomingRealtime() returned an unexpected error (%d)."+
				"\n\terror: %v", i, err)
		}
		if round.info != furthestRound {
			t.Errorf("GetUpcomingRealtime() did not return the expected round (%d)."+
				"\nexpected: %v\nrecieved: %v", i, round, furthestRound)
		}
		testWR.Insert(nil, []*Round{round})
	}
}

// Happy path of WaitingRounds.GetSlice().
func TestWaitingRounds_GetSlice(t *testing.T) {
	// Generate rounds and add to new list
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)
	testWR := NewWaitingRounds()

	ri := &pb.RoundInfo{
		ID:         rand.Uint64(),
		State:      uint32(states.QUEUED),
		Timestamps: []uint64{0, 0, 0, 0, 0},
	}

	ri.Timestamps[states.QUEUED] = uint64(netTime.Now().Add(100 * time.Millisecond).UnixNano())

	err := testutils.SignRoundInfoRsa(ri, t)
	if err != nil {
		t.Errorf("Failed to sign round info: %+v", err)
	}

	pubKey, err := testutils.LoadPublicKeyTesting(t)
	if err != nil {
		t.Errorf("Failed to load public key: %v", err)
		t.FailNow()
	}
	rnd := NewRound(ri, pubKey, nil)

	expectedRounds = append(expectedRounds, rnd)
	testWR.Insert(expectedRounds, nil)

	testSlice := testWR.GetSlice()

	// Convert Round slice to round info slice
outer:
	for i, val := range expectedRounds {
		for _, result := range testSlice {
			if val.info.ID == result.ID {
				continue outer
			}
		}
		t.Errorf("element %d not present", i)
	}

}

// Generates two lists of round infos. The first is the expected rounds in the
// correct order after inserting the second list of random round infos.
func createTestRoundInfos(num int, startTime time.Time, t *testing.T) ([]*Round, []*Round) {
	rounds := make([]*Round, num)
	var expectedRounds []*Round
	randomRounds := make([]*Round, num)
	pubKey, err := testutils.LoadPublicKeyTesting(t)
	if err != nil {
		t.Errorf("Failed to load public key: %v", err)
		t.FailNow()
	}

	for i := 0; i < num; i++ {
		include := false
		ri := &pb.RoundInfo{
			ID:         uint64(i),
			State:      uint32(rand.Int63n(int64(states.NUM_STATES) - 1)),
			Timestamps: make([]uint64, current.NUM_STATES),
		}
		ri.Timestamps[states.QUEUED] = uint64(startTime.UnixNano())
		if i%2 == 1 {
			ri.State = uint32(states.QUEUED)
			include = true
		} else if ri.State == uint32(states.QUEUED) {
			ri.State = uint32(states.REALTIME)
		}
		rounds[i] = NewRound(ri, pubKey, nil)
		if include {
			expectedRounds = append(expectedRounds, rounds[i])
		}

		startTime = startTime.Add(100 * time.Millisecond)

	}
	perm := rand.Perm(num)
	for i, v := range perm {
		randomRounds[v] = rounds[i]
	}
	return expectedRounds, randomRounds
}

// WaitingRounds.NumValidRounds() when all listed rounds are valid
func TestWaitingRounds_NumValidRounds_allValid(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
		testWR.Insert([]*Round{round}, nil)
	}

	numValid := testWR.NumValidRounds(time.Now())
	if numValid != len(expectedRounds) {
		t.Errorf("Could not get the correct number of valid rouns: %d vs %d",
			len(expectedRounds), numValid)
	}
}

// WaitingRounds.NumValidRounds() when no listed rounds are valid
func TestWaitingRounds_NumValidRounds_noValid(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(-5*time.Second), t)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
		testWR.Insert([]*Round{round}, nil)
	}

	numValid := testWR.NumValidRounds(time.Now())
	if numValid != 0 {
		t.Errorf("Could not get the correct number of valid rouns: %d vs %d",
			0, numValid)
	}
}

// WaitingRounds.NumValidRounds() when the rounds list is empty
func TestWaitingRounds_NumValidRounds_noRounds(t *testing.T) {
	// Generate rounds and WaitingRound
	testWR := NewWaitingRounds()

	numValid := testWR.NumValidRounds(time.Now())
	if numValid != 0 {
		t.Errorf("Could not get the correct number of valid rouns: %d vs %d",
			0, numValid)
	}
}

// WaitingRounds.NumValidRounds() when half the listed rounds are valid
func TestWaitingRounds_NumValidRounds_halfValid(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(-5*time.Second), t)
	expectedRounds2, _ := createTestRoundInfos(50, netTime.Now().Add(5*time.Second), t)
	expectedRounds = append(expectedRounds, expectedRounds2...)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
		testWR.Insert([]*Round{round}, nil)
	}

	numValid := testWR.NumValidRounds(time.Now())
	if numValid != len(expectedRounds2) {
		t.Errorf("Could not get the correct number of valid rouns: %d vs %d",
			len(expectedRounds2), numValid)
	}
}

// WaitingRounds.HasValidRounds() when all listed rounds are valid
func TestWaitingRounds_HasValidRounds_allValid(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(5*time.Second), t)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
		testWR.Insert([]*Round{round}, nil)
	}

	HasValid := testWR.HasValidRounds(time.Now())
	if !HasValid {
		t.Errorf("didnt not return a valid round when it should")
	}
}

// WaitingRounds.HasValidRounds() when no listed rounds are valid
func TestWaitingRounds_HasValidRounds_noValid(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(-5*time.Second), t)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
		testWR.Insert([]*Round{round}, nil)
	}

	HasValid := testWR.HasValidRounds(time.Now())
	if HasValid {
		t.Errorf("returned valid round when all invalid")
	}
}

// WaitingRounds.HasValidRounds() when the rounds list is empty
func TestWaitingRounds_HasValidRounds_noRounds(t *testing.T) {
	// Generate rounds and WaitingRound
	testWR := NewWaitingRounds()

	HasValid := testWR.HasValidRounds(time.Now())
	if HasValid {
		t.Errorf("returned valid round when all invalid")
	}
}

// WaitingRounds.HasValidRounds() when half the listed rounds are valid
func TestWaitingRounds_HasValidRounds_halfValid(t *testing.T) {
	// Generate rounds and WaitingRound
	expectedRounds, _ := createTestRoundInfos(25, netTime.Now().Add(-5*time.Second), t)
	expectedRounds2, _ := createTestRoundInfos(50, netTime.Now().Add(5*time.Second), t)
	expectedRounds = append(expectedRounds, expectedRounds2...)
	testWR := NewWaitingRounds()

	for i, round := range expectedRounds {
		err := testutils.SignRoundInfoRsa(round.info, t)
		if err != nil {
			t.Errorf("Failed to sign round info #%d: %+v", i, err)
		}
		testWR.Insert([]*Round{round}, nil)
	}

	hasValid := testWR.HasValidRounds(time.Now())
	if !hasValid {
		t.Errorf("returned that the rounds are invlaid whene there are valid rounds")
	}
}
