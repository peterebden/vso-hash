// Package vsohash provides an implementation of the BuildXL paged VSO-Hash.
// This is based on SHA256 but allows parallelising calculations on large inputs.
//
// See https://github.com/microsoft/BuildXL/blob/master/Documentation/Specs/PagedHash.md
// for more details and the specification.
package vsohash

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"runtime"
)

// Size is the number of bytes of the output hash.
const Size = sha256.Size + 1

// BlockSize is the size of the VSO-Hash blocks. It's defined as always being 2MB.
const BlockSize = 2 * 1024 * 1024

// PageSize is the size of the pages within each block. They're always 64kb
const PageSize = 64 * 1024

const pagesPerBlock = BlockSize / PageSize

const seed = "VSO Content Identifier Seed"

// New returns a new hash. It will perform up to GOMAXPROCS calculations in parallel.
//
// Note that the returned hash does not fully faithfully implement the semantics of Sum(); it does update
// the underlying state (it's quite difficult to implement Go's semantics here).
// After calling Sum, the caller should not call other functions on the hash object.
func New() hash.Hash {
	return NewParallel(runtime.GOMAXPROCS(0))
}

// NewParallel returns a new hash. It will perform up to the given number of calculations in parallel.
//
// Note that the returned hash does not faithfully implement the semantics of Sum(); it does update
// the underlying state (it's quite difficult to implement Go's semantics here).
// After calling Sum, the caller should not call other functions on the hash object.
func NewParallel(parallelism int) hash.Hash {
	if parallelism <= 0 {
		panic("Parallelism must be strictly positive")
	}
	v := &vsoHash{
		tasks:      make(chan hashTask, parallelism),
		pageHashes: make([]<-chan [sha256.Size]byte, 0, pagesPerBlock),
	}
	v.buffer.Grow(PageSize)
	v.blobID.Grow(2*Size + 1)
	for i := 0; i < parallelism; i++ {
		go v.run()
	}
	runtime.SetFinalizer(v, finalize)
	return v
}

type vsoHash struct {
	// The running buffer of the current page
	buffer bytes.Buffer
	// The calculations of the current set of page hashes
	pageHashes []<-chan [sha256.Size]byte
	// The current blob id (updated as we run through the hash)
	blobID bytes.Buffer
	// The set of waiting hash tasks
	tasks chan hashTask
}

type hashTask struct {
	Input  []byte
	Output chan [sha256.Size]byte
}

// finalize is a GC finalizer function that is run when this hash is collected.
// It closes the internal task channel which permits the background goroutines to exit.
func finalize(v *vsoHash) {
	close(v.tasks)
}

func (v *vsoHash) run() {
	for task := range v.tasks {
		task.Output <- sha256.Sum256(task.Input)
	}
}

func (v *vsoHash) Write(in []byte) (int, error) {
	// Write one page at a time
	for {
		// If this data fits within the buffer and doesn't finish a page, just keep it for later.
		if len(in)+v.buffer.Len() < PageSize {
			v.buffer.Write(in)
			return len(in), nil
		}
		// If we've got some data written already, take it into account.
		// From above we know that we will finish at least one page here.
		if v.buffer.Len() > 0 {
			n := PageSize - v.buffer.Len()
			v.buffer.Write(in[:n])
			in = in[n:]
			// We must copy the contents of the buffer since we'll keep it around asynchronously.
			// TODO(peterebden): maybe pool these objects?
			b := [PageSize]byte{}
			copy(b[:], v.buffer.Bytes())
			v.writePage(b[:])
			v.buffer.Reset()
			continue
		}
		// If we get here, there is at least one page size left and nothing in the buffer; write it directly.
		v.writePage(in[:PageSize])
		in = in[PageSize:]
	}
}

// writePage writes one more page to the hash.
func (v *vsoHash) writePage(page []byte) {
	ch := make(chan [sha256.Size]byte, 1)
	v.tasks <- hashTask{Input: page, Output: ch}
	v.pageHashes = append(v.pageHashes, ch)
	// Now see if we need to finish a block.
	if len(v.pageHashes) == pagesPerBlock {
		v.finishBlock()
	}
}

// finishBlock finishes a block and adds it to the current running hash.
func (v *vsoHash) finishBlock() {
	var buf bytes.Buffer
	buf.Grow(pagesPerBlock * sha256.Size)
	for _, page := range v.pageHashes {
		b := <-page
		buf.Write(b[:])
	}
	v.pageHashes = v.pageHashes[:0]
	h := sha256.Sum256(buf.Bytes())
	v.updateBlobID(h[:])
}

// updateBlobID updates the running blob id with the given hash.
// It's a bit fiddly because we have to write different things based on whether we are
// the last block or not, which we generally don't know at the time we do it :(
func (v *vsoHash) updateBlobID(h []byte) {
	if v.blobID.Len() == 0 {
		v.blobID.WriteString(seed)
		v.blobID.Write(h)
		return
	}
	v.blobID.WriteByte(0) // that wasn't the last block
	sum := sha256.Sum256(v.blobID.Bytes())
	v.blobID.Reset()
	v.blobID.Write(sum[:])
	v.blobID.Write(h)
}

// Sum appends the current hash to b and returns the resulting slice.
// As noted above, it currently _does_ change the underlying hash state (it is not easy to
// copy one of these as the stdlib builtin ones do).
// I'm not certain if anyone will really care about this; I only ever seem to call Sum()
// once at the end of a hash but if we _really_ cared we could probably try to modify
// things to support this.
func (v *vsoHash) Sum(b []byte) []byte {
	s := v.sum()
	return append(b, s[:]...)
}

// sum calculates and returns the current hash. Underlying state is updated.
func (v *vsoHash) sum() [Size]byte {
	if v.buffer.Len() != 0 {
		// We have some pending bytes. Add a task for them.
		// Note that we can do this synchronously since we know we won't do anything else with the buffer.
		v.writePage(v.buffer.Bytes())
	}
	if len(v.pageHashes) > 0 || v.blobID.Len() == 0 {
		// We have some pages left, must finish off the last block.
		// Must also ensure this happens at least once if we never write anything to the hash.
		v.finishBlock()
	}
	v.blobID.WriteByte(1) // this is the last block
	b := sha256.Sum256(v.blobID.Bytes())
	ret := [Size]byte{}
	copy(ret[:], b[:])
	ret[Size-1] = 0
	return ret
}

func (v *vsoHash) Reset() {
	v.pageHashes = make([]<-chan [sha256.Size]byte, 0, pagesPerBlock)
	v.blobID.Reset()
	v.buffer.Reset()
}

func (v *vsoHash) Size() int {
	return Size
}

// PageSize is more appropriate here than BlockSize; we write a page at a time which is mildly
// more efficient for us, but there is little difference to writing a whole block at a time.
func (v *vsoHash) BlockSize() int {
	return PageSize
}

// Sum calculates the VSO-Hash for the given input.
func Sum(in []byte) [Size]byte {
	h := New().(*vsoHash)
	h.Write(in)
	return h.sum()
}
