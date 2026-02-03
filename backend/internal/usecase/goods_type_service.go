package usecase

import (
	"context"
	"errors"
	"strings"

	"xiaoheiplay/internal/domain"
)

type GoodsTypeService struct {
	repo    GoodsTypeRepository
	plugins PluginInstallationRepository
}

func NewGoodsTypeService(repo GoodsTypeRepository, plugins PluginInstallationRepository) *GoodsTypeService {
	return &GoodsTypeService{repo: repo, plugins: plugins}
}

func (s *GoodsTypeService) List(ctx context.Context) ([]domain.GoodsType, error) {
	if s.repo == nil {
		return nil, ErrInvalidInput
	}
	return s.repo.ListGoodsTypes(ctx)
}

func (s *GoodsTypeService) Create(ctx context.Context, gt *domain.GoodsType) error {
	if s.repo == nil || gt == nil {
		return ErrInvalidInput
	}
	gt.Name = strings.TrimSpace(gt.Name)
	gt.Code = strings.TrimSpace(gt.Code)
	gt.AutomationCategory = strings.TrimSpace(gt.AutomationCategory)
	gt.AutomationPluginID = strings.TrimSpace(gt.AutomationPluginID)
	gt.AutomationInstanceID = strings.TrimSpace(gt.AutomationInstanceID)
	if gt.Name == "" {
		return ErrInvalidInput
	}
	if gt.AutomationCategory == "" {
		gt.AutomationCategory = "automation"
	}
	if gt.AutomationCategory != "automation" || gt.AutomationPluginID == "" || gt.AutomationInstanceID == "" {
		return errors.New("invalid automation binding")
	}
	if s.plugins != nil {
		if _, err := s.plugins.GetPluginInstallation(ctx, gt.AutomationCategory, gt.AutomationPluginID, gt.AutomationInstanceID); err != nil {
			return errors.New("automation plugin instance not found")
		}
	}
	return s.repo.CreateGoodsType(ctx, gt)
}

func (s *GoodsTypeService) Update(ctx context.Context, gt domain.GoodsType) error {
	if s.repo == nil {
		return ErrInvalidInput
	}
	gt.Name = strings.TrimSpace(gt.Name)
	gt.Code = strings.TrimSpace(gt.Code)
	gt.AutomationCategory = strings.TrimSpace(gt.AutomationCategory)
	gt.AutomationPluginID = strings.TrimSpace(gt.AutomationPluginID)
	gt.AutomationInstanceID = strings.TrimSpace(gt.AutomationInstanceID)
	if gt.ID <= 0 || gt.Name == "" {
		return ErrInvalidInput
	}
	if gt.AutomationCategory == "" {
		gt.AutomationCategory = "automation"
	}
	if gt.AutomationCategory != "automation" || gt.AutomationPluginID == "" || gt.AutomationInstanceID == "" {
		return errors.New("invalid automation binding")
	}
	if s.plugins != nil {
		if _, err := s.plugins.GetPluginInstallation(ctx, gt.AutomationCategory, gt.AutomationPluginID, gt.AutomationInstanceID); err != nil {
			return errors.New("automation plugin instance not found")
		}
	}
	return s.repo.UpdateGoodsType(ctx, gt)
}

func (s *GoodsTypeService) Delete(ctx context.Context, id int64) error {
	if s.repo == nil || id <= 0 {
		return ErrInvalidInput
	}
	return s.repo.DeleteGoodsType(ctx, id)
}
