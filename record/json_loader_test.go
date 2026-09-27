package record

import (
	"fmt"
	"testing"
)

func TestLoadData(t *testing.T) {
	data1, data2, err := LoadTalkingRecordsFromLocalJsonFile("../bench/talking_records_demo_CLCBench.json", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(data1, data2)
}
