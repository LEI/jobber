package jobfile

import (
	"fmt"
	"sort"
	"time"
)

/*
This is an impl of RunLog that is not backed by a file.  In order to
avoid running out of memory, it has a max length, after which it starts
throwing out the oldest entries.
*/
type memOnlyRunLog struct {
	/*
	   We need to support method Get, which returns entries in
	   descending start-time order.  We also need to support method
	   Put, which will usually be called in ASCENDING start-time order.
	   Moreover, Put will be called more frequently than Get.

	   Will keep a list of entries in ascending start-time order,
	   which means that Put will usually run in constant-time.
	*/

	entries []*RunLogEntry
}

func (self *memOnlyRunLog) String() string {
	return fmt.Sprintf(
		"MemRunLog{maxLen: %v}",
		cap(self.entries),
	)
}

func NewMemOnlyRunLog(maxLen int) RunLog {
	if maxLen <= 0 {
		panic("maxLen must be > 0")
	}

	log := memOnlyRunLog{
		entries: make([]*RunLogEntry, 0, maxLen),
	}
	return &log
}

func (self *memOnlyRunLog) MaxLen() int {
	return cap(self.entries)
}

func (self *memOnlyRunLog) Put(newEntry RunLogEntry) error {
	/*
	   If the entries array would be too long after inserting the new
	   entry, we need to remove an entry first.  We remove the oldest
	   entry.

	   Remember: self.entries is sorted in ascending order.
	*/

	// assertions
	if cap(self.entries) == 0 {
		panic("Capacity is 0")
	}

	if len(self.entries)+1 > cap(self.entries) {
		// if the new entry is older than any other, do nothing
		if newEntry.Time.Before(self.entries[0].Time) {
			return nil
		} else {
			// remove oldest entry
			copy(self.entries, self.entries[1:])
			self.entries = self.entries[:len(self.entries)-1]
		}
	}

	// add the entry
	self.entries = append(self.entries, &newEntry)

	// make sure the array is sorted
	for i := len(self.entries) - 1; i >= 1; i-- {
		if newEntry.Time.Before(self.entries[i-1].Time) {
			// swap
			self.entries[i-1], self.entries[i] =
				self.entries[i], self.entries[i-1]
		} else {
			break
		}
	}

	return nil
}

func reverseEntryArray(array []*RunLogEntry) []*RunLogEntry {
	result := make([]*RunLogEntry, len(array))
	i := 0
	for j := len(array) - 1; j >= 0; j-- {
		result[i] = array[j]
		i++
	}
	return result
}

func (self *memOnlyRunLog) GetFromTime(maxTime time.Time,
	timeArr ...time.Time) ([]*RunLogEntry, error) {

	/*
	   Let [e_0, ..., e_n] be the (ascending-ordered) list of entries
	   (self.entries).

	   We must return a descending-ordered sublist of entries
	   [e_j, ..., e_i] (j <= i) s.t.
	    	   - e_j.Time < to
	    	   - e_(j+1).Time >= to
	    	   - e_i.Time >= from
	    	   - e_(i-1).Time < from
	*/

	if len(timeArr) > 1 {
		panic("Too many args.")
	}

	if len(self.entries) == 0 {
		return []*RunLogEntry{}, nil
	}

	var minTime time.Time
	if len(timeArr) >= 1 {
		minTime = timeArr[0]
	} else {
		// set *minTime* to just before the earliest entry's start time
		minTime = self.entries[0].Time.Add(-time.Second)
	}

	if maxTime.Before(minTime) {
		panic("maxTime is before minTime")
	}

	// do binary search to find beginning of range
	startIdx := sort.Search(len(self.entries), func(i int) bool {
		return !self.entries[i].Time.Before(maxTime)
	})
	if startIdx == len(self.entries) {
		return []*RunLogEntry{}, nil
	}

	// do binary search to find end of range
	endIdx := sort.Search(len(self.entries), func(i int) bool {
		return self.entries[i].Time.After(minTime)
	})

	// return in reverse order
	return reverseEntryArray(self.entries[endIdx : startIdx+1]), nil
}

func (self *memOnlyRunLog) GetFromIndex(minIdx int, idxArr ...int) (
	[]*RunLogEntry, error) {

	/*
	   Remember: self.entries is sorted in ascending order.  But we
	   must return in descending order.
	*/

	if len(idxArr) > 1 {
		panic("Too many args.")
	}

	var maxIdx int
	if len(idxArr) >= 1 {
		maxIdx = idxArr[0]
	} else {
		maxIdx = len(self.entries)
	}

	if minIdx > maxIdx {
		panic("minIdx > maxIdx")
	}
	if minIdx >= len(self.entries) {
		panic(fmt.Sprintf("Invalid 'minIdx' index: %v", minIdx))
	}
	if maxIdx > len(self.entries) {
		panic(fmt.Sprintf("Invalid 'maxIdx' index: %v", maxIdx))
	}

	/*
			self.entries is sorted in ascending order.  We must return in
			descending order.

		    self.entries: 0 1 2 3 4 5 6 7
		                  7 6 5 4 3 2 1 0

		    If from == 1 and to == 3 => (5, 7)
		    If from == 0 and to == 3 => (5, 8)
	*/

	// find entries
	actualTo := len(self.entries) - minIdx
	actualFrom := len(self.entries) - maxIdx

	// reverse them
	return reverseEntryArray(self.entries[actualFrom:actualTo]), nil
}

func (self *memOnlyRunLog) GetAll() ([]*RunLogEntry, error) {
	return reverseEntryArray(self.entries), nil
}

func (self *memOnlyRunLog) Len() int {
	return len(self.entries)
}
