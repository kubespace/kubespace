package audit

import "gorm.io/gorm"

type AuditOperationManager struct {
	*gorm.DB
}
