package storages

// example factory
func New(connstr string) (*Storage, error) {
	return &Storage{}, nil
}

type Storage struct {
}

func (p Storage) Close() error {
	panic("implement me")
}
