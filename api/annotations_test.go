package api_test

import (
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
