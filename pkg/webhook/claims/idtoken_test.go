package claims

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GroupTranslation(t *testing.T) {
	baseIDToken := map[string]interface{}{
		"groups": []string{"group1", "group2"},
	}
	receivedIDToken := map[string]interface{}{
		"groups": []string{"test2:group1", "test2:group2"},
	}
	res, err := GenerateTokenFromTemplate(baseIDToken, receivedIDToken)
	assert.NoError(t, err)
	assert.Equal(t, res, []byte(`{"groups":["test2:group1","test2:group2"]}`))
}

func Test_EmptyInput(t *testing.T) {
	res, err := GenerateTokenFromTemplate(map[string]interface{}{}, map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, res, []byte(`{}`))
}

func Test_ProtectedClaims(t *testing.T) {
	baseIDToken := map[string]interface{}{
		"iss":    "iss",
		"sub":    "sub",
		"aud":    "aud",
		"groups": []string{"group1", "group2"},
	}
	receivedIDToken := map[string]interface{}{
		"iss":    "iss2",
		"groups": []string{"test2:group1", "test2:group2"},
	}
	res, err := GenerateTokenFromTemplate(baseIDToken, receivedIDToken)
	assert.NoError(t, err)
	assert.Equal(t, res, []byte(`{"aud":"aud","groups":["test2:group1","test2:group2"],"iss":"iss","sub":"sub"}`))
}

func Test_NotStandardClaims(t *testing.T) {
	baseIDToken := map[string]interface{}{
		"groups": []string{"group1", "group2"},
	}
	receivedIDToken := map[string]interface{}{
		"groups": []string{"test2:group1", "test2:group2"},
		"custom": "custom",
	}
	res, err := GenerateTokenFromTemplate(baseIDToken, receivedIDToken)
	assert.NoError(t, err)
	assert.Equal(t, res, []byte(`{"custom":"custom","groups":["test2:group1","test2:group2"]}`))
}
