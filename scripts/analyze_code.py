#!/usr/bin/env python3
"""
Code Complexity Analysis Script for BatchExec

Generates comprehensive complexity analysis and writes to docs/complexity.md

This Python version provides:
- Better error handling and validation
- More readable code structure
- Easier maintenance and extension
- Better test coverage parsing
- Configuration support
"""

import os
import sys
import subprocess
import json
import re
import configparser
from datetime import datetime
from typing import List, Dict, Tuple, Optional
from dataclasses import dataclass
from pathlib import Path


@dataclass
class ComplexityEntry:
    """Represents a single complexity measurement."""
    complexity: int
    package: str
    function: str
    file: str
    line: int


@dataclass
class CoverageEntry:
    """Represents test coverage for a package."""
    package: str
    status: str
    coverage: str
    cached: bool = False


@dataclass
class SecurityIssue:
    """Represents a security issue found by gosec."""
    severity: str
    confidence: str
    rule: str
    details: str
    file: str
    line: int


@dataclass
class CodeSmell:
    """Represents code smells found by goconst."""
    type: str
    message: str
    file: str
    line: int


@dataclass
class ArchitectureViolation:
    """Represents architecture violations found by go-cleanarch."""
    violation_type: str
    details: str
    file: str


@dataclass
class CodeMetrics:
    """Overall code metrics."""
    total_lines: int
    go_files: int
    packages: int
    coverage_entries: List[CoverageEntry]
    package_metrics: List[Dict[str, any]]


class AnalysisConfig:
    """Configuration for the analysis."""
    
    def __init__(self, config_file: Optional[str] = None):
        # Default values
        self.output_file = "docs/code_analysis.md"
        self.complexity_threshold = 15
        self.top_functions = 10
        self.exclude_test_files = True
        self.exclude_vendor = True
        self.exclude_examples = True
        self.deadcode_exclude_files = ["demo/", "interfaces.go"]
        self.deadcode_exclude_functions = []
        self.deadcode_exclude_public_interfaces = True  # Skip public funcs in interface files
        self.project_name = "Go Project"
        self.package_pattern = "github.com/*/batchexec"  # Generic pattern
        
        # Load from config file if provided
        if config_file and Path(config_file).exists():
            self._load_from_file(config_file)
        elif Path("analyze_config.ini").exists():
            self._load_from_file("analyze_config.ini")
    
    def _load_from_file(self, config_file: str):
        """Load configuration from INI file."""
        config = configparser.ConfigParser()
        config.read(config_file)
        
        if 'analysis' in config:
            section = config['analysis']
            self.output_file = section.get('output_file', self.output_file)
            self.complexity_threshold = section.getint('complexity_threshold', self.complexity_threshold)
            self.top_functions = section.getint('top_functions', self.top_functions)
            self.exclude_test_files = section.getboolean('exclude_test_files', self.exclude_test_files)
            self.exclude_vendor = section.getboolean('exclude_vendor', self.exclude_vendor)
            self.exclude_examples = section.getboolean('exclude_examples', self.exclude_examples)
            self.project_name = section.get('project_name', self.project_name)
            self.package_pattern = section.get('package_pattern', self.package_pattern)
        
        if 'deadcode' in config:
            section = config['deadcode']
            exclude_files = section.get('exclude_files', '')
            if exclude_files:
                self.deadcode_exclude_files = [f.strip() for f in exclude_files.split(',')]
            
            exclude_functions = section.get('exclude_functions', '')
            if exclude_functions:
                self.deadcode_exclude_functions = [f.strip() for f in exclude_functions.split(',')]
            
            self.deadcode_exclude_public_interfaces = section.getboolean('exclude_public_interfaces', self.deadcode_exclude_public_interfaces)
    

