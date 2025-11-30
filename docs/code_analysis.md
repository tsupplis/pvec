# Code Analysis Report

*Generated automatically by `make analyze-code`*

## Overview

This report provides comprehensive code analysis including complexity, security, maintainability metrics, and recommendations for the project.

## Cyclomatic Complexity Analysis

### Top 10 Most Complex Functions


| Complexity | Package | Function | File | Line |
|------------|---------|----------|------|------|
| ⚠️ 14 | `helpdialog` | `GetHelpText` | `pkg/ui/helpdialog/helpdialog.go` | 10 |
| ⚠️ 12 | `mainlist` | `(*listModel).View` | `pkg/ui/mainlist/mainlist.go` | 533 |
| ⚠️ 11 | `configpanel` | `(Model).View` | `pkg/ui/configpanel/configpanel.go` | 215 |
| ✅ 9 | `mainlist` | `(*listModel).handleFunctionKeys` | `pkg/ui/mainlist/mainlist.go` | 286 |
| ✅ 8 | `mainlist` | `(*listModel).executeAction` | `pkg/ui/mainlist/mainlist.go` | 481 |
| ✅ 8 | `mainlist` | `(*listModel).handleNavigationKeys` | `pkg/ui/mainlist/mainlist.go` | 352 |
| ✅ 8 | `mainlist` | `(*listModel).Update` | `pkg/ui/mainlist/mainlist.go` | 136 |
| ✅ 8 | `proxmox` | `(*HTTPClient).GetNodes` | `pkg/proxmox/client.go` | 122 |
| ✅ 7 | `mainlist` | `(*listModel).renderRow` | `pkg/ui/mainlist/mainlist.go` | 691 |
| ✅ 7 | `detailsdialog` | `formatValue` | `pkg/ui/detailsdialog/detailsdialog.go` | 312 |

### Functions Requiring Attention (Complexity > 15)

✅ **No high-complexity functions found** - All functions are below the complexity threshold!
## Cognitive Complexity Analysis

### Top 10 Most Cognitively Complex Functions

| Complexity | Package | Function | File | Line |
|------------|---------|----------|------|------|
| 🔶 19 | `helpdialog` | `GetHelpText` | `pkg/ui/helpdialog/helpdialog.go:10` | 1 |
| 🔶 18 | `mainlist` | `(*listModel).View` | `pkg/ui/mainlist/mainlist.go:533` | 1 |
| 🔶 16 | `configpanel` | `(Model).View` | `pkg/ui/configpanel/configpanel.go:215` | 1 |
| ⚠️ 13 | `configpanel` | `(*Model).save` | `pkg/ui/configpanel/configpanel.go:177` | 1 |



## Static Analysis Results

✅ **No code quality issues found!**
## Go Vet Analysis

✅ **No go vet issues found** - Code passes all built-in static checks!


## Staticcheck Analysis

✅ **No staticcheck issues found** - Code meets advanced static analysis standards!


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

- **Total Lines of Code:** 3,912
- **Go Files:** 21
- **Packages:** 10

### Package Details

| Package | Go Files | Test Files | Lines | Test Lines | Test Coverage |
|---------|----------|------------|-------|------------|---------------|
| `pvec` | 1 | 0 | 94 | 0 | 0.0% |
| `test-client` | 1 | 0 | 58 | 0 | N/A |
| `actions` | 1 | 1 | 141 | 144 | 100.0% |
| `config` | 1 | 1 | 102 | 187 | 94.1% |
| `models` | 1 | 1 | 119 | 145 | 100.0% |
| `proxmox` | 3 | 2 | 395 | 426 | 73.2% |
| `actiondialog` | 1 | 1 | 234 | 54 | 91.5% |
| `configpanel` | 1 | 0 | 340 | 0 | 0.0% |
| `detailsdialog` | 1 | 1 | 393 | 71 | 75.9% |
| `helpdialog` | 1 | 1 | 129 | 48 | 94.0% |
| `mainlist` | 1 | 0 | 890 | 0 | 0.0% |
| **TOTAL** | **13** | **8** | **2,895** | **1,075** | - |

### Test Coverage Summary

| Package | Status | Coverage |
|---------|--------|----------|
| `pvec` | ❌ - | 0.0% |
| `actions` | ✅ ok (cached) | 100.0% |
| `config` | ✅ ok (cached) | 94.1% |
| `models` | ✅ ok (cached) | 100.0% |
| `proxmox` | ✅ ok (cached) | 73.2% |
| `actiondialog` | ✅ ok (cached) | 91.5% |
| `configpanel` | ❌ - | 0.0% |
| `detailsdialog` | ✅ ok (cached) | 75.9% |
| `helpdialog` | ✅ ok (cached) | 94.0% |
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


**Generated:** 2025-11-30 00:20:12

---

*To regenerate this report, run: `make analyze-code`*
*Report location: `docs/code_analysis.md`*
