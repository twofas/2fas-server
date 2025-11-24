package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"path/filepath"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/adapters"
	"github.com/twofas/2fas-server/internal/api/icons/domain"
	"github.com/twofas/2fas-server/internal/common/storage"
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

	lightIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.LightIconUrl))

	lightIconImg, err := h.IconsStorage.Get(lightIconStoragePath)
	if err != nil {
		return fmt.Errorf("failed to get the icon from the storage: %w", err)
	}

	lightIconPng, err := png.Decode(lightIconImg)
	if err != nil {
		return fmt.Errorf("failed to decode the icon as pgn: %w", err)
	}

	lightIconId := uuid.New()
	lightIconNewPath := filepath.Join(iconsStoragePath, lightIconId.String()+".png")
	newLightIconLocation, err := h.IconsStorage.Move(lightIconStoragePath, lightIconNewPath)
	if err != nil {
		return fmt.Errorf("failed to move icons storage: %w", err)
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
		return fmt.Errorf("failed to save light icon: %w", err)
	}

	iconsIds := []string{
		lightIcon.Id.String(),
	}

	if iconRequest.DarkIconUrl != "" { //nolint:dupl
		darkIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.DarkIconUrl))

		darkIconImg, err := h.IconsStorage.Get(darkIconStoragePath)
		if err != nil {
			return fmt.Errorf("failed to get dark icon: %w", err)
		}

		darkIconPng, err := png.Decode(darkIconImg)
		if err != nil {
			return fmt.Errorf("failed to decode dark icon: %w", err)
		}

		darkIconId := uuid.New()
		darkIconNewPath := filepath.Join(iconsStoragePath, darkIconId.String()+".png")
		newDarkIconLocation, err := h.IconsStorage.Move(darkIconStoragePath, darkIconNewPath)
		if err != nil {
			return fmt.Errorf("failed to move dark icon: %w", err)
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
			return fmt.Errorf("failed to save dark icon: %w", err)
		}

		iconsIds = append(iconsIds, darkIconId.String())
	}

	iconsJson, err := json.Marshal(iconsIds)
	if err != nil {
		return fmt.Errorf("failed to marshal icon ids: %w", err)
	}

	var webServiceIconsCollectionsIds []string

	err = json.Unmarshal(webService.IconsCollections, &webServiceIconsCollectionsIds)
	if err != nil {
		return fmt.Errorf("failed to decode icons collection from web service: %w", err)
	}
	if len(webServiceIconsCollectionsIds) == 1 {
		if err := h.updateIconsCollection(webServiceIconsCollectionsIds[0], iconRequest.ServiceName, iconsJson); err != nil {
			return fmt.Errorf("failed to update icons collection %q: %w", webServiceIconsCollectionsIds[0], err)
		}
	} else {
		webService.IconsCollections, err = h.replaceIconsCollections(webServiceIconsCollectionsIds, iconRequest.ServiceName, iconsJson)
		if err != nil {
			return fmt.Errorf("failed to replace icons collections: %w", err)
		}
	}
	if err := h.WebServiceRepository.Update(webService); err != nil {
		return fmt.Errorf("failed to update web service %q: %w", webService.Id.String(), err)
	}

	err = h.IconsRequestsRepository.Delete(iconRequest)
	if err != nil {
		return fmt.Errorf("failed to delete icon request %q: %w", iconRequest.Id.String(), err)
	}

	return nil
}

func (h *UpdateWebServiceFromIconRequestHandler) updateIconsCollection(iconsCollectionID string, name string, iconsJson []byte) error {
	id, err := uuid.Parse(iconsCollectionID)
	if err != nil {
		return fmt.Errorf("invalid icons collection id: %w", err)
	}

	return h.IconsCollectionsRepository.Update(&domain.IconsCollection{
		Id:    id,
		Name:  name,
		Icons: iconsJson,
	})
}

func (h *UpdateWebServiceFromIconRequestHandler) replaceIconsCollections(oldCollectionIds []string, serviceName string, iconsJson []byte) (datatypes.JSON, error) {
	for _, outdatedIconsCollectionId := range oldCollectionIds {
		if err := h.deleteIconsCollection(outdatedIconsCollectionId); err != nil {
			return nil, fmt.Errorf("failed to delete icons collection %q: %w", outdatedIconsCollectionId, err)
		}
	}

	iconsCollectionId := uuid.New()
	iconsCollection := &domain.IconsCollection{
		Id:    iconsCollectionId,
		Name:  serviceName,
		Icons: iconsJson,
	}

	if err := h.IconsCollectionsRepository.Save(iconsCollection); err != nil {
		return nil, fmt.Errorf("failed to save new icons collection: %w", err)
	}

	iconsCollectionsIds := []string{iconsCollectionId.String()}
	bb, err := json.Marshal(iconsCollectionsIds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal icons collections as json: %w", err)
	}
	return bb, nil
}

