//go:build goexperiment.swissmap

package zgo

import "unsafe"

//go:linkname rand runtime.rand
func rand() uint64

type MapType struct {
	Type
	Key   *Type
	Elem  *Type
	Group *Type // internal type representing a slot group
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher    func(unsafe.Pointer, uintptr) uintptr
	GroupSize uintptr // == Group.Size_
	SlotSize  uintptr // size of key/elem slot
	ElemOff   uintptr // offset of elem in key/elem slot
	Flags     uint32
}

// Flag values
const (
	SwissMapNeedKeyUpdate = 1 << iota
	SwissMapHashMightPanic
	SwissMapIndirectKey
	SwissMapIndirectElem
)

func (mt *MapType) NeedKeyUpdate() bool { // true if we need to update key on an overwrite
	return mt.Flags&SwissMapNeedKeyUpdate != 0
}
func (mt *MapType) HashMightPanic() bool { // true if hash function might panic
	return mt.Flags&SwissMapHashMightPanic != 0
}
func (mt *MapType) IndirectKey() bool { // store ptr to key instead of key itself
	return mt.Flags&SwissMapIndirectKey != 0
}
func (mt *MapType) IndirectElem() bool { // store ptr to elem instead of elem itself
	return mt.Flags&SwissMapIndirectElem != 0
}

type Map struct {
	// The number of filled slots (i.e. the number of elements in all
	// tables). Excludes deleted slots.
	// Must be first (known by the compiler, for len() builtin).
	used uint64

	// seed is the hash seed, computed as a unique random number per map.
	seed uintptr

	// The directory of tables.
	//
	// Normally dirPtr points to an array of table pointers
	//
	// dirPtr *[dirLen]*table
	//
	// The length (dirLen) of this array is `1 << globalDepth`. Multiple
	// entries may point to the same table. See top-level comment for more
	// details.
	//
	// Small map optimization: if the map always contained
	// SwissMapGroupSlots or fewer entries, it fits entirely in a
	// single group. In that case dirPtr points directly to a single group.
	//
	// dirPtr *group
	//
	// In this case, dirLen is 0. used counts the number of used slots in
	// the group. Note that small maps never have deleted slots (as there
	// is no probe sequence to maintain).
	dirPtr unsafe.Pointer
	dirLen int

	// The number of bits to use in table directory lookups.
	globalDepth uint8

	// The number of bits to shift out of the hash for directory lookups.
	// On 64-bit systems, this is 64 - globalDepth.
	globalShift uint8

	// writing is a flag that is toggled (XOR 1) while the map is being
	// written. Normally it is set to 1 when writing, but if there are
	// multiple concurrent writers, then toggling increases the probability
	// that both sides will detect the race.
	writing uint8

	// clearSeq is a sequence counter of calls to Clear. It is used to
	// detect map clears during iteration.
	clearSeq uint64
}

func (m *Map) Used() uint64 {
	return m.used
}

// Extracts the H1 portion of a hash: the 57 upper bits.
// TODO(prattmic): what about 32-bit systems?
func h1(h uintptr) uintptr {
	return h >> 7
}

// Extracts the H2 portion of a hash: the 7 bits not used for h1.
//
// These are used as an occupied control byte.
func h2(h uintptr) uintptr {
	return h & 0x7f
}

func (m *Map) getWithKeySmall(typ *MapType, hash uintptr, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer, bool) {
	g := groupReference{
		data: m.dirPtr,
	}

	h2 := uint8(h2(hash))
	ctrls := *g.ctrls()

	for i := uintptr(0); i < SwissMapGroupSlots; i++ {
		c := uint8(ctrls)
		ctrls >>= 8
		if c != h2 {
			continue
		}

		slotKey := g.key(typ, i)
		if typ.IndirectKey() {
			slotKey = *((*unsafe.Pointer)(slotKey))
		}

		if typ.Key.Equal(key, slotKey) {
			slotElem := g.elem(typ, i)
			if typ.IndirectElem() {
				slotElem = *((*unsafe.Pointer)(slotElem))
			}
			return slotKey, slotElem, true
		}
	}

	return nil, nil, false
}

