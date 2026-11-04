package discid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Jacques Loussier - Beyond Bach, disc 1: TOC and disc ID as reported by whipper.
var loussier = TOC{
	First: 1, Last: 10, Leadout: 303068,
	Offsets: []int{150, 40262, 59754, 77886, 98395, 116090, 139016, 161180, 238570, 274663},
}

func TestDiscID(t *testing.T) {
	assert.Equal(t, "KvJxY_MY82qy2MVu2GhP2r5NAm4-", loussier.DiscID())
}
