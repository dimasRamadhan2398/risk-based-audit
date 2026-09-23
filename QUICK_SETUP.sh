#!/bin/bash
# Quick setup script to consolidate executive summary routes

echo "═══════════════════════════════════════════════════════════"
echo "EXECUTIVE SUMMARY ROUTE CONSOLIDATION"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Show current state
echo -e "${BLUE}STEP 1: Current State${NC}"
echo "Old files:"
ls -la frontend/pages/audit-result-report/executive-summary*.vue 2>/dev/null || echo "  None found"
echo "New files:"
ls -la frontend/pages/executive-summary/ 2>/dev/null || echo "  None found"
echo ""

# Step 2: Find all references
echo -e "${BLUE}STEP 2: Finding all references to old URL${NC}"
echo "References in codebase:"
grep -r "audit-result-report/executive-summary" frontend/components frontend/pages --include="*.vue" 2>/dev/null | wc -l
echo "Files with old URL:"
grep -r "audit-result-report/executive-summary" frontend/components frontend/pages --include="*.vue" 2>/dev/null | cut -d: -f1 | sort | uniq
echo ""

# Step 3: Check nuxt.config for existing redirects
echo -e "${BLUE}STEP 3: Checking nuxt.config.ts for redirect rules${NC}"
if grep -q "audit-result-report/executive-summary.*redirect" frontend/nuxt.config.ts 2>/dev/null; then
    echo -e "${GREEN}✓ Redirect rules already exist${NC}"
else
    echo -e "${YELLOW}✗ Redirect rules NOT found - Need to add them${NC}"
fi
echo ""

# Step 4: Test dev server
echo -e "${BLUE}STEP 4: Next Steps${NC}"
echo "1. Edit frontend/nuxt.config.ts"
echo "   Add these lines in routeRules section:"
echo ""
echo "   \"/audit-result-report/executive-summary\": { redirect: \"/executive-summary\" },"
echo "   \"/audit-result-report/executive-summary-upload\": { redirect: \"/executive-summary/upload\" },"
echo ""
echo "2. Restart dev server:"
echo "   npm run dev"
echo ""
echo "3. Test URLs:"
echo "   ✓ http://localhost:3000/executive-summary"
echo "   ✓ http://localhost:3000/executive-summary/upload"
echo "   ✓ http://localhost:3000/audit-result-report/executive-summary (should redirect)"
echo ""
echo "4. (Optional) Delete old files:"
echo "   rm frontend/pages/audit-result-report/executive-summary.vue"
echo "   rm frontend/pages/audit-result-report/executive-summary-upload.vue"
echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
