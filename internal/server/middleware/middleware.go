package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"strings"
	"toDoList/internal"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"net/http"
	authErrors "toDoList/internal/server/auth/autherrors"
	auth "toDoList/internal/server/auth/user_auth"
)

type TokenSigner interface {
	NewAccessToken(userID string) (string, error)
	NewRefreshToken(userID string) (string, error)
	ParseAccessToken(token string, opt auth.ParseOptions) (*auth.Claims, error)
	ParseRefreshToken(token string, opt auth.ParseOptions) (*jwt.RegisteredClaims, error)
	GetIssuer() string
	GetAudience() string
}

//nolint:gocognit // сложная логика мидлваря — допускаем
func AuthMiddleware(signer TokenSigner) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := ctx.Cookie("access_token")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": authErrors.ErrMissingAccessToken})
			ctx.Abort()
			return
		}

		claims, err := signer.ParseAccessToken(accessToken, auth.ParseOptions{
			ExpectedIssuer:   signer.GetIssuer(),
			ExpectedAudience: signer.GetAudience(),
			AllowMethods:     []string{"HS256"},
			Leeway:           internal.MinOne,
		})

		//nolint:nestif // Нужно оставить ошибки на проверки
		if err != nil {
			if errors.Is(err, authErrors.ErrInvalidAccessToken) ||
				errors.Is(err, jwt.ErrTokenExpired) {
				refreshToken, errTok := ctx.Cookie("refresh_token")
				if errTok != nil {
					ctx.JSON(
						http.StatusUnauthorized,
						gin.H{"error": authErrors.ErrMissingRefreshToken},
					)
					ctx.Abort()
					return
				}

				refreshClaims, errParseRefresh := signer.ParseRefreshToken(
					refreshToken,
					auth.ParseOptions{
						ExpectedIssuer:   signer.GetIssuer(),
						ExpectedAudience: signer.GetAudience(),
						AllowMethods:     []string{"HS256"},
						Leeway:           internal.MinFive,
					},
				)
				if errParseRefresh != nil {
					ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
					ctx.Abort()
					return
				}

				newAccessToken, errParseAccess := signer.NewAccessToken(refreshClaims.Subject)
				if errParseAccess != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					ctx.Abort()
					return
				}

				ctx.SetCookie(
					"access_token",
					newAccessToken,
					internal.MaxAgeForAccessToken,
					"/",
					"127.0.0.1:8080",
					false,
					true,
				)

				claims, err = signer.ParseAccessToken(newAccessToken, auth.ParseOptions{
					ExpectedIssuer:   signer.GetIssuer(),
					ExpectedAudience: signer.GetAudience(),
					AllowMethods:     []string{"HS256"},
					Leeway:           internal.MinOne,
				})
				if err != nil {
					ctx.JSON(
						http.StatusInternalServerError,
						gin.H{"error": authErrors.ErrFailToParseNewAccessToken},
					)
					ctx.Abort()
					return
				}
			} else {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				ctx.Abort()
				return
			}
		}

		ctx.Set("userID", claims.UserID)
		ctx.Next()
	}
}

func GzipDecompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		encoding := c.GetHeader("Content-Encoding")
		if strings.Contains(encoding, "gzip") {
			gr, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "invalid gzip body",
				})
				return
			}
			defer func(gr *gzip.Reader) {
				err = gr.Close()
				if err != nil {
					log.Printf("failed to close gzip body: %v", err)
				}
			}(gr)
			c.Request.Body = io.NopCloser(gr)
		}
		c.Next()
	}
}
