package uid16r

import (
	"encoding/binary"
	"time"
	"fmt"
	"sync"
	"errors"
	"math/rand"
	"encoding/base32"
)

const (
	epochOffset = uint64(1520845232285679425)
	EncodeStd   = "0345678abcdefghijklmnopqrstuvwxy"
	max         = uint64(1<<64 - 1)
	Size        = 16
	leftPad     = 'y'
)

var (
	EncodedSize = encoding.EncodedLen(Size)
)

type UId16rGen struct {
	clockOffset uint64
	lastTime    uint64
	lastSeq     uint8
	clockMutex  sync.Mutex
	timeFunc    func() uint64
	randFunc    func([]byte)
}

func NewUId16rGen() *UId16rGen {
	g := UId16rGen{
		lastSeq:  0xff,
		timeFunc: defaultTimeFunc,
		randFunc: defaultRandFunc,
	}
	return &g
}

var encoding = base32.NewEncoding(EncodeStd).WithPadding(base32.NoPadding)

type UId16r [Size]byte

func (u UId16r) Bytes() []byte {
	return u[:]
}

func (u UId16r) String() string {
	return encoding.EncodeToString(u[:])
}

func (u *UId16r) Shorten() string {
	input := u.String()
	i := 0
	for n := len(input) - 1; i < n && input[i] == leftPad; i++ {

	}
	return string([]byte(input)[i:])
}

func (gen *UId16rGen) FromString(input string) (u UId16r, err error) {
	v := []byte(input)
	k := EncodedSize - len(v)
	if k < 0 {
		err = errors.New("uid16r: invalid encoding")
		return
	} else if k > 0 {
		w := make([]byte, EncodedSize)
		for i := 0; i < k; i++ {
			w[i] = leftPad
		}
		copy(w[k:], input)
		v = w
	}
	_, err = encoding.Decode(u[:], v)
	return
}

func (gen *UId16rGen) FromBytes(input []byte) (u UId16r, err error) {
	l := len(input)
	if l != Size {
		err = fmt.Errorf("uid16r: must be %d bytes max long", Size)

	} else {
		copy(u[:], input)
	}
	return
}

func (gen *UId16rGen) New() UId16r {
	timeNow := gen.timeFunc()
	gen.clockMutex.Lock()
	if timeNow < gen.lastTime {
		gen.clockOffset = (gen.lastTime - timeNow) + 1
	} else
	if timeNow == gen.lastTime {
		//chk it's not zero (!)
		gen.lastSeq--
	} else {
		gen.lastSeq = 0xff
	}
	gen.lastTime = timeNow
	seq := gen.lastSeq
	timeNow = max - timeNow - gen.clockOffset
	gen.clockMutex.Unlock()
	u := UId16r{}
	binary.BigEndian.PutUint64(u[0:], timeNow)
	u[8] = seq
	gen.randFunc(u[9:])
	return u
}

func defaultRandFunc(slice []byte) {

	rand.Read(slice)
}

func defaultTimeFunc() uint64 {
	return uint64(time.Now().UnixNano()) - epochOffset
}

func maxId16r() UId16r {
	u := UId16r{}
	for i := 0; i < len(u); i++ {
		u[i] = 255
	}
	return u
}

func minId16r() UId16r {
	u := UId16r{}
	for i := 0; i < len(u); i++ {
		u[i] = 0
	}
	return u
}