func (m *Map) directoryIndex(hash uintptr) uintptr {
	if m.dirLen == 1 {
		return 0
	}
	return hash >> (m.globalShift & 63)
}

const PtrSize = unsafe.Sizeof(uintptr(0))

func (m *Map) directoryAt(i uintptr) *table {
	return *(**table)(unsafe.Pointer(uintptr(m.dirPtr) + PtrSize*i))
}

func (m *Map) getWithKey(typ *MapType, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer, bool) {
	if m.Used() == 0 {
		return nil, nil, false
	}

	if m.writing != 0 {
		panic("concurrent map read and map write")
	}

	hash := typ.Hasher(key, m.seed)

	if m.dirLen == 0 {
		return m.getWithKeySmall(typ, hash, key)
	}

	idx := m.directoryIndex(hash)
	return m.directoryAt(idx).getWithKey(typ, hash, key)
}

type MapIterator struct {
	Key  unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/compile/internal/walk/range.go).
	Elem unsafe.Pointer // Must be in second position (see cmd/compile/internal/walk/range.go).
	typ  *MapType
	m    *Map

	// Randomize iteration order by starting iteration at a random slot
	// offset. The offset into the directory uses a separate offset, as it
	// must adjust when the directory grows.
	entryOffset uint64
	dirOffset   uint64

	// Snapshot of Map.clearSeq at iteration initialization time. Used to
	// detect clear during iteration.
	clearSeq uint64

	// Value of Map.globalDepth during the last call to Next. Used to
	// detect directory grow during iteration.
	globalDepth uint8

	// dirIdx is the current directory index, prior to adjustment by
	// dirOffset.
	dirIdx int

	// tab is the table at dirIdx during the previous call to Next.
	tab *table

	// group is the group at entryIdx during the previous call to Next.
	group groupReference

	// entryIdx is the current entry index, prior to adjustment by entryOffset.
	// The lower 3 bits of the index are the slot index, and the upper bits
	// are the group index.
	entryIdx uint64
}

func (it *MapIterator) nextDirIdx() {
	// Skip other entries in the directory that refer to the same
	// logical table. There are two cases of this:
	//
	// Consider this directory:
	//
	// - 0: *t1
	// - 1: *t1
	// - 2: *t2a
	// - 3: *t2b
	//
	// At some point, the directory grew to accommodate a split of
	// t2. t1 did not split, so entries 0 and 1 both point to t1.
	// t2 did split, so the two halves were installed in entries 2
	// and 3.
	//
	// If dirIdx is 0 and it.tab is t1, then we should skip past
	// entry 1 to avoid repeating t1.
	//
	// If dirIdx is 2 and it.tab is t2 (pre-split), then we should
	// skip past entry 3 because our pre-split t2 already covers
	// all keys from t2a and t2b (except for new insertions, which
	// iteration need not return).
	//
	// We can achieve both of these by using to difference between
	// the directory and table depth to compute how many entries
	// the table covers.
	entries := 1 << (it.m.globalDepth - it.tab.localDepth)
	it.dirIdx += entries
	it.tab = nil
	it.group = groupReference{}
	it.entryIdx = 0
}

// Init initializes Iter for iteration.
func (it *MapIterator) Init(typ *MapType, m *Map) {
	it.typ = typ

	if m == nil || m.used == 0 {
		return
	}

	dirIdx := 0
	var groupSmall groupReference
	if m.dirLen <= 0 {
		// Use dirIdx == -1 as sentinel for small maps.
		dirIdx = -1
		groupSmall.data = m.dirPtr
	}

	it.m = m
	it.entryOffset = rand()
	it.dirOffset = rand()
	it.globalDepth = m.globalDepth
	it.dirIdx = dirIdx
	it.group = groupSmall
	it.clearSeq = m.clearSeq

	it.Next()
}

