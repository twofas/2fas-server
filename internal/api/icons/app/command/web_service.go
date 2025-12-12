package command

import (
	"encoding/json"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/adapters"
	"github.com/twofas/2fas-server/internal/api/icons/domain"
)

type MatchRule struct {
	Field      string `json:"field" validate:"oneof=label issuer account"`
	Text       string `json:"text" validate:"max=64"`
	Matcher    string `json:"matcher" validate:"oneof=contains starts_with ends_with equals regex"`
	IgnoreCase bool   `json:"ignore_case"`
}

type CreateWebService struct {
	Id               uuid.UUID
	Name             string       `json:"name" validate:"required,max=64"`
	Description      string       `json:"description" validate:"omitempty,max=512"`
	Issuers          []string     `json:"issuers" validate:"required,max=128"`
	IconsCollections []string     `json:"icons_collections" validate:"required"`
	MatchRules       []*MatchRule `json:"match_rules"`
	Tags             []string     `json:"tags"`
}

type CreateWebServiceHandler struct {
	Repository domain.WebServicesRepository
}

func (h *CreateWebServiceHandler) Handle(cmd *CreateWebService) error {
	issuers, err := json.Marshal(cmd.Issuers)
	if err != nil {
		return err
	}

	tags, err := json.Marshal(cmd.Tags)
	if err != nil {
		return err
	}

	iconsCollections, err := json.Marshal(cmd.IconsCollections)
	if err != nil {
		return err
	}

	matchRules, err := json.Marshal(cmd.MatchRules)
	if err != nil {
		return err
	}

	conflict, err := h.Repository.FindByName(cmd.Name)
	if err != nil {
		var notFound adapters.WebServiceCouldNotBeFoundError
		if !errors.As(err, &notFound) {
			return err
		}
	}
	if conflict != nil {
		return domain.WebServiceAlreadyExistsError{Name: cmd.Name}
	}

	webService := &domain.WebService{
		Id:               cmd.Id,
		Name:             cmd.Name,
		Description:      cmd.Description,
		Issuers:          issuers,
		Tags:             tags,
		IconsCollections: iconsCollections,
		MatchRules:       matchRules,
	}

	return h.Repository.Save(webService)
}

type UpdateWebService struct {
	Id               string       `uri:"service_id" validate:"required,uuid4"`
	Name             string       `json:"name" validate:"omitempty,max=64"`
	Description      string       `json:"description" validate:"omitempty,max=512"`
	Issuers          []string     `json:"issuers"`
	Tags             []string     `json:"tags"`
	IconsCollections []string     `json:"icons_collections"`
	MatchRules       []*MatchRule `json:"match_rules"`
}

type UpdateWebServiceHandler struct {
	Repository domain.WebServicesRepository
}

func (h *UpdateWebServiceHandler) Handle(cmd *UpdateWebService) error {
	id, _ := uuid.Parse(cmd.Id)

	webService, err := h.Repository.FindById(id)
	if err != nil {
		return err
	}

	if cmd.Issuers != nil {
		issuers, err := json.Marshal(cmd.Issuers)
		if err != nil {
			return err
		}

		webService.Issuers = issuers
	}

	if cmd.Tags != nil {
		tags, err := json.Marshal(cmd.Tags)
		if err != nil {
			return err
		}

		webService.Tags = tags
	}

	if cmd.IconsCollections != nil {
		iconsCollections, err := json.Marshal(cmd.IconsCollections)
		if err != nil {
			return err
		}

		webService.IconsCollections = iconsCollections
	}

	if cmd.MatchRules != nil {
		matchRules, err := json.Marshal(cmd.MatchRules)
		if err != nil {
			return err
		}

		webService.MatchRules = matchRules
	}

	if cmd.Name != "" {
		webService.Name = cmd.Name
	}

	if cmd.Description != "" {
		webService.Description = cmd.Description
	}

	return h.Repository.Update(webService)
}

type DeleteWebService struct {
	Id string `uri:"service_id" validate:"required,uuid4"`
}

type DeleteWebServiceHandler struct {
	Repository domain.WebServicesRepository
}

func (h *DeleteWebServiceHandler) Handle(cmd *DeleteWebService) error {
	id, _ := uuid.Parse(cmd.Id)

	webService, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	return h.Repository.Delete(webService)
}

type DeleteAllWebServices struct{}

type DeleteAllWebServicesHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeleteAllWebServicesHandler) Handle(cmd *DeleteAllWebServices) {
	sql, _, _ := h.Qb.Truncate("web_services").ToSQL()

	h.Database.Exec(sql)
}
