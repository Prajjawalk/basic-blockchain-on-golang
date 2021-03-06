package cuckoo

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
)

type bucket []fingerprint
type fingerprint []byte

var hasher = sha1.New()

const retries = 500 // how many times do we try to move items around?

// Cuckoo filter based on https://www.pdl.cmu.edu/PDL-FTP/FS/cuckoo-conext2014.pdf
type Cuckoo struct {
	buckets []bucket
	m       uint // buckets
	b       uint // entries per bucket
	f       uint // fingerprint length
	n       uint // filter capacity (rename cap?)
}

//func main() {
//	c := NewCuckoo(10, 0.1)
//	c.insert("hello")
//	c.insert("world")
//	ok := c.lookup("world")
//	fmt.Printf("%v\n", ok)
//	c.delete("world")
//	ok = c.lookup("world")
//	fmt.Printf("%v\n", ok)
//}

// n = len(items), fp = false positive rate

// NewCuckoo
func NewCuckoo(n uint, fp float64) *Cuckoo {
	b := uint(4)
	f := fingerprintLength(b, fp)
	m := nextPower(n / f * 8)
	buckets := make([]bucket, m)
	for i := uint(0); i < m; i++ {
		buckets[i] = make(bucket, b)
	}
	return &Cuckoo{
		buckets: buckets,
		m:       m,
		b:       b,
		f:       f,
		n:       n,
	}
}

// delete the fingerprint from the cuckoo filter
func (c *Cuckoo) Delete(needle string) {
	i1, i2, f := c.hashes(needle)
	// try to remove from f1
	b1 := c.buckets[i1%c.m]
	if ind, ok := b1.contains(f); ok {
		b1[ind] = nil
		return
	}

	b2 := c.buckets[i2%c.m]
	if ind, ok := b2.contains(f); ok {
		b2[ind] = nil
		return
	}
}

// lookup needle in the cuckoo filter
func (c *Cuckoo) Lookup(needle string) bool {
	i1, i2, f := c.hashes(needle)
	_, b1 := c.buckets[i1%c.m].contains(f)
	_, b2 := c.buckets[i2%c.m].contains(f)
	return b1 || b2
}

func (b bucket) contains(f fingerprint) (int, bool) {
	for i, x := range b {
		if bytes.Equal(x, f) {
			return i, true
		}
	}
	return -1, false
}

func (c *Cuckoo) Insert(input string) {
	i1, i2, f := c.hashes(input)
	// first try bucket one
	b1 := c.buckets[i1%c.m]
	if i, err := b1.nextIndex(); err == nil {
		b1[i] = f
		return
	}

	b2 := c.buckets[i2%c.m]
	if i, err := b2.nextIndex(); err == nil {
		b2[i] = f
		return
	}

	// else we need to start relocating items
	i := i1
	for r := 0; r < retries; r++ {
		index := i % c.m
		entryIndex := rand.Intn(int(c.b))
		// swap
		f, c.buckets[index][entryIndex] = c.buckets[index][entryIndex], f
		i = i ^ uint(binary.BigEndian.Uint32(hash(f)))
		b := c.buckets[i%c.m]
		if idx, err := b.nextIndex(); err == nil {
			b[idx] = f
			return
		}
	}
	panic("cuckoo filter full")
}

// nextIndex returns the next index for entry, or an error if the bucket is full
func (b bucket) nextIndex() (int, error) {
	for i, f := range b {
		if f == nil {
			return i, nil
		}
	}
	return -1, errors.New("bucket full")
}

// hashes returns h1, h2 and the fingerprint
func (c *Cuckoo) hashes(data string) (uint, uint, fingerprint) {
	h := hash([]byte(data))
	f := h[0:c.f]
	i1 := uint(binary.BigEndian.Uint32(h))
	i2 := i1 ^ uint(binary.BigEndian.Uint32(hash(f)))
	return i1, i2, fingerprint(f)
}

func hash(data []byte) []byte {
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)
	hasher.Reset()
	return hash
}

func fingerprintLength(b uint, e float64) uint {
	f := uint(math.Ceil(math.Log(2 * float64(b) / e)))
	f /= 8
	if f < 1 {
		return 1
	}
	return f
}

func nextPower(i uint) uint {
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}