type table struct {
	// The number of filled slots (i.e. the number of elements in the table).
	Used uint16

	// The total number of slots (always 2^N). Equal to
	// `(groups.lengthMask+1)*SwissMapGroupSlots`.
	capacity uint16

	// The number of slots we can still fill without needing to rehash.
	//
	// We rehash when used + tombstones > loadFactor*capacity, including
	// tombstones so the table doesn't overfill with tombstones. This field
	// counts down remaining empty slots before the next rehash.
	GrowthLeft uint16

	// The number of bits used by directory lookups above this table. Note
	// that this may be less then globalDepth, if the directory has grown
	// but this table has not yet been split.
	localDepth uint8

	// Index of this table in the Map directory. This is the index of the
	// _first_ location in the directory. The table may occur in multiple
	// sequential indicies.
	//
	// index is -1 if the table is stale (no longer installed in the
	// directory).
	index int

	// groups is an array of slot groups. Each group holds SwissMapGroupSlots
	// key/elem slots and their control bytes. A table has a fixed size
	// groups array. The table is replaced (in rehash) when more space is
	// required.
	//
	// TODO(prattmic): keys and elements are interleaved to maximize
	// locality, but it comes at the expense of wasted space for some types
	// (consider uint8 key, uint64 element). Consider placing all keys
	// together in these cases to save space.
	groups groupsReference
}

type groupsReference struct {
	// data points to an array of groups. See groupReference above for the
	// definition of group.
	data unsafe.Pointer // data *[length]typ.Group

	// lengthMask is the number of groups in data minus one (note that
	// length must be a power of two). This allows computing i%length
	// quickly using bitwise AND.
	lengthMask uint64
}

// group returns the group at index i.
func (g *groupsReference) group(typ *MapType, i uint64) groupReference {
	// TODO(prattmic): Do something here about truncation on cast to
	// uintptr on 32-bit systems?
	offset := uintptr(i) * typ.GroupSize

	return groupReference{
		data: unsafe.Pointer(uintptr(g.data) + offset),
	}
}

type groupReference struct {
	// data points to the group, which is described by typ.Group and has
	// layout:
	//
	// type group struct {
	// 	ctrls ctrlGroup
	// 	slots [SwissMapGroupSlots]slot
	// }
	//
	// type slot struct {
	// 	key  typ.Key
	// 	elem typ.Elem
	// }
	data unsafe.Pointer // data *typ.Group
}

func checkBigEndian() bool {
	var i uint16 = 0x1
	b := (*[2]byte)(unsafe.Pointer(&i))
	return b[0] == 0
}

var isBigEndian = checkBigEndian()

const (
	ctrlGroupsSize   = unsafe.Sizeof(ctrlGroup(0))
	groupSlotsOffset = ctrlGroupsSize
)

type ctrl uint8

// ctrlGroup is a fixed size array of abi.SwissMapGroupSlots control bytes
// stored in a uint64.
type ctrlGroup uint64

// matchEmpty returns the set of slots in the group that are empty.
func (g ctrlGroup) matchEmpty() bitset {
	return ctrlGroupMatchEmpty(g)
}

// get returns the i-th control byte.
func (g *ctrlGroup) get(i uintptr) ctrl {
	if isBigEndian {
		return *(*ctrl)(unsafe.Add(unsafe.Pointer(g), 7-i))
	}
	return *(*ctrl)(unsafe.Add(unsafe.Pointer(g), i))
}

// matchH2 returns the set of slots which are full and for which the 7-bit hash
// matches the given value. May return false positives.
func (g ctrlGroup) matchH2(h uintptr) bitset {
	return ctrlGroupMatchH2(g, h)
}

