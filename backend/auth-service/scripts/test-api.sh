#!/bin/bash

# Auth Service API Testing Script
# Usage: ./test-api.sh [base_url]
# Default base_url: http://localhost:8001/api/v1

set -e

# Configuration
BASE_URL="${1:-http://localhost:8001/api/v1}"
TOKEN=""
REFRESH_TOKEN=""
USER_ID=""
DEVICE_ID=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_section() {
    echo ""
    echo "========================================"
    echo "$1"
    echo "========================================"
}

# Make a request and capture response
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    local headers="-H Content-Type: application/json"

    if [ -n "$token" ]; then
        headers="$headers -H Authorization: Bearer $token"
    fi

    if [ -n "$data" ]; then
        curl -s -X "$method" "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            ${token:+-H "Authorization: Bearer $token"} \
            -d "$data"
    else
        curl -s -X "$method" "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            ${token:+-H "Authorization: Bearer $token"}
    fi
}

# Check if server is running
check_server() {
    log_section "Checking Server"

    response=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/auth/login" 2>/dev/null || echo "000")

    if [ "$response" = "000" ]; then
        log_error "Cannot connect to server at ${BASE_URL}"
        log_info "Make sure the server is running: ./auth serve"
        exit 1
    fi

    log_success "Server is running at ${BASE_URL}"
}

# ============================================
# AUTH ROUTES TESTS
# ============================================

