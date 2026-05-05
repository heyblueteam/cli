package common

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var cuidCounter uint32

// NewCuid returns a 25-char collision-resistant id compatible with the
// `cuid` JS library used by the API. Layout: 'c' + 8-char base36 timestamp +
// 4-char base36 counter + 4-char base36 fingerprint + 8 random base36 chars.
//
// The API's `isCuid` check only validates that the string starts with 'c',
// but downstream code expects roughly this format.
func NewCuid() string {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 36)
	ts = padLeft(ts, 8)

	c := atomic.AddUint32(&cuidCounter, 1) % (36 * 36 * 36 * 36)
	counter := padLeft(strconv.FormatUint(uint64(c), 36), 4)

	fp := fingerprint()

	var randBytes [5]byte
	_, _ = rand.Read(randBytes[:])
	randVal := binary.BigEndian.Uint64(append([]byte{0, 0, 0}, randBytes[:]...))
	rnd := padLeft(strconv.FormatUint(randVal, 36), 8)
	if len(rnd) > 8 {
		rnd = rnd[len(rnd)-8:]
	}

	return "c" + ts + counter + fp + rnd
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return fmt.Sprintf("%0*s", width, s)
}

func fingerprint() string {
	pid := uint32(os.Getpid())
	host, _ := os.Hostname()
	var h uint32 = 5381
	for i := 0; i < len(host); i++ {
		h = (h << 5) + h + uint32(host[i])
	}
	combined := (pid + h) % (36 * 36 * 36 * 36)
	return padLeft(strconv.FormatUint(uint64(combined), 36), 4)
}
