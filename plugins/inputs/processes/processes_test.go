// +build linux

package processes

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/lavaorg/telex/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcesses(t *testing.T) {
	processes := &Processes{
		readProcFile: readProcFile,
	}
	var acc testutil.Accumulator

	err := processes.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasInt64Field("processes", "running"))
	assert.True(t, acc.HasInt64Field("processes", "sleeping"))
	assert.True(t, acc.HasInt64Field("processes", "stopped"))
	assert.True(t, acc.HasInt64Field("processes", "total"))
	total, ok := acc.Get("processes")
	t.Logf("processes:%v\n", total)
	require.True(t, ok)
	assert.True(t, total.Fields["total"].(int64) > 0)
}

func TestFromProcFiles(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("This test only runs on linux")
	}
	tester := tester{}
	processes := &Processes{
		readProcFile: tester.testProcFile,
	}

	var acc testutil.Accumulator
	err := processes.Gather(&acc)
	require.NoError(t, err)

	fields := getEmptyFields()
	fields["sleeping"] = tester.calls
	fields["total_threads"] = tester.calls * 2
	fields["total"] = tester.calls

	acc.AssertContainsTaggedFields(t, "processes", fields, map[string]string{})
}

func TestFromProcFilesWithSpaceInCmd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("This test only runs on linux")
	}
	tester := tester{}
	processes := &Processes{
		readProcFile: tester.testProcFile2,
	}

	var acc testutil.Accumulator
	err := processes.Gather(&acc)
	require.NoError(t, err)

	fields := getEmptyFields()
	fields["sleeping"] = tester.calls
	fields["total_threads"] = tester.calls * 2
	fields["total"] = tester.calls

	acc.AssertContainsTaggedFields(t, "processes", fields, map[string]string{})
}

// struct for counting calls to testProcFile
type tester struct {
	calls int64
}

func (t *tester) testProcFile(_ string) ([]byte, error) {
	t.calls++
	return []byte(fmt.Sprintf(testProcStat, "S", "2")), nil
}

func (t *tester) testProcFile2(_ string) ([]byte, error) {
	t.calls++
	return []byte(fmt.Sprintf(testProcStat2, "S", "2")), nil
}

const testProcStat = `10 (rcuob/0) %s 2 0 0 0 -1 2129984 0 0 0 0 0 0 0 0 20 0 %s 0 11 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 0 0 0 0 0 0 0 0 0 0 0 0 0 0
`

const testProcStat2 = `10 (rcuob 0) %s 2 0 0 0 -1 2129984 0 0 0 0 0 0 0 0 20 0 %s 0 11 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 0 0 0 0 0 0 0 0 0 0 0 0 0 0
`