// Portable implementation of matchH2.
//
// Note: On AMD64, this is an intrinsic implemented with SIMD instructions. See
// note on bitset about the packed instrinsified return value.
func ctrlGroupMatchH2(g ctrlGroup, h uintptr) bitset {
	// NB: This generic matching routine produces false positive matches when
	// h is 2^N and the control bytes have a seq of 2^N followed by 2^N+1. For
	// example: if ctrls==0x0302 and h=02, we'll compute v as 0x0100. When we
	// subtract off 0x0101 the first 2 bytes we'll become 0xffff and both be
	// considered matches of h. The false positive matches are not a problem,
	// just a rare inefficiency. Note that they only occur if there is a real
	// match and never occur on ctrlEmpty, or ctrlDeleted. The subsequent key
	// comparisons ensure that there is no correctness issue.
	v := uint64(g) ^ (bitsetLSB * uint64(h))
	return bitset(((v - bitsetLSB) &^ v) & bitsetMSB)
}

// Portable implementation of matchEmpty.
//
// Note: On AMD64, this is an intrinsic implemented with SIMD instructions. See
// note on bitset about the packed instrinsified return value.
func ctrlGroupMatchEmpty(g ctrlGroup) bitset {
	// An empty slot is   1000 0000
	// A deleted slot is  1111 1110
	// A full slot is     0??? ????
	//
	// A slot is empty iff bit 7 is set and bit 1 is not. We could select any
	// of the other bits here (e.g. v << 1 would also work).
	v := uint64(g)
	return bitset((v &^ (v << 6)) & bitsetMSB)
}

// matchFull returns the set of slots in the group that are full.
func (g ctrlGroup) matchFull() bitset {
	return ctrlGroupMatchFull(g)
}

// Portable implementation of matchFull.
//
// Note: On AMD64, this is an intrinsic implemented with SIMD instructions. See
// note on bitset about the packed instrinsified return value.
func ctrlGroupMatchFull(g ctrlGroup) bitset {
	// An empty slot is  1000 0000
	// A deleted slot is 1111 1110
	// A full slot is    0??? ????
	//
	// A slot is full iff bit 7 is unset.
	v := uint64(g)
	return bitset(^v & bitsetMSB)
}

// ctrls returns the group control word.
func (g *groupReference) ctrls() *ctrlGroup {
	return (*ctrlGroup)(g.data)
}

// key returns a pointer to the key at index i.
func (g *groupReference) key(typ *MapType, i uintptr) unsafe.Pointer {
	offset := groupSlotsOffset + i*typ.SlotSize

	return unsafe.Pointer(uintptr(g.data) + offset)
}

// elem returns a pointer to the element at index i.
func (g *groupReference) elem(typ *MapType, i uintptr) unsafe.Pointer {
	offset := groupSlotsOffset + i*typ.SlotSize + typ.ElemOff

	return unsafe.Pointer(uintptr(g.data) + offset)
}

const (
	// Number of bits in the group.slot count.
	SwissMapGroupSlotsBits = 3

	// Number of slots in a group.
	SwissMapGroupSlots = 1 << SwissMapGroupSlotsBits // 8

	// Maximum key or elem size to keep inline (instead of mallocing per element).
	// Must fit in a uint8.
	SwissMapMaxKeyBytes  = 128
	SwissMapMaxElemBytes = 128

	ctrlEmpty = 0b10000000
	bitsetLSB = 0x0101010101010101

	// Value of control word with all empty slots.
	SwissMapCtrlEmpty = bitsetLSB * uint64(ctrlEmpty)
)

