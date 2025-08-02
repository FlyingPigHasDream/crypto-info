package services

import (
	"go-web-study/internal/dao"
	"go-web-study/internal/models"
)

// HealthService 健康检查服务
type HealthService struct {
	healthDAO dao.HealthDAO
}

// NewHealthService 创建新的健康检查服务实例
func NewHealthService() *HealthService {
	return &HealthService{
		healthDAO: dao.NewHealthDAO(),
	}
}

// GetSystemHealth 获取系统健康状态
func (s *HealthService) GetSystemHealth() models.Response {
	return *s.healthDAO.CheckSystemHealth()
}

// GetDetailedHealth 获取详细的健康检查信息
func (s *HealthService) GetDetailedHealth() map[string]interface{} {
	return map[string]interface{}{
		"system":      s.healthDAO.CheckSystemHealth(),
		"database":    s.healthDAO.CheckDatabaseHealth(),
		"external_api": s.healthDAO.CheckExternalAPIHealth(),
	}
}