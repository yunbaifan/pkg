package jwt

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestMain(m *testing.M) {
	NewJwtToken("123123123")
	m.Run()
}
func Test_GenerateToken(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	base := BaseClaims{
		UUID:        uuid.NewV4(),
		ID:          1,
		Username:    "admin",
		NickName:    "nickName",
		AuthorityId: 1,
	}
	inter, err := JWTToken.GenerateToken(base, "issuer")
	data, err := JWTToken.ParseToken(inter)

	t.Log(data, err)
}
