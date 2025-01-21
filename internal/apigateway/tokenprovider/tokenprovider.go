package tokenprovider

type TokenProvider interface {
	GenerateToken(data map[string]interface{}) (string, error)
	VerifyToken(token string) (map[string]interface{}, error)
	Name() string
}
