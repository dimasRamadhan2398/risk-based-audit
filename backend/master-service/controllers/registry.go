package controllers

import (
	companyCtrl "master-service/controllers/company"
	departmentCtrl "master-service/controllers/department"
	employeeCtrl "master-service/controllers/employee"
	"master-service/pkg/validations"
	"master-service/services"
)

type Registry struct {
	service   services.IServiceRegistry
	validator *validations.Validator
}

type IControllerRegistry interface {
	GetCompany() companyCtrl.CompanyControllerInterface
	GetDepartment() departmentCtrl.DepartmentControllerInterface
	GetEmployee() employeeCtrl.EmployeeControllerInterface
}

func NewControllerRegistry(service services.IServiceRegistry, validator *validations.Validator) IControllerRegistry {
	return &Registry{
		service:   service,
		validator: validator,
	}
}

func (r *Registry) GetCompany() companyCtrl.CompanyControllerInterface {
	return companyCtrl.NewCompanyController(r.service.GetCompany(), r.validator)
}

func (r *Registry) GetDepartment() departmentCtrl.DepartmentControllerInterface {
	return departmentCtrl.NewDepartmentController(r.service.GetDepartment(), r.validator)
}

func (r *Registry) GetEmployee() employeeCtrl.EmployeeControllerInterface {
	return employeeCtrl.NewEmployeeController(r.service.GetEmployee(), r.validator)
}
