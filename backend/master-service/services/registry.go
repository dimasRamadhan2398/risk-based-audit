package services

import (
	companyRepo "master-service/repositories/company"
	companySvc "master-service/services/company"
	departmentRepo "master-service/repositories/department"
	departmentSvc "master-service/services/department"
	employeeRepo "master-service/repositories/employee"
	employeeSvc "master-service/services/employee"

	"gorm.io/gorm"
)

type IServiceRegistry interface {
	GetCompany() companySvc.CompanyServiceInterface
	GetDepartment() departmentSvc.DepartmentServiceInterface
	GetEmployee() employeeSvc.EmployeeServiceInterface
}

type Registry struct {
	company   companySvc.CompanyServiceInterface
	department departmentSvc.DepartmentServiceInterface
	employee   employeeSvc.EmployeeServiceInterface
}

func NewServiceRegistry(db *gorm.DB) IServiceRegistry {
	return &Registry{
		company:   companySvc.NewCompanyService(companyRepo.NewCompanyRepository(db)),
		department: departmentSvc.NewDepartmentService(departmentRepo.NewDepartmentRepository(db)),
		employee:   employeeSvc.NewEmployeeService(employeeRepo.NewEmployeeRepository(db)),
	}
}

func (r *Registry) GetCompany() companySvc.CompanyServiceInterface {
	return r.company
}

func (r *Registry) GetDepartment() departmentSvc.DepartmentServiceInterface {
	return r.department
}

func (r *Registry) GetEmployee() employeeSvc.EmployeeServiceInterface {
	return r.employee
}
