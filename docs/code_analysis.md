# Code Analysis Report

*Generated automatically by `make analyze-code`*

## Overview

This report provides comprehensive code analysis including complexity, security, maintainability metrics, and recommendations for the project.

## Cyclomatic Complexity Analysis

### Top 10 Most Complex Functions


| Complexity | Package | Function | File | Line |
|------------|---------|----------|------|------|
| ✅ 10 | `detailsdialog` | `formatValue` | `pkg/ui/detailsdialog/detailsdialog.go` | 384 |
| ✅ 10 | `main` | `handleKeyEvent` | `main.go` | 102 |
| ✅ 8 | `main` | `handleRuneKey` | `main.go` | 169 |
| ✅ 7 | `configpanel` | `(*ConfigPanel).handleSave` | `pkg/ui/configpanel/configpanel.go` | 81 |
| ✅ 7 | `proxmox` | `(*HTTPClient).GetNodes` | `pkg/proxmox/client.go` | 123 |
| ✅ 7 | `config` | `(*ViperLoader).Load` | `pkg/config/config.go` | 44 |
| ✅ 7 | `main` | `executeAction` | `main.go` | 266 |
| ✅ 7 | `main` | `setupKeyHandlers` | `examples/test-ui/main.go` | 86 |
| ✅ 7 | `main` | `main` | `examples/test-client/main.go` | 13 |
| ✅ 6 | `mainlist` | `(*MainList).updateTable` | `pkg/ui/mainlist/mainlist.go` | 204 |

### Functions Requiring Attention (Complexity > 15)

✅ **No high-complexity functions found** - All functions are below the complexity threshold!
## Cognitive Complexity Analysis

### Top 10 Most Cognitively Complex Functions

✅ **No complex functions found** - All functions have low cognitive complexity!



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

- **Total Lines of Code:** 4,359
- **Go Files:** 26
- **Packages:** 11

### Package Details

| Package | Go Files | Test Files | Lines | Test Lines | Test Coverage |
|---------|----------|------------|-------|------------|---------------|
| `pvec` | 1 | 0 | 355 | 0 | 0.0% |
| `test-client` | 1 | 0 | 58 | 0 | N/A |
| `test-ui` | 1 | 0 | 156 | 0 | N/A |
| `actions` | 1 | 1 | 141 | 144 | 100.0% |
| `config` | 1 | 1 | 101 | 187 | 93.9% |
| `models` | 1 | 1 | 119 | 145 | 100.0% |
| `proxmox` | 3 | 2 | 457 | 450 | 73.9% |
| `actiondialog` | 1 | 1 | 107 | 126 | 81.0% |
| `colors` | 1 | 1 | 38 | 105 | N/A |
| `configpanel` | 1 | 1 | 198 | 141 | 67.2% |
| `detailsdialog` | 1 | 1 | 452 | 359 | 69.2% |
| `helpdialog` | 1 | 1 | 96 | 36 | 76.5% |
| `mainlist` | 1 | 1 | 341 | 261 | 89.7% |
| **TOTAL** | **15** | **11** | **2,619** | **1,954** | - |

### Test Coverage Summary

| Package | Status | Coverage |
|---------|--------|----------|
| `pvec` | ❌ - | 0.0% |
| `actions` | ✅ ok (cached) | 100.0% |
| `config` | ✅ ok (cached) | 93.9% |
| `models` | ✅ ok (cached) | 100.0% |
| `proxmox` | ✅ ok | 73.9% |
| `actiondialog` | ✅ ok (cached) | 81.0% |
| `colors` | ✅ ok (cached) | N/A |
| `configpanel` | ✅ ok (cached) | 67.2% |
| `detailsdialog` | ✅ ok | 69.2% |
| `helpdialog` | ✅ ok (cached) | 76.5% |
| `mainlist` | ✅ ok (cached) | 89.7% |

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


**Generated:** 2025-11-12 20:21:02

---

*To regenerate this report, run: `make analyze-code`*
*Report location: `docs/code_analysis.md`*