func (it *MapIterator) Next() {
	if it.m == nil {
		// Map was empty at Iter.Init.
		it.Key = nil
		it.Elem = nil
		return
	}

	if it.m.writing != 0 {
		panic("concurrent map iteration and map write")
	}

	if it.dirIdx < 0 {
		// Map was small at Init.
		for ; it.entryIdx < SwissMapGroupSlots; it.entryIdx++ {
			k := uintptr(it.entryIdx+it.entryOffset) % SwissMapGroupSlots

			if (it.group.ctrls().get(k) & ctrlEmpty) == ctrlEmpty {
				// Empty or deleted.
				continue
			}

			key := it.group.key(it.typ, k)
			if it.typ.IndirectKey() {
				key = *((*unsafe.Pointer)(key))
			}

			// As below, if we have grown to a full map since Init,
			// we continue to use the old group to decide the keys
			// to return, but must look them up again in the new
			// tables.
			grown := it.m.dirLen > 0
			var elem unsafe.Pointer
			if grown {
				var ok bool
				newKey, newElem, ok := it.m.getWithKey(it.typ, key)
				if !ok {
					// See comment below.
					if it.clearSeq == it.m.clearSeq && !it.typ.Key.Equal(key, key) {
						elem = it.group.elem(it.typ, k)
						if it.typ.IndirectElem() {
							elem = *((*unsafe.Pointer)(elem))
						}
					} else {
						continue
					}
				} else {
					key = newKey
					elem = newElem
				}
			} else {
				elem = it.group.elem(it.typ, k)
				if it.typ.IndirectElem() {
					elem = *((*unsafe.Pointer)(elem))
				}
			}

			it.entryIdx++
			it.Key = key
			it.Elem = elem
			return
		}
		it.Key = nil
		it.Elem = nil
		return
	}

	if it.globalDepth != it.m.globalDepth {
		// Directory has grown since the last call to Next. Adjust our
		// directory index.
		//
		// Consider:
		//
		// Before:
		// - 0: *t1
		// - 1: *t2  <- dirIdx
		//
		// After:
		// - 0: *t1a (split)
		// - 1: *t1b (split)
		// - 2: *t2  <- dirIdx
		// - 3: *t2
		//
		// That is, we want to double the current index when the
		// directory size doubles (or quadruple when the directory size
		// quadruples, etc).
		//
		// The actual (randomized) dirIdx is computed below as:
		//
		// dirIdx := (it.dirIdx + it.dirOffset) % it.m.dirLen
		//
		// Multiplication is associative across modulo operations,
		// A * (B % C) = (A * B) % (A * C),
		// provided that A is positive.
		//
		// Thus we can achieve this by adjusting it.dirIdx,
		// it.dirOffset, and it.m.dirLen individually.
		orders := it.m.globalDepth - it.globalDepth
		it.dirIdx <<= orders
		it.dirOffset <<= orders
		// it.m.dirLen was already adjusted when the directory grew.

		it.globalDepth = it.m.globalDepth
	}

	// Continue iteration until we find a full slot.
	for ; it.dirIdx < it.m.dirLen; it.nextDirIdx() {
		// Resolve the table.
		if it.tab == nil {
			dirIdx := int((uint64(it.dirIdx) + it.dirOffset) & uint64(it.m.dirLen-1))
			newTab := it.m.directoryAt(uintptr(dirIdx))
			if newTab.index != dirIdx {
				// Normally we skip past all duplicates of the
				// same entry in the table (see updates to
				// it.dirIdx at the end of the loop below), so
				// this case wouldn't occur.
				//
				// But on the very first call, we have a
				// completely randomized dirIdx that may refer
				// to a middle of a run of tables in the
				// directory. Do a one-time adjustment of the
				// offset to ensure we start at first index for
				// newTable.
				diff := dirIdx - newTab.index
				it.dirOffset -= uint64(diff)
				// dirIdx = newTab.index
			}
			it.tab = newTab
		}

		// N.B. Use it.tab, not newTab. It is important to use the old
		// table for key selection if the table has grown. See comment
		// on grown below.

		entryMask := uint64(it.tab.capacity) - 1
		if it.entryIdx > entryMask {
			// Continue to next table.
			continue
		}

		// Fast path: skip matching and directly check if entryIdx is a
		// full slot.
		//
		// In the slow path below, we perform an 8-slot match check to
		// look for full slots within the group.
		//
		// However, with a max load factor of 7/8, each slot in a
		// mostly full map has a high probability of being full. Thus
		// it is cheaper to check a single slot than do a full control
		// match.

		entryIdx := (it.entryIdx + it.entryOffset) & entryMask
		slotIdx := uintptr(entryIdx & (SwissMapGroupSlots - 1))
		if slotIdx == 0 || it.group.data == nil {
			// Only compute the group (a) when we switch
			// groups (slotIdx rolls over) and (b) on the
			// first iteration in this table (slotIdx may
			// not be zero due to entryOffset).
			groupIdx := entryIdx >> SwissMapGroupSlotsBits
			it.group = it.tab.groups.group(it.typ, groupIdx)
		}

		if (it.group.ctrls().get(slotIdx) & ctrlEmpty) == 0 {
			// Slot full.

			key := it.group.key(it.typ, slotIdx)
			if it.typ.IndirectKey() {
				key = *((*unsafe.Pointer)(key))
			}

			grown := it.tab.index == -1
			var elem unsafe.Pointer
			if grown {
				newKey, newElem, ok := it.grownKeyElem(key, slotIdx)
				if !ok {
					// This entry doesn't exist
					// anymore. Continue to the
					// next one.
					goto next
				} else {
					key = newKey
					elem = newElem
				}
			} else {
				elem = it.group.elem(it.typ, slotIdx)
				if it.typ.IndirectElem() {
					elem = *((*unsafe.Pointer)(elem))
				}
			}

			it.entryIdx++
			it.Key = key
			it.Elem = elem
			return
		}

	next:
		it.entryIdx++

		// Slow path: use a match on the control word to jump ahead to
		// the next full slot.
		//
		// This is highly effective for maps with particularly low load
		// (e.g., map allocated with large hint but few insertions).
		//
		// For maps with medium load (e.g., 3-4 empty slots per group)
		// it also tends to work pretty well. Since slots within a
		// group are filled in order, then if there have been no
		// deletions, a match will allow skipping past all empty slots
		// at once.
		//
		// Note: it is tempting to cache the group match result in the
		// iterator to use across Next calls. However because entries
		// may be deleted between calls later calls would still need to
		// double-check the control value.

		var groupMatch bitset
		for it.entryIdx <= entryMask {
			entryIdx := (it.entryIdx + it.entryOffset) & entryMask
			slotIdx := uintptr(entryIdx & (SwissMapGroupSlots - 1))

			if slotIdx == 0 || it.group.data == nil {
				// Only compute the group (a) when we switch
				// groups (slotIdx rolls over) and (b) on the
				// first iteration in this table (slotIdx may
				// not be zero due to entryOffset).
				groupIdx := entryIdx >> SwissMapGroupSlotsBits
				it.group = it.tab.groups.group(it.typ, groupIdx)
			}

			if groupMatch == 0 {
				groupMatch = it.group.ctrls().matchFull()

				if slotIdx != 0 {
					// Starting in the middle of the group.
					// Ignore earlier groups.
					groupMatch = groupMatch.removeBelow(slotIdx)
				}

				// Skip over groups that are composed of only empty or
				// deleted slots.
				if groupMatch == 0 {
					// Jump past remaining slots in this
					// group.
					it.entryIdx += SwissMapGroupSlots - uint64(slotIdx)
					continue
				}

				i := groupMatch.first()
				it.entryIdx += uint64(i - slotIdx)
				if it.entryIdx > entryMask {
					// Past the end of this table's iteration.
					continue
				}
				entryIdx += uint64(i - slotIdx)
				slotIdx = i
			}

			key := it.group.key(it.typ, slotIdx)
			if it.typ.IndirectKey() {
				key = *((*unsafe.Pointer)(key))
			}

			// If the table has changed since the last
			// call, then it has grown or split. In this
			// case, further mutations (changes to
			// key->elem or deletions) will not be visible
			// in our snapshot table. Instead we must
			// consult the new table by doing a full
			// lookup.
			//
			// We still use our old table to decide which
			// keys to lookup in order to avoid returning
			// the same key twice.
			grown := it.tab.index == -1
			var elem unsafe.Pointer
			if grown {
				newKey, newElem, ok := it.grownKeyElem(key, slotIdx)
				if !ok {
					// This entry doesn't exist anymore.
					// Continue to the next one.
					groupMatch = groupMatch.removeFirst()
					if groupMatch == 0 {
						// No more entries in this
						// group. Continue to next
						// group.
						it.entryIdx += SwissMapGroupSlots - uint64(slotIdx)
						continue
					}

					// Next full slot.
					i := groupMatch.first()
					it.entryIdx += uint64(i - slotIdx)
					continue
				} else {
					key = newKey
					elem = newElem
				}
			} else {
				elem = it.group.elem(it.typ, slotIdx)
				if it.typ.IndirectElem() {
					elem = *((*unsafe.Pointer)(elem))
				}
			}

			// Jump ahead to the next full slot or next group.
			groupMatch = groupMatch.removeFirst()
			if groupMatch == 0 {
				// No more entries in
				// this group. Continue
				// to next group.
				it.entryIdx += SwissMapGroupSlots - uint64(slotIdx)
			} else {
				// Next full slot.
				i := groupMatch.first()
				it.entryIdx += uint64(i - slotIdx)
			}

			it.Key = key
			it.Elem = elem
			return
		}

		// Continue to next table.
	}

	it.Key = nil
	it.Elem = nil
}

