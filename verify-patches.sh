#!/bin/bash
# Verification script for doc hygiene patches
# Run after applying all patches to verify consistency

echo "🔍 Verifying doc hygiene patches..."
echo ""

PASS=0
FAIL=0

# Test 1: Check evidence citations exist
echo "✓ Test 1: Evidence citations"
if grep -q "swagger.json#/paths" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md; then
    echo "  ✅ Found swagger.json citations"
    ((PASS++))
else
    echo "  ❌ Missing swagger.json citations"
    ((FAIL++))
fi

# Test 2: Check POST /applications documented as missing
echo "✓ Test 2: POST /applications status"
if grep -q "NO POST\|no POST\|does NOT exist" SEVALLA_SPEC_FINDINGS.md; then
    echo "  ✅ POST /applications absence documented"
    ((PASS++))
else
    echo "  ❌ POST /applications status unclear"
    ((FAIL++))
fi

# Test 3: Check database strategy is deprecate (not fix)
echo "✓ Test 3: Database strategy"
if grep -q "Deprecate.*kinsta_database" ANALYSIS_SUMMARY.md; then
    echo "  ✅ Database deprecation strategy found"
    ((PASS++))
else
    echo "  ❌ Database strategy unclear"
    ((FAIL++))
fi

# Test 4: Check operation.data described as opaque
echo "✓ Test 4: operation.data opaqueness"
if grep -q "opaque\|OPAQUE" SEVALLA_SPEC_FINDINGS.md; then
    echo "  ✅ operation.data opaqueness documented"
    ((PASS++))
else
    echo "  ❌ operation.data opaqueness not clear"
    ((FAIL++))
fi

# Test 5: Check MUST NOT language for exclusions
echo "✓ Test 5: Explicit exclusions"
if grep -q "MUST NOT" specs/00-adr-provider-split.md; then
    echo "  ✅ MUST NOT exclusions found in ADR"
    ((PASS++))
else
    echo "  ❌ Exclusions not explicit enough"
    ((FAIL++))
fi

# Test 6: Check synchronous database operations documented
echo "✓ Test 6: Database synchronous operations"
if grep -q "synchronous\|Synchronous.*200" ANALYSIS_SUMMARY.md; then
    echo "  ✅ Synchronous database operations documented"
    ((PASS++))
else
    echo "  ❌ Database operation type unclear"
    ((FAIL++))
fi

# Test 7: Check application priority reflects blocker
echo "✓ Test 7: Application resource priority"
if grep -q "P2.*blocked\|blocked.*P2" SPECS_ROADMAP.md ANALYSIS_SUMMARY.md; then
    echo "  ✅ Application resource marked as blocked"
    ((PASS++))
else
    echo "  ❌ Application blocker not reflected in priority"
    ((FAIL++))
fi

# Test 8: Check manual state migration documented
echo "✓ Test 8: State migration clarity"
if grep -q "manual\|cannot automatically" ANALYSIS_SUMMARY.md; then
    echo "  ✅ Manual state migration documented"
    ((PASS++))
else
    echo "  ❌ State migration process unclear"
    ((FAIL++))
fi

# Test 9: Check canonical docs exist
echo "✓ Test 9: Canonical docs"
if [ -f "specs/00-adr-provider-split.md" ] && [ -f "specs/02-operations-polling-contract.md" ]; then
    echo "  ✅ Both canonical docs exist"
    ((PASS++))
else
    echo "  ❌ Missing canonical docs"
    ((FAIL++))
fi

# Test 10: Check static sites status clarified
echo "✓ Test 10: Static sites status"
if grep -q "deprecated.*MyKinsta.*active.*Sevalla" ANALYSIS_SUMMARY.md SEVALLA_SPEC_FINDINGS.md; then
    echo "  ✅ Static sites status clarified"
    ((PASS++))
else
    echo "  ❌ Static sites status unclear"
    ((FAIL++))
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Results: $PASS passed, $FAIL failed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $FAIL -eq 0 ]; then
    echo "✅ All verification tests passed!"
    exit 0
else
    echo "❌ Some tests failed. Review PATCH_PLAN.md"
    exit 1
fi
