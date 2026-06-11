package controllers

import (
	departmentCtrl "master-service/controllers/department"
	"master-service/pkg/validations"
	"master-service/services"
)

type Registry struct {
	service   services.IServiceRegistry
	validator *validations.Validator
}

type IControllerRegistry interface {
	GetDepartment() departmentCtrl.DepartmentControllerInterface
}

func NewControllerRegistry(service services.IServiceRegistry, validator *validations.Validator) IControllerRegistry {
	return &Registry{
		service:   service,
		validator: validator,
	}
}

func (r *Registry) GetDepartment() departmentCtrl.DepartmentControllerInterface {
	return departmentCtrl.NewDepartmentController(r.service.GetDepartment(), r.validator)
}