// Return the appropriate key/elem for key at slotIdx index within it.group, if
// any.
func (it *MapIterator) grownKeyElem(key unsafe.Pointer, slotIdx uintptr) (unsafe.Pointer, unsafe.Pointer, bool) {
	newKey, newElem, ok := it.m.getWithKey(it.typ, key)
	if !ok {
		// Key has likely been deleted, and
		// should be skipped.
		//
		// One exception is keys that don't
		// compare equal to themselves (e.g.,
		// NaN). These keys cannot be looked
		// up, so getWithKey will fail even if
		// the key exists.
		//
		// However, we are in luck because such
		// keys cannot be updated and they
		// cannot be deleted except with clear.
		// Thus if no clear has occurred, the
		// key/elem must still exist exactly as
		// in the old groups, so we can return
		// them from there.
		//
		// TODO(prattmic): Consider checking
		// clearSeq early. If a clear occurred,
		// Next could always return
		// immediately, as iteration doesn't
		// need to return anything added after
		// clear.
		if it.clearSeq == it.m.clearSeq && !it.typ.Key.Equal(key, key) {
			elem := it.group.elem(it.typ, slotIdx)
			if it.typ.IndirectElem() {
				elem = *((*unsafe.Pointer)(elem))
			}
			return key, elem, true
		}

		// This entry doesn't exist anymore.
		return nil, nil, false
	}

	return newKey, newElem, true
}

