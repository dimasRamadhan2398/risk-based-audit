package services

import (
	departmentRepo "master-service/repositories/department"
	departmentSvc "master-service/services/department"

	"gorm.io/gorm"
)

type IServiceRegistry interface {
	GetDepartment() departmentSvc.DepartmentServiceInterface
}

type Registry struct {
	department departmentSvc.DepartmentServiceInterface
}

func NewServiceRegistry(db *gorm.DB) IServiceRegistry {
	return &Registry{
		department: departmentSvc.NewDepartmentService(departmentRepo.NewDepartmentRepository(db)),
	}
}

func (r *Registry) GetDepartment() departmentSvc.DepartmentServiceInterface {
	return r.department
}
