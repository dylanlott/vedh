package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatsQuery(t *testing.T) {
	s := testAPI(t)
	formats, err := s.Formats(authCtx(mastershake))
	assert.NoError(t, err)
	assert.Len(t, formats, 2)
	assert.Equal(t, "EDH", formats[0].ID)
	assert.Equal(t, "GENERIC_DUEL", formats[1].ID)
	assert.NotEmpty(t, formats[0].Zones)
	assert.NotEmpty(t, formats[1].PhaseSequence)
}
