# Code Analysis Report

*Generated automatically by `make analyze-code`*

## Overview

This report provides comprehensive code analysis including complexity, security, maintainability metrics, and recommendations for the project.

## Cyclomatic Complexity Analysis

### Top 10 Most Complex Functions


| Complexity | Package | Function | File | Line |
|------------|---------|----------|------|------|
| ✅ 9 | `mainlist` | `(*listModel).renderMainList` | `pkg/ui/mainlist/mainlist.go` | 586 |
| ✅ 9 | `mainlist` | `(*listModel).handleFunctionKeys` | `pkg/ui/mainlist/mainlist.go` | 311 |
| ✅ 9 | `configpanel` | `(Model).View` | `pkg/ui/configpanel/configpanel.go` | 237 |
| ✅ 8 | `mainlist` | `(*listModel).View` | `pkg/ui/mainlist/mainlist.go` | 559 |
| ✅ 8 | `mainlist` | `(*listModel).executeAction` | `pkg/ui/mainlist/mainlist.go` | 507 |
| ✅ 8 | `mainlist` | `(*listModel).handleNavigationKeys` | `pkg/ui/mainlist/mainlist.go` | 378 |
| ✅ 8 | `mainlist` | `(*listModel).Update` | `pkg/ui/mainlist/mainlist.go` | 135 |
| ✅ 8 | `configpanel` | `(Model).handleKeyMsg` | `pkg/ui/configpanel/configpanel.go` | 125 |
| ✅ 8 | `proxmox` | `(*HTTPClient).GetNodes` | `pkg/proxmox/client.go` | 122 |
| ✅ 7 | `mainlist` | `(*listModel).renderRow` | `pkg/ui/mainlist/mainlist.go` | 658 |

### Functions Requiring Attention (Complexity > 15)

✅ **No high-complexity functions found** - All functions are below the complexity threshold!
## Cognitive Complexity Analysis

### Top 10 Most Cognitively Complex Functions

| Complexity | Package | Function | File | Line |
|------------|---------|----------|------|------|
| ⚠️ 14 | `configpanel` | `(Model).View` | `pkg/ui/configpanel/configpanel.go:237` | 1 |
| ⚠️ 14 | `mainlist` | `(*listModel).renderMainList` | `pkg/ui/mainlist/mainlist.go:586` | 1 |
| ⚠️ 13 | `configpanel` | `(*Model).save` | `pkg/ui/configpanel/configpanel.go:199` | 1 |
| ⚠️ 11 | `mainlist` | `(*listModel).handleConfigPanelMsg` | `pkg/ui/mainlist/mainlist.go:160` | 1 |



## Static Analysis Results

### Code Quality Issues

| File | Line | Column | Issue | Linter |
|------|------|--------|-------|--------|
| `pkg/ui/mainlist/mainlist.go` | 641 | 16 | SA1019: strings.Title has been deprecated since Go 1.18 and an alternative has been available since Go 1.0: The rule Title uses for word boundaries does not handle Unicode punctuation properly. Use golang.org/x/text/cases instead. | `staticcheck` |
## Go Vet Analysis

✅ **No go vet issues found** - Code passes all built-in static checks!


## Staticcheck Analysis

### Staticcheck Issues Found: 1

| File | Line | Column | Issue | Rule |
|------|------|--------|-------|------|
| `pkg/ui/mainlist/mainlist.go` | 641 | 16 | strings.Title has been deprecated since Go 1.18 and an alternative has been available since Go 1.0: The rule Title uses for word boundaries does not handle Unicode punctuation properly. Use golang.org/x/text/cases instead. | `SA1019` |


## Security Analysis Results

✅ **No security issues found** - Great job maintaining secure code!


## Vulnerability Analysis (govulncheck)

✅ **No known vulnerabilities found** - Dependencies are secure!


## Code Quality Issues

✅ **No code smells detected** - Code is clean and well-structured!


## Dead Code Analysis

✅ **No unused code found** - All functions are properly utilized!


## Architecture Analysis

✅ **Clean Architecture Maintained** - No dependency violations detected!



## Code Metrics

### Project Overview

- **Total Lines of Code:** 3,584
- **Go Files:** 19
- **Packages:** 9

### Package Details

| Package | Go Files | Test Files | Lines | Test Lines | Test Coverage |
|---------|----------|------------|-------|------------|---------------|
| `pvec` | 1 | 0 | 94 | 0 | 0.0% |
| `test-client` | 1 | 0 | 58 | 0 | N/A |
| `actions` | 1 | 1 | 141 | 144 | 100.0% |
| `config` | 1 | 1 | 102 | 187 | 94.1% |
| `models` | 1 | 1 | 119 | 145 | 100.0% |
| `proxmox` | 3 | 2 | 395 | 426 | 73.2% |
| `configpanel` | 1 | 0 | 352 | 0 | 0.0% |
| `detailsdialog` | 1 | 1 | 381 | 71 | 74.5% |
| `helpdialog` | 1 | 1 | 95 | 48 | 100.0% |
| `mainlist` | 1 | 0 | 884 | 0 | 0.0% |
| **TOTAL** | **12** | **7** | **2,621** | **1,021** | - |

### Test Coverage Summary

| Package | Status | Coverage |
|---------|--------|----------|
| `pvec` | ❌ - | 0.0% |
| `actions` | ✅ ok (cached) | 100.0% |
| `config` | ✅ ok (cached) | 94.1% |
| `models` | ✅ ok (cached) | 100.0% |
| `proxmox` | ✅ ok (cached) | 73.2% |
| `configpanel` | ❌ - | 0.0% |
| `detailsdialog` | ✅ ok | 74.5% |
| `helpdialog` | ✅ ok | 100.0% |
| `mainlist` | ❌ - | 0.0% |

## Complexity Guidelines

### Cyclomatic Complexity Scale

| Range | Level | Icon | Description | Action Required |
|-------|-------|------|-------------|-----------------|
| 1-10 | Simple | ✅ | Easy to test and maintain | None - keep as is |
| 11-15 | Moderate | ⚠️ | Consider refactoring for clarity | Review and monitor |
| 16-25 | High | 🔶 | Should be refactored | Plan refactoring |
| 26+ | Very High | 🔴 | Requires immediate attention | Refactor immediately |

### Recommendations

Based on the analysis above:

1. **High Priority**: Functions with complexity > 20 should be refactored immediately
2. **Medium Priority**: Functions with complexity 15-20 should be reviewed and possibly split
3. **Test Functions**: High complexity in test functions may indicate need for test helpers or sub-tests
4. **Maintain Coverage**: Keep test coverage above 80% while reducing complexity

### Refactoring Strategies

- **Extract Method**: Break large functions into smaller, focused functions
- **Extract Class/Package**: Group related functionality
- **Use Table-Driven Tests**: Reduce complexity in test functions
- **State Machines**: For complex conditional logic
- **Strategy Pattern**: For multiple algorithmic approaches

## Analysis Timestamp


**Generated:** 2025-11-30 01:28:03

---

*To regenerate this report, run: `make analyze-code`*
*Report location: `docs/code_analysis.md`*
