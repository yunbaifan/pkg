package sortedmap

import "testing"

var (
	_orderedMap *OrderedMap[string, int]
)

func TestMain(m *testing.M) {
	_orderedMap = NewInit[string, int]()
	m.Run()
}

func Test_Set(t *testing.T) {
	_orderedMap.Set("a", 1)
	t.Log(_orderedMap.Get("a"))
}

func Test_GetEntryMaps(t *testing.T) {

	type args struct {
		key string
		val int
	}

	tests := []args{
		{"l", 1},
		{"f", 1},
		{"v", 1},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if val, ok := _orderedMap.Set(tt.key, tt.val); ok {
				if val != tt.val {
					t.Errorf("Get() = %v, want %v", val, tt.val)
				}
			}
			t.Logf("%v", _orderedMap.GetMaps())
		})
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if val, ok := _orderedMap.Get(tt.key); ok {
				if val != tt.val {
					t.Errorf("Get() = %v, want %v", val, tt.val)
				}
			} else {
				t.Errorf("Get() = %v, want %v", ok, true)
			}
		})
	}
}
