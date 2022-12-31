package domain

type IconsRelationsRepository interface {
	DeleteAll(icon *Icon) error
}
