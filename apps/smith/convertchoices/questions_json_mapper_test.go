package convertchoices

import (
	"encoding/json"
	"testing"
)

func TestQuestionsJSONListsTheLevelAndTheFourChoices(t *testing.T) {
	raw, err := QuestionsJSON()
	if err != nil {
		t.Fatal(err)
	}
	var got []questionJSON
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	keys := []string{}
	byKey := map[string]questionJSON{}
	for _, q := range got {
		keys = append(keys, q.Key)
		byKey[q.Key] = q
	}
	want := []string{"level", "lint", "test", "build", "ci"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	}
	if byKey["level"].Default != "conservative" || byKey["level"].Flag != "level" || len(byKey["level"].When) != 0 {
		t.Errorf("level question = %+v", byKey["level"])
	}
	for _, key := range []string{"lint", "test", "build", "ci"} {
		q := byKey[key]
		if q.When["level"] != "full" || q.Flag != key || q.Default != q.Options[0].Value {
			t.Errorf("%s question = %+v", key, q)
		}
	}
	if byKey["lint"].Default != Keep || byKey["build"].Default != Smith || len(byKey["test"].Options) != 3 {
		t.Errorf("defaults or answers are not smith's: lint %+v, build %+v, test %+v", byKey["lint"], byKey["build"], byKey["test"])
	}
}
