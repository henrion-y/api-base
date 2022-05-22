package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/henrion-y/api-base/utils"
)

type JWTAuthMiddleware struct {
	authService utils.AuthService
}

func NewJWTAuthMiddleware(authService utils.AuthService) (*JWTAuthMiddleware, error) {
	return &JWTAuthMiddleware{authService: authService}, nil
}

// Handler JWTAuthMiddleware 中间件，检查token
func (m *JWTAuthMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"code":    -1,
				"message": "无权限访问，请求未携带token",
			})
			ctx.Abort() // 结束后续操作
			return
		}

		// 按空格拆分
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    -1,
				"message": "请求头中auth格式有误",
			})
			ctx.Abort()
			return
		}

		// 解析token包含的信息
		claims, err := m.authService.ParseToken(parts[1])
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    -1,
				"message": "无效的Token",
			})
			ctx.Abort()
			return
		}

		// 将当前请求的claims信息保存到请求的上下文c上
		ctx.Set("claims", claims)
		ctx.Next() // 后续的处理函数可以用过ctx.Get("claims")来获取当前请求的用户信息
	}
}
