package command

import (
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/icons/domain"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/storage"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"image/png"
	"path/filepath"
)

type CreateIconRequest struct {
	Id          uuid.UUID
	CallerId    string   `json:"caller_id" validate:"required,max=128"`
	ServiceName string   `json:"service_name" validate:"required,max=128"`
	Issuers     []string `json:"issuers" validate:"required,max=128"`
	Description string   `json:"description" validate:"omitempty,max=512"`
	LightIcon   string   `json:"light_icon" validate:"required,base64"`
	DarkIcon    string   `json:"dark_icon" validate:"omitempty,base64"`
}

type CreateIconRequestHandler struct {
	Storage    storage.FileSystemStorage
	Repository domain.IconsRequestsRepository
}

func (h *CreateIconRequestHandler) Handle(cmd *CreateIconRequest) error {
	_, rawImg, err := processB64PngImage(cmd.LightIcon)

	if err != nil {
		return err
	}

	lightIconPath := filepath.Join(iconsStoragePath, "ir_"+uuid.New().String()+".light.png")

	lightIconLocation, err := h.Storage.Save(lightIconPath, rawImg)

	if err != nil {
		return err
	}

	issuers, err := json.Marshal(cmd.Issuers)

	if err != nil {
		return err
	}

	iconRequest := &domain.IconRequest{
		Id:           cmd.Id,
		CallerId:     cmd.CallerId,
		Issuers:      issuers,
		ServiceName:  cmd.ServiceName,
		Description:  cmd.Description,
		LightIconUrl: lightIconLocation,
	}

	if cmd.DarkIcon != "" {
		_, darkIconRaw, err := processB64PngImage(cmd.DarkIcon)

		if err != nil {
			return err
		}

		darkIconPath := filepath.Join(iconsStoragePath, "ir_"+uuid.New().String()+".dark.png")

		darkIconLocation, err := h.Storage.Save(darkIconPath, darkIconRaw)

		if err != nil {
			return err
		}

		iconRequest.DarkIconUrl = darkIconLocation
	}

	return h.Repository.Save(iconRequest)
}

type DeleteIconRequest struct {
	Id string `uri:"icon_request_id" validate:"required,uuid4"`
}

type DeleteIconRequestHandler struct {
	Repository domain.IconsRequestsRepository
}

func (h *DeleteIconRequestHandler) Handle(cmd *DeleteIconRequest) error {
	id, _ := uuid.Parse(cmd.Id)

	icon, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	return h.Repository.Delete(icon)
}

type DeleteAllIconsRequestsHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeleteAllIconsRequestsHandler) Handle() {
	sql, _, _ := h.Qb.Truncate("icons_requests").ToSQL()

	h.Database.Exec(sql)
}

type UpdateWebServiceFromIconRequest struct {
	IconRequestId string `uri:"icon_request_id" validate:"required,uuid4"`
	WebServiceId  string `json:"web_service_id" validate:"required,uuid4"`
}

type UpdateWebServiceFromIconRequestHandler struct {
	IconsStorage               storage.FileSystemStorage
	WebServiceRepository       domain.WebServicesRepository
	IconsCollectionsRepository domain.IconsCollectionRepository
	IconsRepository            domain.IconsRepository
	IconsRequestsRepository    domain.IconsRequestsRepository
}

