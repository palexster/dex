package claims

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func getProtectedClaims() []string {
	return []string{"iss", "sub", "aud", "exp", "iat", "azp", "nonce", "at_hash", "c_hash"}
}

func generateIDClaims(baseIDClaims map[string]interface{}, customClaims map[string]interface{}) map[string]interface{} {
	finalClaims := map[string]interface{}{}
	maps.Copy(finalClaims, baseIDClaims)
	// Adding the immutable claims to the token
	protectedClaims := getProtectedClaims()
	for claim := range customClaims {
		if !slices.Contains(protectedClaims, claim) {
			finalClaims[claim] = customClaims[claim]
		}
	}
	return finalClaims
}

func GenerateTokenFromTemplate(baseIDClaims map[string]interface{}, customClaims map[string]interface{}) ([]byte,
	error,
) {
	payload, err := json.Marshal(generateIDClaims(baseIDClaims, customClaims))
	if err != nil {
		return []byte{}, fmt.Errorf("could not serialize claims: %v", err)
	}
	return payload, nil
}
