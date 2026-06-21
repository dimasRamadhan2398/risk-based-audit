# RBIA Master Service API Testing Guide

**Project:** Risk-Based Internal Audit (RBIA)  
**Service:** Master Service  
**Date:** June 2026  
**Version:** 1.0

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [API Base URL](#api-base-url)
4. [Department Endpoints](#department-endpoints)
5. [Employee Endpoints](#employee-endpoints)
6. [Error Responses](#error-responses)
7. [Postman Collection Import](#postman-collection-import)
8. [Testing Workflow](#testing-workflow)

---

## Overview

This document provides comprehensive testing guides for the Master Service API, specifically covering **Employee** and **Department** CRUD (Create, Read, Update, Delete) operations with pagination support.

### Base URL
```
http://localhost:8002/api/v1
```

---

## Prerequisites

Before testing the API, ensure the following:

1. ✅ **PostgreSQL is running** on port 5432
2. ✅ **Database `master_service`** exists
3. ✅ **Run database migrations:**
   ```bash
   cd backend/master-service
   go run cmd/migrate.go migrate
   ```
4. ✅ **Start the master service:**
   ```bash
   go run cmd/serve.go serve
   ```
5. ✅ **Required reference data exists:**
   - Companies (for Department & Employee)
   - Job Roles (for Employee)
   - Departments (for Employee relationships)

---

## API Base URL

| Environment | URL |
|-------------|-----|
| Local Development | `http://localhost:8002/api/v1` |
| Docker | `http://localhost:8003/api/v1` |

### Health Check Endpoint
```
GET {{baseUrl}}/../health
```
Returns:
```json
{
  "service": "master-service",
  "status": "ok"
}
```

---

## Department Endpoints

### 1. List Departments (Paginated)

**Endpoint:** `GET {{baseUrl}}/departments`

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | int | No | 1 | Page number |
| `page_size` | int | No | 10 | Items per page |
| `search` | string | No | - | Search by code, name, or description |

**Example Request:**
```
GET {{baseUrl}}/departments?page=1&page_size=10
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Departments fetched successfully",
  "data": {
    "departments": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "department_code": "DEPT001",
        "department_name": "Finance Department",
        "department_description": "Handles all financial operations",
        "level": 1,
        "is_active": true,
        "company_id": "...",
        "pic_id": "...",
        "business_unit_id": "...",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 5,
      "total_pages": 1
    }
  }
}
```

---

### 2. Get Department by ID

**Endpoint:** `GET {{baseUrl}}/departments/:id`

**Example Request:**
```
GET {{baseUrl}}/departments/123e4567-e89b-12d3-a456-426614174000
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Department fetched successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "department_code": "DEPT001",
    "department_name": "Finance Department",
    "department_description": "Handles all financial operations",
    "level": 1,
    "is_active": true,
    "company_id": "...",
    "pic_id": "...",
    "business_unit_id": "...",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 3. Create Department

**Endpoint:** `POST {{baseUrl}}/departments`

**Headers:**
```
Content-Type: application/json
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `department_code` | string | ✅ Yes | Unique department code (e.g., "DEPT001") |
| `department_name` | string | ✅ Yes | Department name |
| `pic_id` | UUID | ✅ Yes | Person in Charge ID |
| `level` | int | ✅ Yes | Department level |
| `company_id` | UUID | ✅ Yes | Company ID |
| `department_description` | string | No | Department description |
| `business_unit_id` | UUID | No | Business Unit ID |
| `is_active` | boolean | No | Active status (default: true) |

**Example Request:**
```json
POST {{baseUrl}}/departments
{
  "department_code": "DEPT001",
  "department_name": "Finance Department",
  "department_description": "Handles all financial operations",
  "pic_id": "123e4567-e89b-12d3-a456-426614174001",
  "level": 1,
  "company_id": "123e4567-e89b-12d3-a456-426614174002",
  "business_unit_id": "123e4567-e89b-12d3-a456-426614174003",
  "is_active": true
}
```

**Expected Response (201 Created):**
```json
{
  "success": true,
  "message": "Department created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "department_code": "DEPT001",
    "department_name": "Finance Department",
    "department_description": "Handles all financial operations",
    "level": 1,
    "is_active": true,
    "company_id": "...",
    "pic_id": "...",
    "business_unit_id": "...",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 4. Update Department

**Endpoint:** `PUT {{baseUrl}}/departments/:id`

**Headers:**
```
Content-Type: application/json
```

**Note:** All fields are **optional** - only include fields you want to update.

**Example Request:**
```json
PUT {{baseUrl}}/departments/123e4567-e89b-12d3-a456-426614174000
{
  "department_name": "Updated Finance Department",
  "department_description": "Updated description",
  "level": 2,
  "is_active": false
}
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Department updated successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "department_code": "DEPT001",
    "department_name": "Updated Finance Department",
    "department_description": "Updated description",
    "level": 2,
    "is_active": false,
    ...
  }
}
```

---

### 5. Delete Department

**Endpoint:** `DELETE {{baseUrl}}/departments/:id`

**Example Request:**
```
DELETE {{baseUrl}}/departments/123e4567-e89b-12d3-a456-426614174000
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Department deleted successfully",
  "data": null
}
```

---

## Employee Endpoints

### 1. List Employees (Paginated)

**Endpoint:** `GET {{baseUrl}}/employees`

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | int | No | 1 | Page number |
| `page_size` | int | No | 10 | Items per page |
| `search` | string | No | - | Search by employee_code, full_name, or email |

**Example Request:**
```
GET {{baseUrl}}/employees?page=1&page_size=10
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Employees fetched successfully",
  "data": {
    "employees": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "employee_code": "EMP001",
        "full_name": "John Doe",
        "email": "john.doe@company.com",
        "phone": "081234567890",
        "level_grade": 5,
        "is_active": true,
        "join_date": "2024-01-15T00:00:00Z",
        "company_id": "...",
        "department_id": "...",
        "job_role_id": "...",
        "work_location_id": "...",
        "manager_id": "...",
        "residence_address": "Jl. Sudirman No. 123",
        "residence_city": "Jakarta",
        "residence_province": "DKI Jakarta",
        "residence_postal_code": "12190",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 25,
      "total_pages": 3
    }
  }
}
```

---

### 2. Get Employee by ID

**Endpoint:** `GET {{baseUrl}}/employees/:id`

**Example Request:**
```
GET {{baseUrl}}/employees/123e4567-e89b-12d3-a456-426614174000
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Employee fetched successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "employee_code": "EMP001",
    "full_name": "John Doe",
    "email": "john.doe@company.com",
    "phone": "081234567890",
    "level_grade": 5,
    "is_active": true,
    "join_date": "2024-01-15T00:00:00Z",
    ...
  }
}
```

---

### 3. Create Employee

**Endpoint:** `POST {{baseUrl}}/employees`

**Headers:**
```
Content-Type: application/json
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `employee_code` | string | ✅ Yes | Unique employee code (e.g., "EMP001") |
| `full_name` | string | ✅ Yes | Employee's full name |
| `email` | string | ✅ Yes | Valid email address |
| `company_id` | UUID | ✅ Yes | Company ID |
| `department_id` | UUID | ✅ Yes | Department ID |
| `job_role_id` | UUID | ✅ Yes | Job Role ID |
| `level_grade` | int | ✅ Yes | Employee level/grade |
| `join_date` | string | ✅ Yes | Join date (format: YYYY-MM-DD) |
| `phone` | string | No | Phone number |
| `work_location_id` | UUID | No | Work Location ID |
| `manager_id` | UUID | No | Manager's Employee ID |
| `residence_address` | string | No | Home address |
| `residence_city` | string | No | City |
| `residence_province` | string | No | Province |
| `residence_postal_code` | string | No | Postal code |
| `is_active` | boolean | No | Active status (default: true) |

