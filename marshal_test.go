package lightweight_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/litwinow/lightweight"
)

var tts = []interface{}{
	int(-2137),
	int8(127),
	int16(12744),
	int32(-241849124),
	int64(-4328494328943244324),
	uint(18238943342984432443),
	uint8(234),
	uint16(43289),
	uint32(432942389),
	uint64(5932849324429344324),
	[]int{2, -1432478247832253454, 3, -7},
	[]uint{2, 1, 3, 7},
	[5]int{1, -43432, 3, 4, 5},
	[5]uint{1, 2, 3, 4, 5},
	"twoja stara XDDD",
	"źóżź∂ż∆ż∂",
	true,
	false,
	float64(48912348912389231.3213123123),
	float32(321321332.3213123123),
}

func TestMarshal(t *testing.T) {
	for _, in := range tts {
		t.Run(fmt.Sprintf("%T", in), func(t *testing.T) {
			raw, err := lightweight.Marshal(in)
			require.NoError(t, err)
			t.Logf("size: %v", len(raw))
			out := reflect.New(reflect.TypeOf(in))
			require.NoError(t, lightweight.Unmarshal(raw, out.Interface()))

			require.Equal(t, in, out.Elem().Interface())
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	for _, in := range tts {
		b.Run(fmt.Sprintf("%T", in), func(b *testing.B) {
			raw, err := lightweight.Marshal(in)
			require.NoError(b, err)
			out := reflect.New(reflect.TypeOf(in))
			require.NoError(b, lightweight.Unmarshal(raw, out.Interface()))

			require.Equal(b, in, out.Elem().Interface())
		})
	}
}
