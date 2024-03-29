package jwt

import (
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type BaseClaims struct {
	UUID        uuid.UUID
	ID          int
	Username    string
	NickName    string
	AuthorityId uint
}
type JWTInter interface {
	GenerateToken(baseClaims BaseClaims, issuer string) (string, error)
	ParseToken(tokenString string) (*CustomClaims, error)
}

var (
	JWTToken JWTInter = (*jwtToken)(nil)
	once     sync.Once
)

func NewJwtToken(secret string) JWTInter {
	once.Do(func() {
		JWTToken = &jwtToken{
			secret: secret,
		}
	})
	return JWTToken
}

type CustomClaims struct {
	BaseClaims
	jwt.StandardClaims
}

type jwtToken struct {
	secret string
}

func (j *jwtToken) GenerateToken(baseClaims BaseClaims, issuer string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)
	claims := CustomClaims{
		baseClaims,
		jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 1000, // 签名生效时间
			ExpiresAt: expireTime.Unix(),
			Issuer:    issuer, // 签名的发行者
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(j.secret))
	return token, err
}

func (j *jwtToken) ParseToken(tokenString string) (claims *CustomClaims, err error) {
	// 使用jwt.ParseWithClaims方法解析token，这个token是前端传给我们的,获得一个*Token类型的对象
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		// 处理token解析后的各种错误
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors == jwt.ValidationErrorExpired {
				return nil, errors.New("登录已过期，请重新登录")
			} else {
				return nil, errors.New("token不可用," + err.Error())
			}
		}
	}
	// 转换为*CustomClaims类型并返回
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// 如果解析成功并且token是可用的
		return claims, nil
	}
	return nil, errors.New("解析token失败")
}