**Example Request:**
```json
POST {{baseUrl}}/employees
{
  "employee_code": "EMP001",
  "full_name": "John Doe",
  "email": "john.doe@company.com",
  "phone": "081234567890",
  "company_id": "123e4567-e89b-12d3-a456-426614174001",
  "department_id": "123e4567-e89b-12d3-a456-426614174002",
  "job_role_id": "123e4567-e89b-12d3-a456-426614174003",
  "level_grade": 5,
  "work_location_id": "123e4567-e89b-12d3-a456-426614174004",
  "manager_id": "123e4567-e89b-12d3-a456-426614174005",
  "residence_address": "Jl. Sudirman No. 123",
  "residence_city": "Jakarta",
  "residence_province": "DKI Jakarta",
  "residence_postal_code": "12190",
  "is_active": true,
  "join_date": "2024-01-15"
}
```

**Expected Response (201 Created):**
```json
{
  "success": true,
  "message": "Employee created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "employee_code": "EMP001",
    "full_name": "John Doe",
    "email": "john.doe@company.com",
    ...
  }
}
```

---

### 4. Update Employee

**Endpoint:** `PUT {{baseUrl}}/employees/:id`

**Headers:**
```
Content-Type: application/json
```

**Note:** All fields are **optional** - only include fields you want to update.

