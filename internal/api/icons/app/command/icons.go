package command

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/twofas/2fas-server/internal/api/icons/domain"
	"github.com/twofas/2fas-server/internal/common/storage"
	"gorm.io/gorm"
	"image"
	"image/png"
	"io"
	"path/filepath"
)

var iconsStoragePath = "2fas-icons"

func processB64PngImage(b64Img string) (image.Image, io.Reader, error) {
	img, err := base64.StdEncoding.DecodeString(b64Img)

	if err != nil {
		return nil, nil, errors.New("Cannot decode b64")
	}

	imageBytes := bytes.NewReader(img)

	if err != nil {
		return nil, nil, err
	}

	pngImg, err := png.Decode(imageBytes)

	if err != nil {
		return nil, nil, err
	}

	imageBytes.Seek(0, io.SeekStart)

	err = validateImage(pngImg)

	if err != nil {
		return nil, nil, err
	}

	return pngImg, imageBytes, nil
}

func validateImage(img image.Image) error {
	imageWidth := img.Bounds().Dx()
	imageHeight := img.Bounds().Dy()

	validDimensions := []*image.Point{
		{X: 120, Y: 120},
		{X: 80, Y: 80},
		{X: 40, Y: 40},
	}

	for _, dimension := range validDimensions {
		if imageWidth == dimension.X && imageHeight == dimension.Y {
			return nil
		}
	}

	errMsg := fmt.Sprintf("Invalid image dimensions [%d %d]: allowed options are 120x120, 80x80, 40x40", imageWidth, imageHeight)

	return errors.New(errMsg)
}

// CreateIcon

type CreateIcon struct {
	Id   uuid.UUID
	Name string `json:"name" validate:"required,max=128"`
	Icon string `json:"icon" validate:"required,base64"`
	Type string `json:"type" validate:"required,oneof=light dark"`
}

type CreateIconHandler struct {
	Repository domain.IconsRepository
	Storage    storage.FileSystemStorage
}

func (h *CreateIconHandler) Handle(cmd *CreateIcon) error {
	pngImg, rawImg, err := processB64PngImage(cmd.Icon)

	if err != nil {
		return err
	}

	iconFilename := cmd.Id.String() + ".png"

	storagePath := filepath.Join(iconsStoragePath, iconFilename)
	location, err := h.Storage.Save(storagePath, rawImg)

	if err != nil {
		return err
	}

	icon := &domain.Icon{
		Id:     cmd.Id,
		Name:   cmd.Name,
		Url:    location,
		Width:  pngImg.Bounds().Dx(),
		Height: pngImg.Bounds().Dy(),
		Type:   cmd.Type,
	}

	return h.Repository.Save(icon)
}

// UpdateIcon

type UpdateIcon struct {
	Id   string `uri:"icon_id" validate:"required,uuid4"`
	Name string `json:"name" validate:"omitempty,max=128"`
	Icon string `json:"icon" validate:"omitempty,base64"`
	Type string `json:"type" validate:"omitempty,oneof=light dark"`
}

type UpdateIconHandler struct {
	Repository domain.IconsRepository
	Storage    storage.FileSystemStorage
}

func (h *UpdateIconHandler) Handle(cmd *UpdateIcon) error {
	id, _ := uuid.Parse(cmd.Id)

	icon, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	if cmd.Icon != "" {
		pngImg, rawImg, err := processB64PngImage(cmd.Icon)

		if err != nil {
			return err
		}

		storagePath := filepath.Join(iconsStoragePath, cmd.Id+".png")

		location, err := h.Storage.Save(storagePath, rawImg)

		if err != nil {
			return err
		}

		icon.Url = location
		icon.Width = pngImg.Bounds().Dx()
		icon.Height = pngImg.Bounds().Dy()
	}

	if cmd.Name != "" {
		icon.Name = cmd.Name
	}

	if cmd.Type != "" {
		icon.Type = cmd.Type
	}

	return h.Repository.Update(icon)
}

// DeleteIcon

type DeleteIcon struct {
	Id string `uri:"icon_id" validate:"required,uuid4"`
}

type DeleteIconHandler struct {
	Repository              domain.IconsRepository
	IconsRelationRepository domain.IconsRelationsRepository
}

func (h *DeleteIconHandler) Handle(cmd *DeleteIcon) error {
	id, _ := uuid.Parse(cmd.Id)

	icon, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	err = h.Repository.Delete(icon)

	if err != nil {
		return err
	}

	return h.IconsRelationRepository.DeleteAll(icon)
}

type DeleteAllIcons struct{}

type DeleteAllIconsHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeleteAllIconsHandler) Handle(cmd *DeleteAllIcons) {
	sql, _, _ := h.Qb.Truncate("icons").ToSQL()

	h.Database.Exec(sql)
}
