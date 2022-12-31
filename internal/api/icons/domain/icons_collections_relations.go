package domain

type IconsCollectionsRelationsRepository interface {
	DeleteAll(iconsCollection *IconsCollection) error
}
