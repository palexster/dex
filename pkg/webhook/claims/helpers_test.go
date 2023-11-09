package claims

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstrainScope(t *testing.T) {
	res := constrainScope(map[string]interface{}{
		"claim1": "value1",
		"claim2": "value2",
	}, []string{"claim1", "claim3"})
	assert.Equal(t, res, map[string]interface{}{
		"claim1": "value1",
	})
}
