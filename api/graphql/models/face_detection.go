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
	minX, maxX float32
	minY, maxY float32
}

// ToDBFaceRectangle converts a pixel absolute rectangle to a relative FaceRectangle to be saved in the database
func ToDBFaceRectangle(imgRec image.Rectangle, imagePath string) (*FaceRectangle, error) {
	size, err := image_helpers.GetPhotoDimensions(imagePath)
	if err != nil {
		return nil, err
	}

	return &FaceRectangle{
		minX: float32(imgRec.Min.X) / float32(size.Width),
		maxX: float32(imgRec.Max.X) / float32(size.Width),
		minY: float32(imgRec.Min.Y) / float32(size.Height),
		maxY: float32(imgRec.Max.Y) / float32(size.Height),
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

	fr.minX = float32(minX)
	fr.minX = float32(maxX)
	fr.minX = float32(minY)
	fr.minX = float32(maxY)

	return nil
}

// Value tells GORM how to save into the database
func (fr FaceRectangle) Value() (driver.Value, error) {
	result := fmt.Sprintf("%f:%f:%f:%f", fr.minX, fr.maxX, fr.minY, fr.maxY)
	return result, nil
}
