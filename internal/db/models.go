package db

import (
	"time"

	"gorm.io/gorm"
)

// Server represents a target server managed by the panel.
type Server struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"uniqueIndex;size:255"`
	Host           string         `json:"host" gorm:"size:255"`
	SSHPort        int            `json:"ssh_port" gorm:"default:22"`
	AuthType       string         `json:"auth_type" gorm:"default:password;size:50"`
	AuthCredential string         `json:"-" gorm:"size:8192"`
	GPTPort        int            `json:"gpt_port" gorm:"default:3001"`
	GPTMode        string         `json:"gpt_mode" gorm:"default:standalone;size:50"`
	Status         string         `json:"status" gorm:"default:unknown;size:50"`
	Version        string         `json:"version" gorm:"size:100"`
	LastHealthAt   *time.Time     `json:"last_health_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// DeployLog records a deployment action on a server.
type DeployLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ServerID  uint           `json:"server_id"`
	Server    Server         `json:"server,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"`
	Action    string         `json:"action" gorm:"size:50"`       // install | upgrade | restart | rollback
	GPTVersion string        `json:"gpt_version" gorm:"size:50"`
	Status    string         `json:"status" gorm:"default:pending;size:50"` // pending | running | success | failed
	Log       string         `json:"log" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
