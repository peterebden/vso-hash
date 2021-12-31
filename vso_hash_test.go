package vsohash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAlgorithmIsStable(t *testing.T) {
	// This case is taken from VsoHashAlgorithmTests.cs
	const expected = "36668b653db0b48d3aa1f2fddcea481b34a310c166b9b041a5b23b59be02e5db00"
	checksum := Sum(make([]byte, 5*BlockSize))
	assert.Equal(t, expected, hex.EncodeToString(checksum[:]))
}

func TestBlockHashesDoNotChange(t *testing.T) {
	// These cases are taken from VsoHashTests.cs
	for lim, hash := range map[int]string{
		0:             "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		1:             "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PageSize - 1:  "12078762b8eda8a5499c46e9b5c7f8d37bab3a684571aefc2a7d3abcd56c093e",
		PageSize:      "b00be365b41949cf3571f69f8f5ad95514f6afdfe094aba614ecd34bd828272b",
		PageSize + 1:  "4a3e85babdd4243495a3617e9316bdf9cdc4526f97aa0e435a47226876c3d167",
		BlockSize - 1: "3492b19ddcd76ea1ed5c07090a021705adc7e7d5c5aad4ff619fd12faeceb197",
		BlockSize:     "e8deef25ed53357d2a738d7156067e69892a7bdc190818cd2ad698a3a1f95e03",
	} {
		t.Run(fmt.Sprintf("Sequential%d", lim), func(t *testing.T) {
			in := make([]byte, lim)
			for i := 0; i < lim; i++ {
				in[i] = byte(i & 0xff)
			}
			h := New()
			h.Write(in)
			h.Sum(nil)
			sum := lastBlockSum(h.(*vsoHash))
			assert.Equal(t, hash, hex.EncodeToString(sum))
		})
	}
}

// lastBlockSum is a helper for tests; it is pretty nasty in terms of how deeply it reaches into the
// internals of the hasher, but we don't provide block hashing as an external implementation and it's
// hard to do so without providing a heap of custom entry points.
func lastBlockSum(v *vsoHash) []byte {
	if b := v.blobID.Bytes(); len(b) > sha256.Size {
		return b[len(b)-sha256.Size-1 : len(b)-1]
	}
	h := sha256.Sum256(v.buffer.Bytes())
	return h[:]
}

func TestBlobIDsDoNotChange(t *testing.T) {
	// These cases are taken from VsoHashTests.cs
	for lim, hash := range map[int]string{
		0:               "1e57cf2792a900d06c1cdfb3c453f35bc86f72788aa9724c96c929d1cc6b456a00",
		1:               "3da32150b5e69b54e7ad1765d9573bc5e6e05d3b6529556c1b4a436a76a511f400",
		PageSize - 1:    "4ae1ad6462d75d117a5dafcf98167981371a4b21e1cee49d0b982de2ce01032300",
		PageSize:        "85840e1cb7cbfd78b464921c54c96f68c19066f20860efa8cce671b40ba5162300",
		PageSize + 1:    "d92a37c547f9d5b6b7b791a24f587da8189cca14ebc8511d2482e7448763e2bd00",
		BlockSize - 1:   "1c3c73f7e829e84a5ba05631195105fb49e033fa23bda6d379b3e46b5d73ef3700",
		BlockSize:       "6dae3ed3e623aed293297c289c3d20a53083529138b7631e99920ef0d93af3cd00",
		BlockSize + 1:   "1f9f3c008ea37ecb65bc5fb14a420cebb3ca72a9601ec056709a6b431f91807100",
		2*BlockSize - 1: "df0e0db15e866592dbfa9bca74e6d547d67789f7eb088839fc1a5cefa862353700",
		2 * BlockSize:   "5e3a80b2acb2284cd21a08979c49cbb80874e1377940699b07a8abee9175113200",
		2*BlockSize + 1: "b9a44a420593fa18453b3be7b63922df43c93ff52d88f2cab26fe1fadba7003100",
	} {
		t.Run(fmt.Sprintf("Sequential%d", lim), func(t *testing.T) {
			in := make([]byte, lim)
			for i := 0; i < lim; i++ {
				in[i] = byte(i & 0xff)
			}
			sum := Sum(in)
			assert.Equal(t, hash, hex.EncodeToString(sum[:]))
		})
	}
}