func (h *UpdateWebServiceFromIconRequestHandler) Handle(cmd *UpdateWebServiceFromIconRequest) error {
	webServiceId, err := uuid.Parse(cmd.WebServiceId)

	if err != nil {
		return err
	}

	iconRequestId, err := uuid.Parse(cmd.IconRequestId)

	if err != nil {
		return err
	}

	iconRequest, err := h.IconsRequestsRepository.FindById(iconRequestId)

	if err != nil {
		return err
	}

	webService, err := h.WebServiceRepository.FindById(webServiceId)

	if err != nil {
		return err
	}

	iconsCollectionId := uuid.New()

	lightIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.LightIconUrl))

	lightIconImg, err := h.IconsStorage.Get(lightIconStoragePath)

	if err != nil {
		return err
	}

	lightIconPng, err := png.Decode(lightIconImg)

	if err != nil {
		return err
	}

	lightIconId := uuid.New()
	lightIconNewPath := filepath.Join(iconsStoragePath, lightIconId.String()+".png")
	newLightIconLocation, err := h.IconsStorage.Move(lightIconStoragePath, lightIconNewPath)

	if err != nil {
		return err
	}

	lightIcon := &domain.Icon{
		Id:     lightIconId,
		Name:   iconRequest.ServiceName,
		Url:    newLightIconLocation,
		Width:  lightIconPng.Bounds().Dx(),
		Height: lightIconPng.Bounds().Dy(),
		Type:   domain.Light,
	}

	err = h.IconsRepository.Save(lightIcon)

	if err != nil {
		return err
	}

	iconsIds := []string{
		lightIcon.Id.String(),
	}

	if iconRequest.DarkIconUrl != "" {
		darkIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.DarkIconUrl))

		darkIconImg, err := h.IconsStorage.Get(darkIconStoragePath)

		if err != nil {
			return err
		}

		darkIconPng, err := png.Decode(darkIconImg)

		if err != nil {
			return err
		}

		darkIconId := uuid.New()
		darkIconNewPath := filepath.Join(iconsStoragePath, darkIconId.String()+".png")
		newDarkIconLocation, err := h.IconsStorage.Move(darkIconStoragePath, darkIconNewPath)

		if err != nil {
			return err
		}

		darkIcon := &domain.Icon{
			Id:     darkIconId,
			Name:   iconRequest.ServiceName,
			Url:    newDarkIconLocation,
			Width:  darkIconPng.Bounds().Dx(),
			Height: darkIconPng.Bounds().Dy(),
			Type:   domain.Dark,
		}

		err = h.IconsRepository.Save(darkIcon)

		if err != nil {
			return err
		}

		iconsIds = append(iconsIds, darkIconId.String())
	}

	iconsJson, err := json.Marshal(iconsIds)

	iconsCollection := &domain.IconsCollection{
		Id:    iconsCollectionId,
		Name:  iconRequest.ServiceName,
		Icons: iconsJson,
	}

	err = h.IconsCollectionsRepository.Save(iconsCollection)

	if err != nil {
		return err
	}

	var webServiceIconsCollectionsIds []string

	err = json.Unmarshal(webService.IconsCollections, &webServiceIconsCollectionsIds)

	if err != nil {
		return err
	}

	for _, outdatedIconsCollectionId := range webServiceIconsCollectionsIds {
		id, _ := uuid.Parse(outdatedIconsCollectionId)

		outDatedIconsCollection, err := h.IconsCollectionsRepository.FindById(id)

		if err != nil {
			logging.
				WithField("icon_collection_id", outdatedIconsCollectionId).
				Error("Out of date icons collection cannot be found")
		}

		err = h.IconsCollectionsRepository.Delete(outDatedIconsCollection)

		if err != nil {
			logging.
				WithField("icon_collection_id", outdatedIconsCollectionId).
				Error("Cannot delete out of date icons collection")
		}

		var outdatedCollectionIcons []string

		err = json.Unmarshal(outDatedIconsCollection.Icons, &outdatedCollectionIcons)

		if err != nil {
			return err
		}

		for _, outdatedIconId := range webServiceIconsCollectionsIds {
			iconId, _ := uuid.Parse(outdatedIconId)
			iconToDelete, err := h.IconsRepository.FindById(iconId)

			if err == nil {
				h.IconsRepository.Delete(iconToDelete)
			}
		}
	}

	webService.IconsCollections = datatypes.JSON(`["` + iconsCollection.Id.String() + `"]`)

	h.WebServiceRepository.Update(webService)

	err = h.IconsRequestsRepository.Delete(iconRequest)

	if err != nil {
		return err
	}

	return nil
}

type TransformIconRequestToWebService struct {
	WebServiceId  uuid.UUID
	IconRequestId string `uri:"icon_request_id" validate:"required,uuid4"`
}

