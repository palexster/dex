package claims

import "golang.org/x/exp/slices"

func constrainScope(claims map[string]interface{}, acceptedClaims []string) map[string]interface{} {
	scopedClaims := make(map[string]interface{})
	for k, v := range claims {
		if slices.Contains(acceptedClaims, k) {
			scopedClaims[k] = v
		}
	}
	return scopedClaims
}