**Example Request:**
```json
PUT {{baseUrl}}/employees/123e4567-e89b-12d3-a456-426614174000
{
  "full_name": "Jane Doe",
  "email": "jane.doe@company.com",
  "phone": "089876543210",
  "level_grade": 6,
  "is_active": true
}
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Employee updated successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "employee_code": "EMP001",
    "full_name": "Jane Doe",
    "email": "jane.doe@company.com",
    ...
  }
}
```

---

### 5. Delete Employee

**Endpoint:** `DELETE {{baseUrl}}/employees/:id`

**Example Request:**
```
DELETE {{baseUrl}}/employees/123e4567-e89b-12d3-a456-426614174000
```

**Expected Response (200 OK):**
```json
{
  "success": true,
  "message": "Employee deleted successfully",
  "data": null
}
```

---

## Error Responses

### 400 Bad Request (Validation Error)

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Key: 'CreateEmployeeRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
    "details": ""
  }
}
```

### 404 Not Found

```json
{
  "success": false,
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "Record not found",
    "details": ""
  }
}
```

### 409 Conflict (Duplicate Code)

```json
{
  "success": false,
  "error": {
    "code": "DEPARTMENT_CODE_ALREADY_EXISTS",
    "message": "Department code already exists",
    "details": ""
  }
}
```

### 409 Conflict (Duplicate Email)

```json
{
  "success": false,
  "error": {
    "code": "EMPLOYEE_EMAIL_ALREADY_EXISTS",
    "message": "Employee email already exists",
    "details": ""
  }
}
```

---

## Postman Collection Import

### Steps to Import:

1. Open **Postman** application
2. Click **Import** button (top-left)
3. Select **File** tab
4. Drag and drop or browse to select:
   ```
   docs/Postman_Collection_RBIA_Master_Service.json
   ```
5. Click **Continue**
6. Review the collection details and click **Import**

### Collection Features:

- ✅ All CRUD endpoints for Department and Employee
- ✅ Pre-configured environment variable: `{{baseUrl}}`
- ✅ Sample request bodies
- ✅ Auto-save created IDs to collection variables
- ✅ Test scripts for validation

### Environment Variables Setup:

Create a new environment with:

| Variable | Initial Value | Current Value |
|----------|--------------|---------------|
| `baseUrl` | `http://localhost:8002/api/v1` | `http://localhost:8002/api/v1` |
| `departmentId` | (empty) | (auto-saved) |
| `employeeId` | (empty) | (auto-saved) |

---

## Testing Workflow

### Department CRUD Flow:

```
1. POST Create Department
   └── Copy the returned "id"

2. GET by ID
   └── Use the copied ID:
       GET /departments/{{departmentId}}

3. PUT Update
   └── Use the copied ID:
       PUT /departments/{{departmentId}}

4. DELETE
   └── Use the copied ID:
       DELETE /departments/{{departmentId}}
```

### Employee CRUD Flow:

```
1. POST Create Employee
   └── Copy the returned "id"

2. GET by ID
   └── Use the copied ID:
       GET /employees/{{employeeId}}

3. PUT Update
   └── Use the copied ID:
       PUT /employees/{{employeeId}}

4. DELETE
   └── Use the copied ID:
       DELETE /employees/{{employeeId}}
```

### Pagination Testing:

```
1. Create 15+ records

2. Test page 1 with page_size=5
   GET /departments?page=1&page_size=5
   └── Should return first 5 items

3. Test page 2
   GET /departments?page=2&page_size=5
   └── Should return items 6-10

4. Test search
   GET /departments?search=finance
   └── Should filter results matching "finance"
```

### Validation Testing:

```
1. Test required fields
   POST with missing required fields
   └── Should return 400 Bad Request

2. Test duplicate code
   POST with existing department_code
   └── Should return 409 Conflict

3. Test invalid UUID
   GET /departments/invalid-uuid
   └── Should return 400 Bad Request
```

---

## Additional Notes

### Soft Delete

Both Department and Employee use **soft delete** - records are marked as deleted but not physically removed from the database. They will not appear in list queries but can be found in the database if needed.

### Pagination

- Default page size: 10
- Maximum page size: 100
- Search is case-insensitive and matches partial strings

### UUID Format

All ID references must be valid UUIDs in the format:
```
123e4567-e89b-12d3-a456-426614174000
```

---

## Support

For issues or questions, please refer to:
- Project Documentation: `docs/overview.MD`
- Requirements: `docs/requirements.MD`
- Backend Code: `backend/master-service/`

---

*Document generated: June 2026*
