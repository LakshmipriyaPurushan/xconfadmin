package rfc

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func TestFeatureRuleMarshaling(t *testing.T) {

	src := `{
    "applicationType": "stb",
    "featureIds": [
        "d471efce-b7d6-4419-a40e-5a095e8b6318",
        "7a98f5d9-9652-47a4-9ee9-4814db8aaa24"
    ],
    "id": "8a0dce3d-0f98-4cd5-8d93-cdb9cefb5211",
    "name": "Test_BLE_NS",
    "priority": 1,
    "rule": {
        "compoundParts": [],
        "condition": {
            "fixedArg": {
                "bean": {
                    "value": {
                        "java.lang.String": "34:1F:E4:B7:5E:D0"
                    }
                }
            },
            "freeArg": {
                "name": "estbMacAddress",
                "type": "STRING"
            },
            "operation": "IS"
        },
        "negated": false
    }
}`

	var featureRule FeatureRule
	err := json.Unmarshal([]byte(src), &featureRule)
	assert.NilError(t, err)

	t.Logf("\n\nfeatureRule = %v\n\n", featureRule)

	t.Logf("\n\nfeatureRule.Rule = %v\n\n", featureRule.Rule)

	t.Logf("\n\nfeatureRule.Rule.Condition = %v\n\n", featureRule.Rule.Condition)

	t.Logf("\n\nfeatureRule.Rule.Condition.FixedArg = %v\n\n", featureRule.Rule.Condition.FixedArg)

	t.Logf("\n\nfeatureRule.Rule.Condition.FreeArg = %v\n\n", featureRule.Rule.Condition.FreeArg)
}
