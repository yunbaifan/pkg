package sortedmap

import (
	"testing"

	"gocv.io/x/gocv"
)

var (
	_orderedMap *OrderedMap[string, int]
)

func TestMain(m *testing.M) {
	_orderedMap = NewInit[string, int]()
	m.Run()
}

// func Test_Set(t *testing.T) {
// 	_orderedMap.Set("a", 1)
// 	t.Log(_orderedMap.Get("a"))
// }

func Test_GetEntryMaps(t *testing.T) {

	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		panic(err)
	}
	defer webcam.Close()

	window := gocv.NewWindow("Hello")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			println("Cannot read device")
			return
		}
		if img.Empty() {
			continue
		}

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
