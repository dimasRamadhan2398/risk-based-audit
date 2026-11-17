package masterclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// Lookups used by the COSO internal control assessment (services/ic_assessment).
// Nothing here is cached: an assessment proposal must see current data.

// DepartmentDetail is GET /api/v1/departments/:id
type DepartmentDetail struct {
	ID             string  `json:"id"`
	DepartmentCode string  `json:"department_code"`
	DepartmentName string  `json:"department_name"`
	CompanyID      string  `json:"company_id"`
	PicID          *string `json:"pic_id"`
	IsActive       bool    `json:"is_active"`
}

// HasPIC reports whether the department has a person in charge
func (d *DepartmentDetail) HasPIC() bool {
	if d == nil || d.PicID == nil {
		return false
	}
	id, err := uuid.Parse(strings.TrimSpace(*d.PicID))
	return err == nil && id != uuid.Nil
}

// Location is GET /api/v1/locations/:id
type Location struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	City     string `json:"city"`
	IsActive bool   `json:"is_active"`
}

// Employee is one row of GET /api/v1/employees
type Employee struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	DepartmentID string `json:"department_id"`
	IsActive     bool   `json:"is_active"`
}

// GetDepartment fetches one department. ErrNotFound when master-service
// answers without a usable department.
func (c *MasterClient) GetDepartment(ctx context.Context, authorization, id string) (*DepartmentDetail, error) {
	data, err := c.getData(ctx, authorization, "/api/v1/departments/"+url.PathEscape(strings.TrimSpace(id)), nil)
	if err != nil {
		return nil, err
	}
	var d DepartmentDetail
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("decode department: %w", err)
	}
	if strings.TrimSpace(d.ID) == "" {
		return nil, ErrNotFound
	}
	return &d, nil
}

// GetLocation fetches one location (branch)
func (c *MasterClient) GetLocation(ctx context.Context, authorization, id string) (*Location, error) {
	data, err := c.getData(ctx, authorization, "/api/v1/locations/"+url.PathEscape(strings.TrimSpace(id)), nil)
	if err != nil {
		return nil, err
	}
	var l Location
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("decode location: %w", err)
	}
	if strings.TrimSpace(l.ID) == "" {
		return nil, ErrNotFound
	}
	return &l, nil
}

// GetEmployeeByCode fetches GET /api/v1/employees/by-code/:code
func (c *MasterClient) GetEmployeeByCode(ctx context.Context, authorization, code string) (*Employee, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrNotFound
	}
	data, err := c.getData(ctx, authorization, "/api/v1/employees/by-code/"+url.PathEscape(code), nil)
	if err != nil {
		return nil, err
	}
	var e Employee
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("decode employee: %w", err)
	}
	if strings.TrimSpace(e.ID) == "" {
		return nil, ErrNotFound
	}
	return &e, nil
}

// employeePageSize is master-service's maximum page_size for GET /employees
const employeePageSize = 100

// maxEmployeePages bounds the pagination (100 x 200 = 20k employees)
const maxEmployeePages = 200

// ListActiveDepartmentEmployees returns the active employees of one
// department (GET /api/v1/employees?department_id=&is_active=true, all
// pages). The result is filtered here as well, so a master-service that does
// not know the filters yet (it would answer with every employee) still gives
// the right set.
func (c *MasterClient) ListActiveDepartmentEmployees(ctx context.Context, authorization, departmentID string) ([]Employee, error) {
	departmentID = strings.TrimSpace(departmentID)
	var out []Employee
	for page := 1; page <= maxEmployeePages; page++ {
		q := url.Values{}
		q.Set("department_id", departmentID)
		q.Set("is_active", "true")
		q.Set("page", fmt.Sprint(page))
		q.Set("page_size", fmt.Sprint(employeePageSize))
		data, err := c.getData(ctx, authorization, "/api/v1/employees", q)
		if err != nil {
			return nil, err
		}
		var wrapper struct {
			Employees  []Employee `json:"employees"`
			Pagination struct {
				TotalPages int `json:"total_pages"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return nil, fmt.Errorf("decode employees: %w", err)
		}
		for _, e := range wrapper.Employees {
			if e.IsActive && strings.EqualFold(strings.TrimSpace(e.DepartmentID), departmentID) {
				out = append(out, e)
			}
		}
		if len(wrapper.Employees) == 0 || page >= wrapper.Pagination.TotalPages {
			return out, nil
		}
	}
	return nil, fmt.Errorf("employees: more than %d pages", maxEmployeePages)
}
