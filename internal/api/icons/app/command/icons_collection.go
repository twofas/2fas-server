package command

import (
	"encoding/json"
	domain2 "github.com/2fas/api/internal/api/icons/domain"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreateIconCollection

type CreateIconsCollection struct {
	Id          uuid.UUID
	Name        string         `json:"name" validate:"required,max=128"`
	Description string         `json:"description" validate:"omitempty,max=512"`
	Icons       datatypes.JSON `json:"icons"`
}

type CreateIconsCollectionHandler struct {
	Repository domain2.IconsCollectionRepository
}

func (h *CreateIconsCollectionHandler) Handle(cmd *CreateIconsCollection) error {
	collection := &domain2.IconsCollection{
		Id:          cmd.Id,
		Name:        cmd.Name,
		Description: cmd.Description,
		Icons:       cmd.Icons,
	}

	return h.Repository.Save(collection)
}

// UpdateIconsCollection

type UpdateIconsCollection struct {
	Id          string   `uri:"collection_id" validate:"required,uuid4"`
	Name        string   `json:"name"`
	Description string   `json:"description" validate:"omitempty,max=512"`
	Icons       []string `json:"icons"`
}

type UpdateIconsCollectionHandler struct {
	Repository domain2.IconsCollectionRepository
}

func (h *UpdateIconsCollectionHandler) Handle(cmd *UpdateIconsCollection) error {
	id, _ := uuid.Parse(cmd.Id)

	collection, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	if cmd.Name != "" {
		collection.Name = cmd.Name
	}

	if cmd.Description != "" {
		collection.Description = cmd.Description
	}

	if cmd.Icons != nil {
		icons, err := json.Marshal(cmd.Icons)

		if err != nil {
			return err
		}

		collection.Icons = icons
	}

	return h.Repository.Update(collection)
}

// DeleteIcon

type DeleteIconsCollection struct {
	Id string `uri:"collection_id" validate:"required,uuid4"`
}

type DeleteIconsCollectionHandler struct {
	Repository                          domain2.IconsCollectionRepository
	IconsCollectionsRelationsRepository domain2.IconsCollectionsRelationsRepository
}

func (h *DeleteIconsCollectionHandler) Handle(cmd *DeleteIconsCollection) error {
	id, _ := uuid.Parse(cmd.Id)

	collection, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	err = h.Repository.Delete(collection)

	if err != nil {
		return err
	}

	return h.IconsCollectionsRelationsRepository.DeleteAll(collection)
}

// DeleteAllIconsCollections

type DeleteAllIconsCollections struct{}

type DeleteAllIconsCollectionsHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeleteAllIconsCollectionsHandler) Handle(cmd *DeleteAllIconsCollections) {
	sql, _, _ := h.Qb.Truncate("icons_collections").ToSQL()

	h.Database.Exec(sql)
}