test_register() {
    log_section "Testing POST /auth/register"

    # Generate unique email
    TIMESTAMP=$(date +%s)
    DATA=$(cat <<EOF
{
    "username": "testuser${TIMESTAMP}",
    "email": "testuser${TIMESTAMP}@example.com",
    "password": "password123",
    "full_name": "Test User",
    "phone": "081234567890",
    "department": "IT"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    # Extract user ID if registration was successful
    if echo "$response" | grep -q "\"id\""; then
        USER_ID=$(echo "$response" | jq -r '.user.id // .id // empty' 2>/dev/null)
        log_success "User registered successfully with ID: $USER_ID"
    else
        log_warning "Registration returned unexpected response"
    fi
}

test_login() {
    log_section "Testing POST /auth/login"

    # First register a user, then login
    TIMESTAMP=$(date +%s)

    # Register first
    REGISTER_DATA=$(cat <<EOF
{
    "username": "logintest${TIMESTAMP}",
    "email": "logintest${TIMESTAMP}@example.com",
    "password": "password123",
    "full_name": "Login Test User"
}
EOF
)

    curl -s -X POST "${BASE_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d "$REGISTER_DATA" > /dev/null

    # Now login
    LOGIN_DATA=$(cat <<EOF
{
    "email": "logintest${TIMESTAMP}@example.com",
    "password": "password123"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d "$LOGIN_DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    if echo "$response" | grep -q '"token"'; then
        TOKEN=$(echo "$response" | jq -r '.token' 2>/dev/null)
        log_success "Login successful, token acquired"
    else
        log_error "Login failed - could not get token"
    fi
}

test_me() {
    log_section "Testing GET /auth/me"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    response=$(curl -s -X GET "${BASE_URL}/auth/me" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    if echo "$response" | grep -q '"email"'; then
        log_success "GET /auth/me successful"
    else
        log_error "GET /auth/me failed"
    fi
}

test_change_password() {
    log_section "Testing POST /auth/change-password"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "old_password": "password123",
    "new_password": "newpassword123"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/auth/change-password" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /auth/change-password executed"
}

test_logout() {
    log_section "Testing POST /auth/logout"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    response=$(curl -s -X POST "${BASE_URL}/auth/logout" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /auth/logout executed"
}

# ============================================
# MFA ROUTES TESTS
# ============================================

test_mfa_status() {
    log_section "Testing GET /mfa/status"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    response=$(curl -s -X GET "${BASE_URL}/mfa/status" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    if echo "$response" | grep -q '"mfa_type"\|"is_enabled"\|"status"'; then
        log_success "GET /mfa/status successful"
    else
        log_error "GET /mfa/status failed"
    fi
}

test_mfa_enroll_totp() {
    log_section "Testing POST /mfa/enroll (TOTP)"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "mfa_type": "totp"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/mfa/enroll" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    if echo "$response" | grep -q '"secret"\|"qr_code_url"\|"mfa_type"'; then
        log_success "POST /mfa/enroll (TOTP) successful"
    else
        log_warning "POST /mfa/enroll may have failed"
    fi
}

test_mfa_enroll_email() {
    log_section "Testing POST /mfa/enroll (Email)"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "mfa_type": "email"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/mfa/enroll" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /mfa/enroll (Email) executed"
}

test_mfa_verify() {
    log_section "Testing POST /mfa/verify"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "code": "123456"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/mfa/verify" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /mfa/verify executed (expected to fail with invalid code)"
}

test_mfa_disable() {
    log_section "Testing POST /mfa/disable"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "password": "password123"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/mfa/disable" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /mfa/disable executed"
}

test_mfa_email_code() {
    log_section "Testing POST /mfa/email-code"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "email": "test@example.com"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/mfa/email-code" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /mfa/email-code executed"
}

test_mfa_verify_email_code() {
    log_section "Testing POST /mfa/verify-email-code"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "code": "123456",
    "email": "test@example.com"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/mfa/verify-email-code" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /mfa/verify-email-code executed"
}

# ============================================
# TRUSTED DEVICES ROUTES TESTS
# ============================================

test_get_trusted_devices() {
    log_section "Testing GET /devices"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    response=$(curl -s -X GET "${BASE_URL}/devices" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "GET /devices executed"
}

test_enroll_trusted_device() {
    log_section "Testing POST /devices/enroll"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    DATA=$(cat <<EOF
{
    "device_fingerprint": "test-fingerprint-$(date +%s)",
    "device_name": "Test Device",
    "device_type": "desktop"
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/devices/enroll" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /devices/enroll executed"
}

test_verify_device_enrollment() {
    log_section "Testing POST /devices/enroll/verify/:token"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    # Use a dummy token for testing
    response=$(curl -s -X POST "${BASE_URL}/devices/enroll/verify/dummy-token" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /devices/enroll/verify/:token executed (expected to fail with invalid token)"
}

test_delete_trusted_device() {
    log_section "Testing DELETE /devices/:device_id"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    # Use a dummy device ID for testing
    response=$(curl -s -X DELETE "${BASE_URL}/devices/00000000-0000-0000-0000-000000000000" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "DELETE /devices/:device_id executed"
}

# ============================================
# USERS ROUTES TESTS
# ============================================

test_list_users() {
    log_section "Testing GET /users"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    response=$(curl -s -X GET "${BASE_URL}/users" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "GET /users executed"
}

test_list_users_with_params() {
    log_section "Testing GET /users with query parameters"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    response=$(curl -s -X GET "${BASE_URL}/users?page=1&page_size=10&search=test" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "GET /users with params executed"
}

test_create_user() {
    log_section "Testing POST /users"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    TIMESTAMP=$(date +%s)
    DATA=$(cat <<EOF
{
    "username": "newuser${TIMESTAMP}",
    "email": "newuser${TIMESTAMP}@example.com",
    "password": "password123",
    "full_name": "New User",
    "department": "Finance",
    "roles": ["user"]
}
EOF
)

    response=$(curl -s -X POST "${BASE_URL}/users" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "POST /users executed"
}

test_get_user() {
    log_section "Testing GET /users/:id"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    # Use the registered user ID or a dummy ID
    TEST_ID="${USER_ID:-00000000-0000-0000-0000-000000000000}"

    response=$(curl -s -X GET "${BASE_URL}/users/${TEST_ID}" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "GET /users/:id executed"
}

test_update_user() {
    log_section "Testing PUT /users/:id"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    TEST_ID="${USER_ID:-00000000-0000-0000-0000-000000000000}"
    DATA=$(cat <<EOF
{
    "full_name": "Updated Name",
    "department": "Updated Department"
}
EOF
)

    response=$(curl -s -X PUT "${BASE_URL}/users/${TEST_ID}" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$DATA")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "PUT /users/:id executed"
}

test_delete_user() {
    log_section "Testing DELETE /users/:id"

    if [ -z "$TOKEN" ]; then
        log_error "No token available, skipping test"
        return
    fi

    # Use a dummy ID for testing
    response=$(curl -s -X DELETE "${BASE_URL}/users/00000000-0000-0000-0000-000000000000" \
        -H "Authorization: Bearer $TOKEN")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    log_success "DELETE /users/:id executed"
}

# ============================================
# UNAUTHORIZED ACCESS TESTS
# ============================================

test_unauthorized_access() {
    log_section "Testing Unauthorized Access (without token)"

    log_info "Testing GET /auth/me without token..."
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "${BASE_URL}/auth/me")
    echo "$response" | head -1 | jq '.' 2>/dev/null || echo "$response"
    log_info "HTTP Status: $(echo "$response" | grep HTTP_CODE | cut -d: -f2)"

    log_info "Testing GET /users without token..."
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "${BASE_URL}/users")
    echo "$response" | head -1 | jq '.' 2>/dev/null || echo "$response"
    log_info "HTTP Status: $(echo "$response" | grep HTTP_CODE | cut -d: -f2)"

    log_info "Testing GET /mfa/status without token..."
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "${BASE_URL}/mfa/status")
    echo "$response" | head -1 | jq '.' 2>/dev/null || echo "$response"
    log_info "HTTP Status: $(echo "$response" | grep HTTP_CODE | cut -d: -f2)"

    log_success "Unauthorized access tests completed"
}

# ============================================
# INVALID TOKEN TESTS
# ============================================

test_invalid_token() {
    log_section "Testing Invalid Token Access"

    log_info "Testing GET /users with invalid token..."
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "${BASE_URL}/users" \
        -H "Authorization: Bearer invalid-token-12345")
    echo "$response" | head -1 | jq '.' 2>/dev/null || echo "$response"
    log_info "HTTP Status: $(echo "$response" | grep HTTP_CODE | cut -d: -f2)"

    log_success "Invalid token tests completed"
}

# ============================================
# RUN ALL TESTS
# ============================================

run_all_tests() {
    log_section "AUTH SERVICE API TEST SUITE"
    log_info "Base URL: ${BASE_URL}"
    log_info "Starting tests..."

    # Check server availability
    check_server

    # Public routes
    log_section "PUBLIC ROUTES"
    test_register
    test_login

    # Unauthorized access tests
    test_unauthorized_access
    test_invalid_token

    # Protected routes (require authentication)
    log_section "PROTECTED ROUTES - AUTH"
    test_me
    test_change_password
    test_logout

    log_section "PROTECTED ROUTES - MFA"
    test_mfa_status
    test_mfa_enroll_totp
    test_mfa_enroll_email
    test_mfa_verify
    test_mfa_disable
    test_mfa_email_code
    test_mfa_verify_email_code

    log_section "PROTECTED ROUTES - TRUSTED DEVICES"
    test_get_trusted_devices
    test_enroll_trusted_device
    test_verify_device_enrollment
    test_delete_trusted_device

    log_section "PROTECTED ROUTES - USERS"
    test_list_users
    test_list_users_with_params
    test_create_user
    test_get_user
    test_update_user
    test_delete_user

    log_section "TEST SUITE COMPLETED"
    log_success "All tests have been executed!"
    log_info "Check the output above for any errors (marked in red)"
}

# Show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -u, --url URL           Base URL (default: http://localhost:8001/api/v1)"
    echo "  --skip-auth            Skip authentication tests"
    echo "  --section SECTION       Run specific section (auth|mfa|devices|users)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all tests on localhost"
    echo "  $0 -u http://localhost:8001/api/v1   # Run on custom URL"
    echo "  $0 --section auth                    # Run only auth tests"
}

# Parse arguments
SKIP_AUTH=false
SECTION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        --skip-auth)
            SKIP_AUTH=true
            shift
            ;;
        --section)
            SECTION="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Run tests based on section
case "$SECTION" in
    auth)
        check_server
        test_register
        test_login
        test_unauthorized_access
        test_invalid_token
        test_me
        test_change_password
        test_logout
        ;;
    mfa)
        check_server
        test_login
        test_mfa_status
        test_mfa_enroll_totp
        test_mfa_enroll_email
        test_mfa_verify
        test_mfa_disable
        test_mfa_email_code
        test_mfa_verify_email_code
        ;;
    devices)
        check_server
        test_login
        test_get_trusted_devices
        test_enroll_trusted_device
        test_verify_device_enrollment
        test_delete_trusted_device
        ;;
    users)
        check_server
        test_login
        test_list_users
        test_list_users_with_params
        test_create_user
        test_get_user
        test_update_user
        test_delete_user
        ;;
    *)
        run_all_tests
        ;;
esac
