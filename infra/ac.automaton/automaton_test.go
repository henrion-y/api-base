package ac_automaton

import (
	"testing"
)

func TestAcAutoMachine_Query(t *testing.T) {
	ac := NewAcAutoMachine()
	ac.AddPattern("花儿")
	ac.AddPattern("这样")
	ac.AddPattern("红")
	ac.Build()
	content := "我是红领巾，我想问下各位，花儿为什么这样红？"
	results := ac.Query(content)
	for _, result := range results {
		t.Log(result)
	}
}
