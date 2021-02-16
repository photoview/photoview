package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"image"
	"strconv"
	"strings"

	"github.com/Kagami/go-face"
	"github.com/photoview/photoview/api/scanner/image_helpers"
)

type FaceGroup struct {
	Model
	Label      *string
	ImageFaces []ImageFace `gorm:"constraint:OnDelete:CASCADE;"`
}

type ImageFace struct {
	Model
	FaceGroupID int            `gorm:"not null;index"`
	MediaID     int            `gorm:"not null;index"`
	Media       Media          `gorm:"constraint:OnDelete:CASCADE;"`
	Descriptor  FaceDescriptor `gorm:"not null"`
	Rectangle   FaceRectangle  `gorm:"not null"`
}

type FaceDescriptor face.Descriptor

// GormDataType datatype used in database
func (fd FaceDescriptor) GormDataType() string {
	return "BLOB"
}

// Scan tells GORM how to convert database data to Go format
func (fd *FaceDescriptor) Scan(value interface{}) error {
	byteValue := value.([]byte)
	reader := bytes.NewReader(byteValue)
	binary.Read(reader, binary.LittleEndian, fd)
	return nil
}

// Value tells GORM how to save into the database
func (fd FaceDescriptor) Value() (driver.Value, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, fd); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type FaceRectangle struct {
	MinX, MaxX float64
	MinY, MaxY float64
}

// ToDBFaceRectangle converts a pixel absolute rectangle to a relative FaceRectangle to be saved in the database
func ToDBFaceRectangle(imgRec image.Rectangle, imagePath string) (*FaceRectangle, error) {
	size, err := image_helpers.GetPhotoDimensions(imagePath)
	if err != nil {
		return nil, err
	}

	return &FaceRectangle{
		MinX: float64(imgRec.Min.X) / float64(size.Width),
		MaxX: float64(imgRec.Max.X) / float64(size.Width),
		MinY: float64(imgRec.Min.Y) / float64(size.Height),
		MaxY: float64(imgRec.Max.Y) / float64(size.Height),
	}, nil
}

// GormDataType datatype used in database
func (fr FaceRectangle) GormDataType() string {
	return "VARCHAR(64)"
}

// Scan tells GORM how to convert database data to Go format
func (fr *FaceRectangle) Scan(value interface{}) error {
	byteArray := value.([]uint8)
	slices := strings.Split(string(byteArray), ":")

	if len(slices) != 4 {
		return fmt.Errorf("Invalid face rectangle format, expected 4 values, got %d", len(slices))
	}

	minX, err := strconv.ParseFloat(slices[0], 32)
	if err != nil {
		return err
	}

	maxX, err := strconv.ParseFloat(slices[0], 32)
	if err != nil {
		return err
	}

	minY, err := strconv.ParseFloat(slices[0], 32)
	if err != nil {
		return err
	}

	maxY, err := strconv.ParseFloat(slices[0], 32)
	if err != nil {
		return err
	}

	fr.MinX = float64(minX)
	fr.MinX = float64(maxX)
	fr.MinX = float64(minY)
	fr.MinX = float64(maxY)

	return nil
}

// Value tells GORM how to save into the database
func (fr FaceRectangle) Value() (driver.Value, error) {
	result := fmt.Sprintf("%f:%f:%f:%f", fr.MinX, fr.MaxX, fr.MinY, fr.MaxY)
	return result, nil
}
