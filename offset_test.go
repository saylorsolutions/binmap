package bin

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func testInitFile(t *testing.T, data []byte) string {
	t.Helper()
	file, err := os.CreateTemp("", "testInitFile-*")
	require.NoError(t, err)
	require.NoError(t, file.Close())
	filename := file.Name()
	t.Cleanup(func() {
		t.Log("Removing temp file:", filename)
		assert.NoError(t, os.Remove(filename))
	})
	require.NoError(t, os.WriteFile(filename, data, 0600))
	return filename
}

func testWithTmpFile(t *testing.T, tmpFile string, user func(file *os.File) error) {
	t.Helper()
	file, err := os.OpenFile(tmpFile, os.O_RDWR, 0600)
	require.NoError(t, err, "Failed to open file:", err)
	defer func() {
		require.NoError(t, file.Close(), "Failed to close file")
	}()
	require.NoError(t, user(file))
}

func testReadFile(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(file)
	require.NoError(t, err, "Failed to read file state:", err)
	return data
}

func TestOffsetData_Write(t *testing.T) {
	writeCases := map[string]struct {
		offset   int64
		data     []byte
		length   uint64
		expected []byte
	}{
		"Write offset from end": {
			offset:   -2,
			data:     []byte{1, 2, 3, 4},
			length:   4,
			expected: []byte{0, 0, 1, 2, 3, 4},
		},
		"Write past end": {
			offset:   5,
			data:     []byte{1, 2, 3, 4},
			length:   4,
			expected: []byte{0, 0, 0, 0, 0, 1, 2, 3, 4},
		},
		"Write overlapping end": {
			offset:   3,
			data:     []byte{1, 2, 3, 4},
			length:   4,
			expected: []byte{0, 0, 0, 1, 2, 3, 4},
		},
		"Write at end": {
			offset:   4,
			data:     []byte{1, 2, 3, 4},
			length:   0,
			expected: []byte{0, 0, 0, 0, 1, 2, 3, 4},
		},
		"Write at beginning": {
			offset:   0,
			data:     []byte{1, 2, 3, 4},
			length:   4,
			expected: []byte{1, 2, 3, 4},
		},
		"Write near beginning": {
			offset:   1,
			data:     []byte{1, 2, 3, 4},
			length:   0,
			expected: []byte{0, 1, 2, 3, 4},
		},
	}
	for name, tc := range writeCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tmpFile := testInitFile(t, []byte{0, 0, 0, 0})
			testWithTmpFile(t, tmpFile, func(file *os.File) error {
				return OffsetData(&tc.offset, &tc.data, &tc.length).Write(file, BigEndian)
			})
			result := testReadFile(t, tmpFile)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOffsetData_Read(t *testing.T) {
	readCases := map[string]struct {
		offset   int64
		length   uint64
		expected []byte
	}{
		"Read from end": {
			offset:   -2,
			length:   0,
			expected: []byte{3, 4},
		},
		"Read from end with length": {
			offset:   -3,
			length:   2,
			expected: []byte{2, 3},
		},
		"Read at end": {
			offset:   5,
			length:   0,
			expected: []byte{2, 3, 4},
		},
		"Read at end with length": {
			offset:   6,
			length:   2,
			expected: []byte{3, 4},
		},
		"Read at beginning": {
			offset:   0,
			length:   0,
			expected: []byte{0, 0, 0, 0, 1, 2, 3, 4},
		},
		"Read at beginning with length": {
			offset:   0,
			length:   5,
			expected: []byte{0, 0, 0, 0, 1},
		},
		"Read near beginning": {
			offset:   2,
			length:   0,
			expected: []byte{0, 0, 1, 2, 3, 4},
		},
		"Read near beginning with length": {
			offset:   2,
			length:   3,
			expected: []byte{0, 0, 1},
		},
	}
	for name, tc := range readCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tmpFile := testInitFile(t, []byte{0, 0, 0, 0, 1, 2, 3, 4})
			var result []byte
			testWithTmpFile(t, tmpFile, func(file *os.File) error {
				return OffsetData(&tc.offset, &result, &tc.length).Read(file, BigEndian)
			})
			assert.Equal(t, tc.expected, result)
		})
	}
}
