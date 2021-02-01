package fixtures_test

import "github.com/adamluzsi/testcase/fixtures"

func ExampleSkipByTag() {
	type Entity struct {
		ID    string `external-resource:"ID" json:"id"`
		Value string `external-resource:"something" json:"value"`
	}

	skipByTagOption := fixtures.SkipByTag("external-resource", "ID")
	ent := fixtures.New(Entity{}, skipByTagOption).(*Entity)
	_ = ent.ID    // no value populated
	_ = ent.Value // value populated
}
