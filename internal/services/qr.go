package services

import (
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
)

// QRService provides QR code generation functionality
type QRService struct {
	logger *logrus.Logger
}

// NewQRService creates a new QR code service
func NewQRService(logger *logrus.Logger) *QRService {
	return &QRService{
		logger: logger,
	}
}

// GenerateQR generates a QR code for the given text
func (s *QRService) GenerateQR(text string) ([]byte, error) {
	s.logger.Debugf("Generating QR code for text: %s", text)
	
	// Generate QR code with medium recovery level and size 256
	qr, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		s.logger.Errorf("Failed to generate QR code: %v", err)
		return nil, err
	}
	
	return qr, nil
}