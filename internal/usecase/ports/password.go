package ports

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) bool
}