func (t *table) getWithKey(typ *MapType, hash uintptr, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer, bool) {
	// To find the location of a key in the table, we compute hash(key). From
	// h1(hash(key)) and the capacity, we construct a probeSeq that visits
	// every group of slots in some interesting order. See [probeSeq].
	//
	// We walk through these indices. At each index, we select the entire
	// group starting with that index and extract potential candidates:
	// occupied slots with a control byte equal to h2(hash(key)). The key
	// at candidate slot i is compared with key; if key == g.slot(i).key
	// we are done and return the slot; if there is an empty slot in the
	// group, we stop and return an error; otherwise we continue to the
	// next probe index. Tombstones (ctrlDeleted) effectively behave like
	// full slots that never match the value we're looking for.
	//
	// The h2 bits ensure when we compare a key we are likely to have
	// actually found the object. That is, the chance is low that keys
	// compare false. Thus, when we search for an object, we are unlikely
	// to call Equal many times. This likelihood can be analyzed as follows
	// (assuming that h2 is a random enough hash function).
	//
	// Let's assume that there are k "wrong" objects that must be examined
	// in a probe sequence. For example, when doing a find on an object
	// that is in the table, k is the number of objects between the start
	// of the probe sequence and the final found object (not including the
	// final found object). The expected number of objects with an h2 match
	// is then k/128. Measurements and analysis indicate that even at high
	// load factors, k is less than 32, meaning that the number of false
	// positive comparisons we must perform is less than 1/8 per find.
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)
	for ; ; seq = seq.next() {
		g := t.groups.group(typ, seq.offset)

		match := g.ctrls().matchH2(h2(hash))

		for match != 0 {
			i := match.first()

			slotKey := g.key(typ, i)
			if typ.IndirectKey() {
				slotKey = *((*unsafe.Pointer)(slotKey))
			}
			if typ.Key.Equal(key, slotKey) {
				slotElem := g.elem(typ, i)
				if typ.IndirectElem() {
					slotElem = *((*unsafe.Pointer)(slotElem))
				}
				return slotKey, slotElem, true
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return nil, nil, false
		}
	}
}