type TransformIconRequestToWebServiceHandler struct {
	IconsStorage               storage.FileSystemStorage
	WebServiceRepository       domain.WebServicesRepository
	IconsRepository            domain.IconsRepository
	IconsCollectionsRepository domain.IconsCollectionRepository
	IconsRequestsRepository    domain.IconsRequestsRepository
}

func (h *TransformIconRequestToWebServiceHandler) Handle(cmd *TransformIconRequestToWebService) error {
	iconRequestId, err := uuid.Parse(cmd.IconRequestId)

	if err != nil {
		return err
	}

	iconRequest, err := h.IconsRequestsRepository.FindById(iconRequestId)

	if err != nil {
		return err
	}

	conflict, err := h.WebServiceRepository.FindByName(iconRequest.ServiceName)

	if conflict != nil {
		return domain.WebServiceAlreadyExistsError{Name: iconRequest.ServiceName}
	}

	iconsCollectionId := uuid.New()

	lightIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.LightIconUrl))

	lightIconImg, err := h.IconsStorage.Get(lightIconStoragePath)

	if err != nil {
		return err
	}

	lightIconPng, err := png.Decode(lightIconImg)

	if err != nil {
		return err
	}

	lightIconId := uuid.New()
	lightIconNewPath := filepath.Join(iconsStoragePath, lightIconId.String()+".png")
	newLightIconLocation, err := h.IconsStorage.Move(lightIconStoragePath, lightIconNewPath)

	if err != nil {
		return err
	}

	lightIcon := &domain.Icon{
		Id:     lightIconId,
		Name:   iconRequest.ServiceName,
		Url:    newLightIconLocation,
		Width:  lightIconPng.Bounds().Dx(),
		Height: lightIconPng.Bounds().Dy(),
		Type:   domain.Light,
	}

	err = h.IconsRepository.Save(lightIcon)

	if err != nil {
		return err
	}

	iconsIds := []string{
		lightIcon.Id.String(),
	}

	if iconRequest.DarkIconUrl != "" {
		darkIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.DarkIconUrl))

		darkIconImg, err := h.IconsStorage.Get(darkIconStoragePath)

		if err != nil {
			return err
		}

		darkIconPng, err := png.Decode(darkIconImg)

		if err != nil {
			return err
		}

		darkIconId := uuid.New()
		darkIconNewPath := filepath.Join(iconsStoragePath, darkIconId.String()+".png")
		newDarkIconLocation, err := h.IconsStorage.Move(darkIconStoragePath, darkIconNewPath)

		if err != nil {
			return err
		}

		darkIcon := &domain.Icon{
			Id:     darkIconId,
			Name:   iconRequest.ServiceName,
			Url:    newDarkIconLocation,
			Width:  darkIconPng.Bounds().Dx(),
			Height: darkIconPng.Bounds().Dy(),
			Type:   domain.Dark,
		}

		err = h.IconsRepository.Save(darkIcon)

		if err != nil {
			return err
		}

		iconsIds = append(iconsIds, darkIconId.String())
	}

	iconsJson, err := json.Marshal(iconsIds)

	iconsCollection := &domain.IconsCollection{
		Id:    iconsCollectionId,
		Name:  iconRequest.ServiceName,
		Icons: iconsJson,
	}

	err = h.IconsCollectionsRepository.Save(iconsCollection)

	if err != nil {
		return err
	}

	webService := &domain.WebService{
		Id:               cmd.WebServiceId,
		Name:             iconRequest.ServiceName,
		Issuers:          iconRequest.Issuers,
		Tags:             datatypes.JSON(`[]`),
		IconsCollections: datatypes.JSON(`["` + iconsCollectionId.String() + `"]`),
		MatchRules:       nil,
	}

	err = h.WebServiceRepository.Save(webService)

	if err != nil {
		return err
	}

	err = h.IconsRequestsRepository.Delete(iconRequest)

	if err != nil {
		return err
	}

	return nil
}
