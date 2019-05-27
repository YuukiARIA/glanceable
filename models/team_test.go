package models

import (
	"reflect"
	"testing"
)

func Test_NewEmpty(t *testing.T) {
	got := NewEmpty()

	expected := Team{
		ID:   0,
		Name: "",
		Auth: map[string]*AuthRule{},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fail()
	}
}

func Test_NewFromJSON(t *testing.T) {
	json := `
{
    "id": 1,
    "name": "example",
    "auth": {
        "owner": {
            "users": ["user1", "user2"],
            "groups": ["group1", "group2"]
        },
        "member": {
            "users": [],
            "groups": ["group3"]
        }
    }
}
`
	got := NewFromJSON([]byte(json))

	expected := Team{
		ID:   1,
		Name: "example",
		Auth: map[string]*AuthRule{
			"owner": {
				RoleName: "owner",
				Users:    []string{"user1", "user2"},
				Groups:   []string{"group1", "group2"},
			},
			"member": &AuthRule{
				RoleName: "member",
				Users:    []string{},
				Groups:   []string{"group3"},
			},
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fail()
	}
}
