package helper

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestSizeWithBinaryUnit(t *testing.T) {
	defer goleak.VerifyNone(t)

	actual := SizeWithBinaryUnit(0)
	assert.Equal(t, "0 B", actual)

	actual = SizeWithBinaryUnit(1)
	assert.Equal(t, "1 B", actual)

	actual = SizeWithBinaryUnit(1023)
	assert.Equal(t, "1023 B", actual)

	actual = SizeWithBinaryUnit(1024)
	assert.Equal(t, "1.0 KiB", actual)

	actual = SizeWithBinaryUnit(2621440)
	assert.Equal(t, "2.5 MiB", actual)

	actual = SizeWithBinaryUnit(1 << 30)
	assert.Equal(t, "1.0 GiB", actual)

	actual = SizeWithBinaryUnit(1<<64 - 1)
	assert.Equal(t, "16.0 EiB", actual)

	//actual = SizeWithBinaryUnit(1 << 80) // Note, too large for the number representation!
	//assert.Equal(t, "1.0 YiB", actual)
}
