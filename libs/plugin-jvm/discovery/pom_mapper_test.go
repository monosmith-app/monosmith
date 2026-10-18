package discovery

import "testing"

func TestAPomsDescriptionIsRead(t *testing.T) {
	pom, err := ParsePom([]byte(`<project><groupId>acme</groupId><artifactId>api</artifactId><version>1.0.0</version><description>
  Serves the API.
</description></project>`))
	if err != nil || pom.Description != "Serves the API." {
		t.Errorf("%+v %v", pom, err)
	}
}
