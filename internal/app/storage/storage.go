package storage

type URLRepository interface {
	Save(hash string, original string) error
	Get(hash string) (string, error)
}