func (h *UpdateWebServiceFromIconRequestHandler) deleteIconsCollection(collectionID string) error {
	id, err := uuid.Parse(collectionID)
	if err != nil {
		return fmt.Errorf("failed to parse 'collectionID' %q: %w", collectionID, err)
	}

	outDatedIconsCollection, err := h.IconsCollectionsRepository.FindById(id)
	if err != nil {
		return fmt.Errorf("failed to find out of date icons collection: %w", err)
	}
	err = h.IconsCollectionsRepository.Delete(outDatedIconsCollection)
	if err != nil {
		return fmt.Errorf("failed to delete out of date icons collection: %w", err)
	}

	var outdatedCollectionIcons []string
	err = json.Unmarshal(outDatedIconsCollection.Icons, &outdatedCollectionIcons)
	if err != nil {
		return fmt.Errorf("failed to decode icons ids: %w", err)
	}

	for _, outdatedIconId := range outdatedCollectionIcons {
		if err := h.deleteIcon(outdatedIconId); err != nil {
			return fmt.Errorf("failed to delete icon %q: %w", outdatedIconId, err)
		}
	}
	return nil
}

func (h *UpdateWebServiceFromIconRequestHandler) deleteIcon(iconIDStr string) error {
	iconID, err := uuid.Parse(iconIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse iconID: %w", err)
	}
	iconToDelete, err := h.IconsRepository.FindById(iconID)
	if err != nil {
		return fmt.Errorf("failed fetch icon by id: %w", err)
	}
	if err := h.IconsRepository.Delete(iconToDelete); err != nil {
		return fmt.Errorf("failed to delete icon: %w", err)
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
		return fmt.Errorf("invalid 'iconRequestId': %w", err)
	}

	iconRequest, err := h.IconsRequestsRepository.FindById(iconRequestId)
	if err != nil {
		return err
	}

	_, err = h.WebServiceRepository.FindByName(iconRequest.ServiceName)
	if err == nil {
		return domain.WebServiceAlreadyExistsError{Name: iconRequest.ServiceName}
	} else {
		var notFound adapters.WebServiceCouldNotBeFound
		if !errors.As(err, &notFound) {
			fmt.Printf("Error is: %T %+v\n", err, err)
			return fmt.Errorf("failed to find web service by name: %w", err)
		}
	}

	iconsCollectionId := uuid.New()

	lightIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.LightIconUrl))

	lightIconImg, err := h.IconsStorage.Get(lightIconStoragePath)
	if err != nil {
		return fmt.Errorf("failed to get light icon: %w", err)
	}

	lightIconPng, err := png.Decode(lightIconImg)
	if err != nil {
		return fmt.Errorf("failed to decode light icon: %w", err)
	}

	lightIconId := uuid.New()
	lightIconNewPath := filepath.Join(iconsStoragePath, lightIconId.String()+".png")
	newLightIconLocation, err := h.IconsStorage.Move(lightIconStoragePath, lightIconNewPath)
	if err != nil {
		return fmt.Errorf("failed to move light icon: %w", err)
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
		return fmt.Errorf("failed to save light icon: %w", err)
	}

	iconsIds := []string{
		lightIcon.Id.String(),
	}

	if iconRequest.DarkIconUrl != "" { //nolint:dupl
		darkIconStoragePath := filepath.Join(iconsStoragePath, filepath.Base(iconRequest.DarkIconUrl))

		darkIconImg, err := h.IconsStorage.Get(darkIconStoragePath)
		if err != nil {
			return fmt.Errorf("failed to get dark icon: %w", err)
		}

		darkIconPng, err := png.Decode(darkIconImg)
		if err != nil {
			return fmt.Errorf("failed to decode dark icon: %w", err)
		}

		darkIconId := uuid.New()
		darkIconNewPath := filepath.Join(iconsStoragePath, darkIconId.String()+".png")
		newDarkIconLocation, err := h.IconsStorage.Move(darkIconStoragePath, darkIconNewPath)
		if err != nil {
			return fmt.Errorf("failed to move dark icon: %w", err)
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
			return fmt.Errorf("failed to save dark icon: %w", err)
		}

		iconsIds = append(iconsIds, darkIconId.String())
	}

	iconsJson, err := json.Marshal(iconsIds)
	if err != nil {
		return fmt.Errorf("failed to encode icon ids: %w", err)
	}

	iconsCollection := &domain.IconsCollection{
		Id:    iconsCollectionId,
		Name:  iconRequest.ServiceName,
		Icons: iconsJson,
	}

	err = h.IconsCollectionsRepository.Save(iconsCollection)
	if err != nil {
		return fmt.Errorf("failed to save icons collection: %w", err)
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
		return fmt.Errorf("failed to save web service: %w", err)
	}

	err = h.IconsRequestsRepository.Delete(iconRequest)
	if err != nil {
		return fmt.Errorf("failed to delete icon request: %w", err)
	}

	return nil
}
