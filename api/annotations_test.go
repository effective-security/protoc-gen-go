package api_test

import (
	"encoding/json"
	"testing"

	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/stretchr/testify/assert"
)

func TestEnumDescription_Parse(t *testing.T) {
	ed := e2e.Role_EnumDescription
	assert.Equal(t, int32(e2e.Role_User), ed.Parse(e2e.Role_User))
	assert.Equal(t, int32(e2e.Role_Admin), ed.Parse(e2e.Role_Admin))
	assert.Equal(t, int32(e2e.Role_Unknown), ed.Parse(e2e.Role_Unknown))
	assert.Equal(t, int32(e2e.Role_User|e2e.Role_Admin), ed.Parse(e2e.Role_User|e2e.Role_Admin))
	assert.Equal(t, int32(e2e.Role_User|e2e.Role_Admin), ed.Parse([]string{e2e.Role_User.String(), e2e.Role_Admin.String()}))
	assert.Equal(t, int32(e2e.Role_User|e2e.Role_Admin), ed.Parse([]int32{int32(e2e.Role_User), int32(e2e.Role_Admin)}))
	assert.Equal(t, int32(e2e.Role_User|e2e.Role_Admin), ed.Parse([]int{int(e2e.Role_User), int(e2e.Role_Admin)}))
	assert.Equal(t, int32(e2e.Role_User|e2e.Role_Admin), ed.Parse("User,Admin"))
	assert.Equal(t, int32(e2e.Role_User|e2e.Role_Admin), ed.Parse("User|Admin"))
}

func TestEnumDescription_ParseEnum(t *testing.T) {
	ed := e2e.Role_EnumDescription
	assert.Equal(t, e2e.Role_User, ed.ParseEnum[e2e.Role]("User"))
	assert.Equal(t, e2e.Role_User|e2e.Role_Admin, ed.ParseEnum[e2e.Role]("User,Admin"))
	assert.Equal(t, e2e.Role_User|e2e.Role_Admin, e2e.Role(0).Parse("User|Admin"))
}

func TestEnumDescription_UnmarshalJSON(t *testing.T) {
	var e e2e.Role
	err := json.Unmarshal([]byte(`"User|Admin"`), &e)
	assert.NoError(t, err)
	assert.Equal(t, e2e.Role_User|e2e.Role_Admin, e)

	err = json.Unmarshal([]byte(`["User","Admin"]`), &e)
	assert.NoError(t, err)
	assert.Equal(t, e2e.Role_User|e2e.Role_Admin, e)

	err = json.Unmarshal([]byte(`[2,16]`), &e)
	assert.NoError(t, err)
	assert.Equal(t, e2e.Role_User|e2e.Role_Admin, e)

	err = json.Unmarshal([]byte("18"), &e)
	assert.NoError(t, err)
	assert.Equal(t, e2e.Role_User|e2e.Role_Admin, e)

	err = json.Unmarshal([]byte(`"Admin"`), &e)
	assert.NoError(t, err)
	assert.Equal(t, e2e.Role_Admin, e)

	// The bitmask value is not a valid enum value, so it returns the unknown value.
	err = json.Unmarshal([]byte(`"18"`), &e)
	assert.NoError(t, err)
	assert.Equal(t, e2e.Role_Unknown, e)

	var list []e2e.Role
	err = json.Unmarshal([]byte(`["User|Admin"]`), &list)
	assert.NoError(t, err)
	assert.Equal(t, []e2e.Role{e2e.Role_User | e2e.Role_Admin}, list)

	err = json.Unmarshal([]byte(`["User","Admin"]`), &list)
	assert.NoError(t, err)
	assert.Equal(t, []e2e.Role{e2e.Role_User, e2e.Role_Admin}, list)
}