class CodeAnalyzer:
    """Main code analysis class."""
    
    def __init__(self, config: AnalysisConfig):
        self.config = config
        self.project_root = Path.cwd()
        
    def check_required_tools(self) -> bool:
        """Check if required tools are installed."""
        tools = ['gocyclo', 'golangci-lint', 'go', 'gosec', 'goconst', 'gocognit', 'guru', 'go-cleanarch', 'govulncheck', 'staticcheck', 'deadcode']
        missing_tools = []
        
        for tool in tools:
            if not self._command_exists(tool):
                missing_tools.append(tool)
        
        if missing_tools:
            print(f"❌ Missing tools: {', '.join(missing_tools)}")
            print("   Please run 'make install-tools' first.")
            return False
        
        return True
    
    def _command_exists(self, command: str) -> bool:
        """Check if a command exists in PATH."""
        try:
            subprocess.run([command, '--help'], 
                          capture_output=True, 
                          check=False, 
                          timeout=10)
            return True
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def _get_package_list(self, as_args=False) -> List[str]:
        """Get list of Go packages, respecting exclusion settings."""
        try:
            result = subprocess.run(['go', 'list', './...'], 
                                  capture_output=True, text=True, check=True)
            packages = [p.strip() for p in result.stdout.strip().split('\n') if p.strip()]
            
            if self.config.exclude_examples:
                packages = [p for p in packages if '/examples/' not in p]
                
            return packages if as_args else packages
        except subprocess.SubprocessError:
            return ['./...'] if as_args else []
    
    def get_complexity_data(self) -> List[ComplexityEntry]:
        """Get cyclomatic complexity data using gocyclo."""
        print("🧮 Analyzing cyclomatic complexity...")
        
        cmd = ['gocyclo', '-over', '1']
        if self.config.exclude_test_files:
            cmd.extend(['-ignore', '_test'])
        cmd.append('.')
        
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=False)
            if result.returncode != 0 and result.stderr:
                print(f"⚠️  gocyclo warning: {result.stderr.strip()}")
            
            return self._parse_complexity_output(result.stdout)
        except subprocess.SubprocessError as e:
            print(f"❌ Error running gocyclo: {e}")
            return []
    
    def _parse_complexity_output(self, output: str) -> List[ComplexityEntry]:
        """Parse gocyclo output into structured data."""
        entries = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
                
            # Parse line format: "15 main main main.go:28:1"
            match = re.match(r'^(\d+)\s+(\S+)\s+(\S+)\s+(.+):(\d+):\d+$', line)
            if match:
                complexity = int(match.group(1))
                package = match.group(2)
                function = match.group(3)
                file = match.group(4)
                line_num = int(match.group(5))
                
                entries.append(ComplexityEntry(
                    complexity=complexity,
                    package=package,
                    function=function,
                    file=file,
                    line=line_num
                ))
        
        return sorted(entries, key=lambda x: x.complexity, reverse=True)
    
    def get_static_analysis(self) -> Dict[str, any]:
        """Run golangci-lint for static analysis."""
        print("🔧 Running static analysis...")
        
        try:
            # First, try without configuration to avoid version conflicts
            result = subprocess.run(['golangci-lint', 'run', '--no-config'], 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            # If no-config fails, try with configuration
            if result.returncode != 0 and "unknown flag" in result.stderr:
                result = subprocess.run(['golangci-lint', 'run'], 
                                      capture_output=True, 
                                      text=True, 
                                      check=False)
            
            issues = []
            if result.stdout.strip():
                issues = self._parse_golangci_output(result.stdout)
            
            # Check if error is about version compatibility
            error_msg = result.stderr.strip() if result.stderr.strip() else None
            if error_msg and "configuration file for golangci-lint v2 with golangci-lint v1" in error_msg:
                error_msg = "Version mismatch: Using golangci-lint v1 with v2 config. Consider upgrading golangci-lint or updating .golangci.yml"
            
            return {
                'issues': issues,
                'error': error_msg,
                'success': len(issues) == 0 and not error_msg
            }
                
        except subprocess.SubprocessError as e:
            return {
                'issues': [],
                'error': f"Error running golangci-lint: {e}",
                'success': False
            }
    
    def get_security_analysis(self) -> List[SecurityIssue]:
        """Run gosec for security analysis."""
        print("🔒 Running security analysis...")
        
        try:
            packages = self._get_package_list()
            if not packages:
                return []
                
            cmd = ['gosec', '-fmt=json'] + packages
            result = subprocess.run(cmd, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            if result.stdout.strip():
                return self._parse_gosec_output(result.stdout)
            return []
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running gosec: {e}")
            return []
    
    def _parse_gosec_output(self, output: str) -> List[SecurityIssue]:
        """Parse gosec JSON output."""
        issues = []
        try:
            data = json.loads(output)
            if 'Issues' in data:
                for issue in data['Issues']:
                    file_path = issue.get('file', '')
                    # Truncate absolute path to project relative path
                    if file_path.startswith('/'):
                        project_root = str(self.project_root)
                        if file_path.startswith(project_root):
                            file_path = file_path[len(project_root):].lstrip('/')
                    
                    issues.append(SecurityIssue(
                        severity=issue.get('severity', 'UNKNOWN'),
                        confidence=issue.get('confidence', 'UNKNOWN'),
                        rule=issue.get('rule_id', ''),
                        details=issue.get('details', ''),
                        file=file_path,
                        line=int(issue.get('line', 0))
                    ))
        except (json.JSONDecodeError, KeyError, ValueError) as e:
            print(f"⚠️  Error parsing gosec output: {e}")
        
        return issues
    
    def get_cognitive_complexity(self) -> List[ComplexityEntry]:
        """Get cognitive complexity data using gocognit."""
        print("🧠 Analyzing cognitive complexity...")
        
        try:
            result = subprocess.run(['gocognit', '-over', '10', '.'], 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            # Parse and filter out test files if configured
            entries = self._parse_gocognit_output(result.stdout)
            if self.config.exclude_test_files:
                entries = [e for e in entries if not ('_test.go' in e.file or e.function.startswith('Test') or e.function.startswith('Benchmark') or e.function.startswith('Example'))]
            
            return entries
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running gocognit: {e}")
            return []
    
    def _parse_gocognit_output(self, output: str) -> List[ComplexityEntry]:
        """Parse gocognit output."""
        entries = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
                
            # Parse line format: "15 main main main.go:28"
            match = re.match(r'^(\d+)\s+(\S+)\s+(\S+)\s+(.+):(\d+)$', line)
            if match:
                complexity = int(match.group(1))
                package = match.group(2)
                function = match.group(3)
                file = match.group(4)
                line_num = int(match.group(5))
                
                entries.append(ComplexityEntry(
                    complexity=complexity,
                    package=package,
                    function=function,
                    file=file,
                    line=line_num
                ))
        
        return sorted(entries, key=lambda x: x.complexity, reverse=True)
    
    def get_code_smells(self) -> List[CodeSmell]:
        """Get code smells using goconst."""
        print("👃 Detecting code smells...")
        
        try:
            packages = self._get_package_list()
            if not packages:
                return []
                
            cmd = ['goconst'] + packages
            result = subprocess.run(cmd, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._parse_goconst_output(result.stdout)
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running goconst: {e}")
            return []
    
    def _parse_goconst_output(self, output: str) -> List[CodeSmell]:
        """Parse goconst output."""
        smells = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
                
            # Parse line format: "./path/file.go:123:45: string literal found 3 times"
            match = re.match(r'^(.+):(\d+):\d+: (.+)$', line)
            if match:
                file = match.group(1)
                line_num = int(match.group(2))
                message = match.group(3)
                
                smells.append(CodeSmell(
                    type="string_duplication",
                    message=message,
                    file=file,
                    line=line_num
                ))
        
        return smells
    
    def get_architecture_analysis(self) -> List[ArchitectureViolation]:
        """Run go-cleanarch for architecture analysis."""
        print("🏗️  Analyzing architecture...")
        
        try:
            result = subprocess.run(['go-cleanarch'], 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._parse_cleanarch_output(result.stdout, result.stderr)
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running go-cleanarch: {e}")
            return []
    
    def _parse_cleanarch_output(self, stdout: str, stderr: str) -> List[ArchitectureViolation]:
        """Parse go-cleanarch output."""
        violations = []
        
        # go-cleanarch outputs to stderr for violations
        output = stderr if stderr.strip() else stdout
        
        for line in output.strip().split('\n'):
            if not line.strip() or "Clean Architecture" in line:
                continue
                
            if "dependency rule" in line.lower() or "import" in line.lower():
                violations.append(ArchitectureViolation(
                    violation_type="dependency_violation",
                    details=line.strip(),
                    file=""
                ))
        
        return violations
    
    def get_staticcheck_analysis(self) -> List[Dict[str, any]]:
        """Run staticcheck for advanced static analysis."""
        print("🔬 Running staticcheck analysis...")
        
        try:
            packages = self._get_package_list()
            if not packages:
                return []
                
            cmd = ['staticcheck'] + packages
            result = subprocess.run(cmd, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._parse_staticcheck_output(result.stdout)
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running staticcheck: {e}")
            return []
    
    def _parse_staticcheck_output(self, output: str) -> List[Dict[str, any]]:
        """Parse staticcheck output."""
        issues = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
                
            # Parse format: "./file.go:line:col: message (SA1000)"
            match = re.match(r'^(.+):(\d+):(\d+):\s+(.+?)\s*(?:\(([^)]+)\))?$', line)
            if match:
                file_path = match.group(1)
                # Truncate path to be relative to project root
                if file_path.startswith('./'):
                    file_path = file_path[2:]
                
                issues.append({
                    'file': file_path,
                    'line': match.group(2),
                    'col': match.group(3),
                    'message': match.group(4),
                    'rule': match.group(5) if match.group(5) else 'staticcheck'
                })
        
        return issues
    
    def get_deadcode_analysis(self) -> List[Dict[str, any]]:
        """Run deadcode analysis to find unused code.
        
        Uses -test flag to only report functions not covered by tests,
        helping distinguish actual dead code from untested public APIs.
        Filters out demo packages and public interface functions.
        """
        print("💀 Detecting dead code...")
        
        try:
            packages = self._get_package_list()
            if not packages:
                return []
                
            cmd = ['deadcode', '-test'] + packages
            result = subprocess.run(cmd, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._filter_deadcode_results(self._parse_deadcode_output(result.stdout))
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running deadcode: {e}")
            return []
    
    def _parse_deadcode_output(self, output: str) -> List[Dict[str, any]]:
        """Parse deadcode output."""
        dead_items = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
                
            # Parse format: "file.go:line:col: function Name is unused"
            match = re.match(r'^(.+):(\d+):(\d+): (.+)$', line)
            if match:
                file_path = match.group(1)
                if file_path.startswith('./'):
                    file_path = file_path[2:]
                
                dead_items.append({
                    'file': file_path,
                    'line': match.group(2),
                    'col': match.group(3),
                    'message': match.group(4)
                })
        
        return dead_items
    
    def _filter_deadcode_results(self, dead_code_items: List[Dict[str, any]]) -> List[Dict[str, any]]:
        """Filter dead code results based on configuration."""
        filtered_items = []
        
        for item in dead_code_items:
            file_path = item['file']
            message = item['message']
            
            # Skip files matching exclude patterns
            should_skip_file = False
            for exclude_pattern in self.config.deadcode_exclude_files:
                if exclude_pattern in file_path or file_path.startswith(exclude_pattern):
                    should_skip_file = True
                    break
            
            if should_skip_file:
                continue
            
            # Skip specific functions if configured
            should_skip_func = False
            for exclude_func in self.config.deadcode_exclude_functions:
                if exclude_func in message:
                    should_skip_func = True
                    break
            
            if should_skip_func:
                continue
            
            # Skip public functions in interface files if configured
            if (self.config.deadcode_exclude_public_interfaces and 
                file_path.endswith('interfaces.go') and 
                'unreachable func:' in message):
                
                # Extract function name from message like "unreachable func: OSCmd.Run"
                func_part = message.split('unreachable func:')[1].strip()
                if '.' in func_part:
                    func_name = func_part.split('.')[-1]
                else:
                    func_name = func_part
                
                # Skip if function name starts with uppercase (public)
                if func_name and func_name[0].isupper():
                    continue
            
            filtered_items.append(item)
        
        return filtered_items
    
    def get_vulnerability_analysis(self) -> List[Dict[str, any]]:
        """Run govulncheck for vulnerability analysis."""
        print("🛡️  Checking for known vulnerabilities...")
        
        try:
            packages = self._get_package_list()
            if not packages:
                return []
                
            cmd = ['govulncheck'] + packages
            result = subprocess.run(cmd, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._parse_govulncheck_output(result.stdout, result.stderr)
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running govulncheck: {e}")
            return []
    
    def _parse_govulncheck_output(self, stdout: str, stderr: str) -> List[Dict[str, any]]:
        """Parse govulncheck output."""
        vulnerabilities = []
        output = stdout + stderr
        
        # Look for vulnerability patterns
        lines = output.strip().split('\n')
        for i, line in enumerate(lines):
            if 'vulnerability' in line.lower() or 'CVE-' in line:
                vulnerabilities.append({
                    'type': 'vulnerability',
                    'message': line.strip(),
                    'details': lines[i+1:i+3] if i+1 < len(lines) else []
                })
        
        return vulnerabilities
    
    def get_vet_analysis(self) -> List[Dict[str, any]]:
        """Run go vet for static analysis."""
        print("🔍 Running go vet analysis...")
        
        try:
            packages = self._get_package_list()
            if not packages:
                return []
                
            cmd = ['go', 'vet'] + packages
            result = subprocess.run(cmd, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._parse_vet_output(result.stderr)  # go vet outputs to stderr
            
        except subprocess.SubprocessError as e:
            print(f"⚠️  Error running go vet: {e}")
            return []
    
    def _parse_vet_output(self, output: str) -> List[Dict[str, any]]:
        """Parse go vet output."""
        issues = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
                
            # Parse format: "./file.go:line:col: message"
            match = re.match(r'^(.+):(\d+):(\d+): (.+)$', line)
            if match:
                file_path = match.group(1)
                # Truncate path to be relative to project root
                if file_path.startswith('./'):
                    file_path = file_path[2:]
                
                issues.append({
                    'file': file_path,
                    'line': match.group(2),
                    'col': match.group(3),
                    'message': match.group(4)
                })
        
        return issues
    
    def get_code_metrics(self) -> CodeMetrics:
        """Gather overall code metrics."""
        print("📊 Gathering code metrics...")
        
        # Count lines of code
        total_lines = self._count_lines_of_code()
        
        # Count Go files
        go_files = len(list(self.project_root.glob('**/*.go'))) - \
                  len(list(self.project_root.glob('**/vendor/**/*.go')))
        
        # Count packages
        packages = self._count_packages()
        
        # Get test coverage
        coverage_entries = self._get_test_coverage()
        
        # Get package-level metrics
        package_metrics = self._get_package_metrics()
        
        return CodeMetrics(
            total_lines=total_lines,
            go_files=go_files,
            packages=packages,
            coverage_entries=coverage_entries,
            package_metrics=package_metrics
        )
    
    def _count_lines_of_code(self) -> int:
        """Count total lines of Go code (excluding vendor and examples)."""
        try:
            cmd = ['find', '.', '-name', '*.go', '-not', '-path', './vendor/*']
            if self.config.exclude_examples:
                cmd.extend(['-not', '-path', './examples/*'])
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            
            files = result.stdout.strip().split('\n')
            if not files or files == ['']:
                return 0
                
            wc_cmd = ['wc', '-l'] + files
            wc_result = subprocess.run(wc_cmd, capture_output=True, text=True, check=True)
            
            lines = wc_result.stdout.strip().split('\n')
            total_line = lines[-1]  # Last line contains total
            return int(total_line.split()[0])
            
        except (subprocess.SubprocessError, ValueError, IndexError):
            return 0
    
    def _count_packages(self) -> int:
        """Count the number of Go packages."""
        try:
            result = subprocess.run(['go', 'list', './...'], 
                                  capture_output=True, 
                                  text=True, 
                                  check=True)
            packages = result.stdout.strip().split('\n')
            packages = [p for p in packages if p.strip()]
            
            if self.config.exclude_examples:
                packages = [p for p in packages if '/examples/' not in p]
                
            return len(packages)
        except subprocess.SubprocessError:
            return 0
    
    def _get_test_coverage(self) -> List[CoverageEntry]:
        """Get test coverage information."""
        print("🧪 Getting test coverage...")
        
        try:
            # Get packages and filter if needed
            result = subprocess.run(['go', 'list', './...'], 
                                  capture_output=True, text=True, check=True)
            packages = [p.strip() for p in result.stdout.strip().split('\n') if p.strip()]
            
            if self.config.exclude_examples:
                packages = [p for p in packages if '/examples/' not in p]
            
            if not packages:
                return []
                
            # Run tests with coverage on filtered packages
            result = subprocess.run(['go', 'test', '-cover'] + packages, 
                                  capture_output=True, 
                                  text=True, 
                                  check=False)
            
            return self._parse_coverage_output(result.stdout)
        except subprocess.SubprocessError:
            return [CoverageEntry("Error", "Unable to run tests", "N/A")]
    
    def _parse_coverage_output(self, output: str) -> List[CoverageEntry]:
        """Parse test coverage output."""
        entries = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
            
            # Parse "ok" lines: "ok  	github.com/tsupplis/batchexec/config	(cached)	coverage: 46.9% of statements"
            if line.startswith('ok'):
                parts = line.split()
                if len(parts) >= 2:
                    package = parts[1]
                    cached = '(cached)' in line
                    status = 'ok (cached)' if cached else 'ok'
                    
                    coverage_match = re.search(r'coverage: ([0-9.]+%)', line)
                    coverage = coverage_match.group(1) if coverage_match else 'N/A'
                    
                    entries.append(CoverageEntry(package, status, coverage, cached))
            
            # Parse other package lines with coverage info
            elif 'coverage:' in line:
                parts = line.strip().split()
                if parts:
                    package = parts[0]
                    coverage_match = re.search(r'coverage: ([0-9.]+%)', line)
                    coverage = coverage_match.group(1) if coverage_match else 'N/A'
                    
                    entries.append(CoverageEntry(package, '-', coverage))
        
        return entries
    
    def _parse_golangci_output(self, output: str) -> List[Dict[str, str]]:
        """Parse golangci-lint output into structured data."""
        issues = []
        
        for line in output.strip().split('\n'):
            if not line.strip():
                continue
            
            # Parse format: "file.go:line:col: message (linter)"
            match = re.match(r'^(.+):(\d+):(\d+):\s+(.+?)\s+\(([^)]+)\)$', line)
            if match:
                issues.append({
                    'file': match.group(1),
                    'line': match.group(2),
                    'col': match.group(3),
                    'message': match.group(4),
                    'linter': match.group(5)
                })
        
        return issues
    
    def _get_package_metrics(self) -> List[Dict[str, any]]:
        """Get detailed metrics per package."""
        print("📊 Getting package-level metrics...")
        
        package_metrics = []
        
        try:
            # Get list of packages
            result = subprocess.run(['go', 'list', './...'], 
                                  capture_output=True, 
                                  text=True, 
                                  check=True)
            packages = [p.strip() for p in result.stdout.strip().split('\n') if p.strip()]
            
            for package in packages:
                package_name = package.split('/')[-1] if '/' in package else package
                
                # Count Go files in package (use configurable pattern)
                base_pattern = self.config.package_pattern.replace('*', 'tsupplis').replace('github.com/tsupplis/batchexec', '.')
                package_dir = package.replace(package.split('/')[0] + '/' + package.split('/')[1] + '/' + package.split('/')[2], '.') if len(package.split('/')) >= 3 else '.'
                if package_dir == '.' or package_dir == package:
                    package_dir = '.'
                
                go_files = list(Path(package_dir).glob('*.go')) if Path(package_dir).exists() else []
                go_files = [f for f in go_files if not f.name.endswith('_test.go')]
                test_files = list(Path(package_dir).glob('*_test.go')) if Path(package_dir).exists() else []
                
                # Count lines in package
                lines = 0
                test_lines = 0
                
                for file in go_files:
                    try:
                        with file.open('r', encoding='utf-8') as f:
                            lines += len(f.readlines())
                    except:
                        pass
                
                for file in test_files:
                    try:
                        with file.open('r', encoding='utf-8') as f:
                            test_lines += len(f.readlines())
                    except:
                        pass
                
                package_metrics.append({
                    'package': package_name,
                    'full_package': package,
                    'go_files': len(go_files),
                    'test_files': len(test_files),
                    'lines': lines,
                    'test_lines': test_lines
                })
        
        except subprocess.SubprocessError:
            pass
        
        return package_metrics
    
    def generate_report(self) -> bool:
        """Generate the complete code analysis report."""
        print(f"📝 Generating code analysis report: {self.config.output_file}")
        
        # Get all analysis data
        complexity_data = self.get_complexity_data()
        cognitive_complexity = self.get_cognitive_complexity()
        static_analysis = self.get_static_analysis()
        vet_analysis = self.get_vet_analysis()
        staticcheck_analysis = self.get_staticcheck_analysis()
        security_issues = self.get_security_analysis()
        vulnerabilities = self.get_vulnerability_analysis()
        code_smells = self.get_code_smells()
        deadcode_analysis = self.get_deadcode_analysis()
        architecture_violations = self.get_architecture_analysis()
        metrics = self.get_code_metrics()
        
        # Generate report
        report_content = self._build_report(
            complexity_data, 
            cognitive_complexity,
            static_analysis,
            vet_analysis,
            staticcheck_analysis,
            security_issues,
            vulnerabilities,
            code_smells,
            deadcode_analysis,
            architecture_violations,
            metrics
        )
        
        # Write to file
        try:
            output_path = Path(self.config.output_file)
            output_path.parent.mkdir(parents=True, exist_ok=True)
            
            with output_path.open('w', encoding='utf-8') as f:
                f.write(report_content)
            
            print(f"✅ Analysis complete! Report generated: {self.config.output_file}")
            self._print_summary(metrics)
            return True
            
        except IOError as e:
            print(f"❌ Error writing report: {e}")
            return False
    
    def _build_report(self, 
                      complexity_data: List[ComplexityEntry],
                      cognitive_complexity: List[ComplexityEntry], 
                      static_analysis: Dict[str, any],
                      vet_analysis: List[Dict[str, any]],
                      staticcheck_analysis: List[Dict[str, any]],
                      security_issues: List[SecurityIssue],
                      vulnerabilities: List[Dict[str, any]],
                      code_smells: List[CodeSmell],
                      deadcode_analysis: List[Dict[str, any]],
                      architecture_violations: List[ArchitectureViolation],
                      metrics: CodeMetrics) -> str:
        """Build the complete markdown report."""
        
        report_parts = [
            self._build_header(),
            self._build_complexity_section(complexity_data),
            self._build_cognitive_complexity_section(cognitive_complexity),
            self._build_static_analysis_section(static_analysis),
            self._build_vet_section(vet_analysis),
            self._build_staticcheck_section(staticcheck_analysis),
            self._build_security_section(security_issues),
            self._build_vulnerability_section(vulnerabilities),
            self._build_code_smells_section(code_smells),
            self._build_deadcode_section(deadcode_analysis),
            self._build_architecture_section(architecture_violations),
            self._build_metrics_section(metrics),
            self._build_guidelines_section(),
            self._build_footer()
        ]
        
        return '\n'.join(report_parts)
    
    def _build_header(self) -> str:
        """Build the report header."""
        return """# Code Analysis Report

*Generated automatically by `make analyze-code`*

## Overview

This report provides comprehensive code analysis including complexity, security, maintainability metrics, and recommendations for the project.

## Cyclomatic Complexity Analysis

### Top 10 Most Complex Functions

"""
    
    def _build_complexity_section(self, complexity_data: List[ComplexityEntry]) -> str:
        """Build the complexity analysis section."""
        sections = []
        
        # Top 10 functions table
        sections.append("| Complexity | Package | Function | File | Line |")
        sections.append("|------------|---------|----------|------|------|")
        
        top_functions = complexity_data[:self.config.top_functions]
        if top_functions:
            for entry in top_functions:
                complexity_emoji = self._get_complexity_emoji(entry.complexity)
                sections.append(f"| {complexity_emoji} {entry.complexity} | `{entry.package}` | `{entry.function}` | `{entry.file}` | {entry.line} |")
        else:
            sections.append("| - | - | No complexity data available | - | - |")
        
        # High complexity functions
        sections.append("\n### Functions Requiring Attention (Complexity > 15)\n")
        
        high_complexity = [e for e in complexity_data if e.complexity > self.config.complexity_threshold]
        if high_complexity:
            sections.append("| Complexity | Package | Function | File | Line | Priority |")
            sections.append("|------------|---------|----------|------|------|----------|")
            
            for entry in high_complexity:
                complexity_emoji = self._get_complexity_emoji(entry.complexity)
                priority = self._get_priority_level(entry.complexity)
                sections.append(f"| {complexity_emoji} {entry.complexity} | `{entry.package}` | `{entry.function}` | `{entry.file}` | {entry.line} | {priority} |")
        else:
            sections.append("✅ **No high-complexity functions found** - All functions are below the complexity threshold!")
        
        return '\n'.join(sections)
    
    def _build_static_analysis_section(self, static_analysis: Dict[str, any]) -> str:
        """Build the static analysis section."""
        sections = ["\n## Static Analysis Results\n"]
        
        if static_analysis['error']:
            sections.append("### Configuration Issues\n")
            sections.append("⚠️ **Warning**: There were issues running static analysis:")
            sections.append(f"```\n{static_analysis['error']}\n```\n")
        
        if static_analysis['issues']:
            sections.append("### Code Quality Issues\n")
            sections.append("| File | Line | Column | Issue | Linter |")
            sections.append("|------|------|--------|-------|--------|")
            
            for issue in static_analysis['issues']:
                file_path = f"`{issue['file']}`"
                sections.append(f"| {file_path} | {issue['line']} | {issue['col']} | {issue['message']} | `{issue['linter']}` |")
        elif static_analysis['success']:
            sections.append("✅ **No code quality issues found!**")
        
        return '\n'.join(sections)
    
    def _build_metrics_section(self, metrics: CodeMetrics) -> str:
        """Build the code metrics section."""
        sections = [
            "\n## Code Metrics\n",
            "### Project Overview\n",
            f"- **Total Lines of Code:** {metrics.total_lines:,}",
            f"- **Go Files:** {metrics.go_files}",
            f"- **Packages:** {metrics.packages}",
            "\n### Package Details\n"
        ]
        
        if metrics.package_metrics:
            sections.extend([
                "| Package | Go Files | Test Files | Lines | Test Lines | Test Coverage |",
                "|---------|----------|------------|-------|------------|---------------|"
            ])
            
            # Create a coverage lookup
            coverage_map = {}
            for entry in metrics.coverage_entries:
                package_name = entry.package.split('/')[-1] if '/' in entry.package else entry.package
                coverage_map[package_name] = entry.coverage
            
            total_lines = 0
            total_test_lines = 0
            total_go_files = 0
            total_test_files = 0
            
            for pkg in metrics.package_metrics:
                coverage = coverage_map.get(pkg['package'], 'N/A')
                if coverage == 'N/A':
                    # Try with full package name
                    for entry in metrics.coverage_entries:
                        if entry.package == pkg['full_package']:
                            coverage = entry.coverage
                            break
                
                sections.append(f"| `{pkg['package']}` | {pkg['go_files']} | {pkg['test_files']} | {pkg['lines']:,} | {pkg['test_lines']:,} | {coverage} |")
                
                total_lines += pkg['lines']
                total_test_lines += pkg['test_lines']
                total_go_files += pkg['go_files']
                total_test_files += pkg['test_files']
            
            # Add totals row
            sections.append(f"| **TOTAL** | **{total_go_files}** | **{total_test_files}** | **{total_lines:,}** | **{total_test_lines:,}** | - |")
        else:
            sections.append("No package metrics available")
        
        # Add summary coverage table
        sections.extend([
            "\n### Test Coverage Summary\n",
            "| Package | Status | Coverage |",
            "|---------|--------|----------|"
        ])
        
        if metrics.coverage_entries:
            for entry in metrics.coverage_entries:
                package_short = entry.package.split('/')[-1] if '/' in entry.package else entry.package
                status_icon = "✅" if entry.status.startswith('ok') else "❌"
                sections.append(f"| `{package_short}` | {status_icon} {entry.status} | {entry.coverage} |")
        else:
            sections.append("| - | ❌ Error | Unable to run tests |")
        
        return '\n'.join(sections)
    
    def _build_cognitive_complexity_section(self, cognitive_complexity: List[ComplexityEntry]) -> str:
        """Build the cognitive complexity section."""
        sections = [
            "## Cognitive Complexity Analysis",
            "",
            "### Top 10 Most Cognitively Complex Functions",
            "",
        ]
        
        if cognitive_complexity:
            sections.extend([
                "| Complexity | Package | Function | File | Line |",
                "|------------|---------|----------|------|------|"
            ])
            
            for entry in cognitive_complexity[:10]:
                icon = "🔴" if entry.complexity > 25 else "🔶" if entry.complexity > 15 else "⚠️" if entry.complexity > 10 else "✅"
                sections.append(
                    f"| {icon} {entry.complexity} | `{entry.package}` | `{entry.function}` | `{entry.file}` | {entry.line} |"
                )
        else:
            sections.append("✅ **No complex functions found** - All functions have low cognitive complexity!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_security_section(self, security_issues: List[SecurityIssue]) -> str:
        """Build the security analysis section."""
        sections = [
            "## Security Analysis Results",
            "",
        ]
        
        if security_issues:
            # Group by severity
            high_issues = [i for i in security_issues if i.severity.upper() == 'HIGH']
            medium_issues = [i for i in security_issues if i.severity.upper() == 'MEDIUM']
            low_issues = [i for i in security_issues if i.severity.upper() == 'LOW']
            
            sections.extend([
                f"### Security Issues Found: {len(security_issues)}",
                "",
                "| Severity | Rule | File | Line | Details |",
                "|----------|------|------|------|---------|"
            ])
            
            for issue in high_issues + medium_issues + low_issues:
                severity_icon = "🔴" if issue.severity.upper() == 'HIGH' else "🔶" if issue.severity.upper() == 'MEDIUM' else "🟡"
                sections.append(
                    f"| {severity_icon} {issue.severity} | `{issue.rule}` | `{issue.file}` | {issue.line} | {issue.details[:100]}{'...' if len(issue.details) > 100 else ''} |"
                )
        else:
            sections.append("✅ **No security issues found** - Great job maintaining secure code!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_code_smells_section(self, code_smells: List[CodeSmell]) -> str:
        """Build the code smells section."""
        sections = [
            "## Code Quality Issues",
            "",
        ]
        
        if code_smells:
            sections.extend([
                f"### Code Smells Found: {len(code_smells)}",
                "",
                "| Type | File | Line | Issue |",
                "|------|------|------|-------|"
            ])
            
            for smell in code_smells[:20]:  # Limit to top 20
                sections.append(
                    f"| 👃 {smell.type.replace('_', ' ').title()} | `{smell.file}` | {smell.line} | {smell.message} |"
                )
            
            if len(code_smells) > 20:
                sections.append(f"*... and {len(code_smells) - 20} more issues*")
        else:
            sections.append("✅ **No code smells detected** - Code is clean and well-structured!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_architecture_section(self, violations: List[ArchitectureViolation]) -> str:
        """Build the architecture analysis section."""
        sections = [
            "## Architecture Analysis",
            "",
        ]
        
        if violations:
            sections.extend([
                f"### Architecture Violations Found: {len(violations)}",
                "",
                "| Type | Details |",
                "|------|---------|"
            ])
            
            for violation in violations:
                sections.append(
                    f"| 🏗️ {violation.violation_type.replace('_', ' ').title()} | {violation.details} |"
                )
        else:
            sections.append("✅ **Clean Architecture Maintained** - No dependency violations detected!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_vet_section(self, vet_issues: List[Dict[str, any]]) -> str:
        """Build the go vet analysis section."""
        sections = [
            "## Go Vet Analysis",
            "",
        ]
        
        if vet_issues:
            sections.extend([
                f"### Go Vet Issues Found: {len(vet_issues)}",
                "",
                "| File | Line | Column | Issue |",
                "|------|------|--------|-------|"
            ])
            
            for issue in vet_issues:
                sections.append(
                    f"| `{issue['file']}` | {issue['line']} | {issue['col']} | {issue['message']} |"
                )
        else:
            sections.append("✅ **No go vet issues found** - Code passes all built-in static checks!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_staticcheck_section(self, staticcheck_issues: List[Dict[str, any]]) -> str:
        """Build the staticcheck analysis section."""
        sections = [
            "## Staticcheck Analysis",
            "",
        ]
        
        if staticcheck_issues:
            sections.extend([
                f"### Staticcheck Issues Found: {len(staticcheck_issues)}",
                "",
                "| File | Line | Column | Issue | Rule |",
                "|------|------|--------|-------|------|"
            ])
            
            for issue in staticcheck_issues:
                sections.append(
                    f"| `{issue['file']}` | {issue['line']} | {issue['col']} | {issue['message']} | `{issue['rule']}` |"
                )
        else:
            sections.append("✅ **No staticcheck issues found** - Code meets advanced static analysis standards!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_deadcode_section(self, deadcode_items: List[Dict[str, any]]) -> str:
        """Build the dead code analysis section."""
        sections = [
            "## Dead Code Analysis",
            "",
        ]
        
        if deadcode_items:
            sections.extend([
                f"### Unused Code Found: {len(deadcode_items)}",
                "",
                "| File | Line | Column | Unused Item |",
                "|------|------|--------|-------------|"
            ])
            
            for item in deadcode_items:
                sections.append(
                    f"| `{item['file']}` | {item['line']} | {item['col']} | {item['message']} |"
                )
        else:
            sections.append("✅ **No unused code found** - All functions are properly utilized!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_vulnerability_section(self, vulnerabilities: List[Dict[str, any]]) -> str:
        """Build the vulnerability analysis section."""
        sections = [
            "## Vulnerability Analysis (govulncheck)",
            "",
        ]
        
        if vulnerabilities:
            sections.extend([
                f"### Known Vulnerabilities Found: {len(vulnerabilities)}",
                "",
                "| Type | Details |",
                "|------|---------|"
            ])
            
            for vuln in vulnerabilities:
                sections.append(
                    f"| 🚨 {vuln['type'].title()} | {vuln['message']} |"
                )
        else:
            sections.append("✅ **No known vulnerabilities found** - Dependencies are secure!")
        
        sections.extend(["", ""])
        return '\n'.join(sections)
    
    def _build_guidelines_section(self) -> str:
        """Build the complexity guidelines section."""
        return """
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
"""
    
    def _build_footer(self) -> str:
        """Build the report footer."""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        return f"""
**Generated:** {timestamp}

---

*To regenerate this report, run: `make analyze-code`*
*Report location: `{self.config.output_file}`*
"""
    
    def _get_complexity_emoji(self, complexity: int) -> str:
        """Get emoji based on complexity level."""
        if complexity <= 10:
            return "✅"
        elif complexity <= 15:
            return "⚠️"
        elif complexity <= 25:
            return "🔶"
        else:
            return "🔴"
    
    def _get_priority_level(self, complexity: int) -> str:
        """Get priority level based on complexity."""
        if complexity >= 26:
            return "🔴 **Critical**"
        elif complexity >= 21:
            return "🔶 **High**"
        elif complexity >= 16:
            return "⚠️ **Medium**"
        else:
            return "✅ **Low**"

    def _print_summary(self, metrics: CodeMetrics) -> None:
        """Print a summary to console."""
        print("")
        print("📋 Summary:")
        print(f"   - Total Lines: {metrics.total_lines:,}")
        print(f"   - Go Files: {metrics.go_files}")
        print(f"   - Packages: {metrics.packages}")
        print("")
        print("🔍 View the full report:")
        print(f"   cat {self.config.output_file}")


def main():
    """Main entry point."""
    # Initialize configuration and analyzer
    config = AnalysisConfig()
    
    print(f"🔍 {config.project_name} Code Analysis")
    print("=" * (len(config.project_name) + 15))
    
    analyzer = CodeAnalyzer(config)
    
    # Check tools
    print("📋 Checking required tools...")
    if not analyzer.check_required_tools():
        sys.exit(1)
    
    # Generate report
    if not analyzer.generate_report():
        sys.exit(1)


if __name__ == '__main__':
    main()