type probeSeq struct {
	mask   uint64
	offset uint64
	index  uint64
}

func makeProbeSeq(hash uintptr, mask uint64) probeSeq {
	return probeSeq{
		mask:   mask,
		offset: uint64(hash) & mask,
		index:  0,
	}
}

func (s probeSeq) next() probeSeq {
	s.index++
	s.offset = (s.offset + s.index) & s.mask
	return s
}

type bitset uint64

const bitsetMSB = 0x8080808080808080

// first returns the relative index of the first control byte in the group that
// is in the set.
//
// Preconditions: b is not 0 (empty).
func (b bitset) first() uintptr {
	return bitsetFirst(b)
}

// Portable implementation of first.
//
// On AMD64, this is replaced with an intrisic that simply does
// TrailingZeros64. There is no need to shift as the bitset is packed.
func bitsetFirst(b bitset) uintptr {
	return uintptr(TrailingZeros64(uint64(b))) >> 3
}

func TrailingZeros64(x uint64) int {
	if x == 0 {
		return 64
	}
	// If popcount is fast, replace code below with return popcount(^x & (x - 1)).
	//
	// x & -x leaves only the right-most bit set in the word. Let k be the
	// index of that bit. Since only a single bit is set, the value is two
	// to the power of k. Multiplying by a power of two is equivalent to
	// left shifting, in this case by k bits. The de Bruijn (64 bit) constant
	// is such that all six bit, consecutive substrings are distinct.
	// Therefore, if we have a left shifted version of this constant we can
	// find by how many bits it was shifted by looking at which six bit
	// substring ended up at the top of the word.
	// (Knuth, volume 4, section 7.3.1)
	return int(deBruijn64tab[(x&-x)*deBruijn64>>(64-6)])
}

var deBruijn64tab = [64]byte{
	0, 1, 56, 2, 57, 49, 28, 3, 61, 58, 42, 50, 38, 29, 17, 4,
	62, 47, 59, 36, 45, 43, 51, 22, 53, 39, 33, 30, 24, 18, 12, 5,
	63, 55, 48, 27, 60, 41, 37, 16, 46, 35, 44, 21, 52, 32, 23, 11,
	54, 26, 40, 15, 34, 20, 31, 10, 25, 14, 19, 9, 13, 8, 7, 6,
}

const deBruijn64 = 0x03f79d71b4ca8b09

// removeFirst clears the first set bit (that is, resets the least significant
// set bit to 0).
func (b bitset) removeFirst() bitset {
	return b & (b - 1)
}

// removeBelow clears all set bits below slot i (non-inclusive).
func (b bitset) removeBelow(i uintptr) bitset {
	return bitsetRemoveBelow(b, i)
}

// Portable implementation of removeBelow.
//
// On AMD64, this is replaced with an intrisic that clears the lower i bits.
func bitsetRemoveBelow(b bitset, i uintptr) bitset {
	// Clear all bits below slot i's byte.
	mask := (uint64(1) << (8 * uint64(i))) - 1
	return b &^ bitset(mask)
}
