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
	uint(18238943342984432443),
	[]int{2, -1432478247832253454, 3, -7},
	[]uint{2, 1, 3, 7},
	[5]int{1, -43432, 3, 4, 5},
	[5]uint{1, 2, 3, 4, 5},
	"twoja stara XDDD",
	"źóżź∂ż∆ż∂",
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
