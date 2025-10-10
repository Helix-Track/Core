#!/usr/bin/env python3
"""
AI QA Test Runner for HelixTrack Core
Tests both SQLite and PostgreSQL configurations
"""

import os
import sys
import time
import requests
import json
from typing import Dict, List, Tuple
from colorama import init, Fore, Style
from datetime import datetime

# Initialize colorama
init(autoreset=True)

# Configuration
SQLITE_API_URL = os.getenv('SQLITE_API_URL', 'http://localhost:8080')
POSTGRES_API_URL = os.getenv('POSTGRES_API_URL', 'http://localhost:8081')
TEST_TIMEOUT = int(os.getenv('TEST_TIMEOUT', '300'))
VERBOSE = os.getenv('VERBOSE', 'true').lower() == 'true'

# Test results
test_results = {
    'total': 0,
    'passed': 0,
    'failed': 0,
    'skipped': 0,
    'tests': []
}

class TestCase:
    """Represents a single test case"""

    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
        self.status = 'pending'
        self.message = ''
        self.duration = 0

    def pass_test(self, message=''):
        self.status = 'passed'
        self.message = message
        test_results['passed'] += 1

    def fail_test(self, message=''):
        self.status = 'failed'
        self.message = message
        test_results['failed'] += 1

    def skip_test(self, message=''):
        self.status = 'skipped'
        self.message = message
        test_results['skipped'] += 1


def log(message: str, level: str = 'INFO'):
    """Log a message with color coding"""
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    if level == 'INFO':
        color = Fore.CYAN
    elif level == 'SUCCESS':
        color = Fore.GREEN
    elif level == 'WARNING':
        color = Fore.YELLOW
    elif level == 'ERROR':
        color = Fore.RED
    else:
        color = Fore.WHITE

    print(f"{color}[{timestamp}] [{level}] {message}{Style.RESET_ALL}")


def wait_for_service(url: str, timeout: int = 60) -> bool:
    """Wait for a service to become healthy"""
    log(f"Waiting for service at {url}", 'INFO')
    start_time = time.time()

    while time.time() - start_time < timeout:
        try:
            response = requests.get(f"{url}/health", timeout=5)
            if response.status_code == 200:
                log(f"Service at {url} is healthy", 'SUCCESS')
                return True
        except requests.exceptions.RequestException:
            pass

        time.sleep(2)

    log(f"Service at {url} did not become healthy within {timeout}s", 'ERROR')
    return False


def test_health_check(api_url: str, db_type: str) -> TestCase:
    """Test: Health check endpoint"""
    test = TestCase(
        f"health_check_{db_type}",
        f"Health check endpoint ({db_type})"
    )

    start_time = time.time()

    try:
        response = requests.get(f"{api_url}/health", timeout=10)
        test.duration = time.time() - start_time

        if response.status_code == 200:
            data = response.json()
            if data.get('status') == 'healthy':
                test.pass_test(f"Health check passed in {test.duration:.2f}s")
            else:
                test.fail_test(f"Unexpected status: {data.get('status')}")
        else:
            test.fail_test(f"HTTP {response.status_code}")

    except Exception as e:
        test.duration = time.time() - start_time
        test.fail_test(f"Exception: {str(e)}")

    return test


def test_service_registration(api_url: str, db_type: str) -> TestCase:
    """Test: Service registration"""
    test = TestCase(
        f"service_registration_{db_type}",
        f"Service registration ({db_type})"
    )

    start_time = time.time()

    try:
        # Admin token for testing
        admin_token = "test-admin-token-with-32-characters-minimum-length"

        # Register a test service
        service_data = {
            "name": f"Test Service {db_type}",
            "type": "authentication",
            "version": "1.0.0",
            "url": "http://test-service:8099",
            "health_check_url": "http://test-service:8099/health",
            "role": "primary",
            "priority": 10,
            "metadata": "{}",
            "admin_token": admin_token
        }

        response = requests.post(
            f"{api_url}/api/services/register",
            json=service_data,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )

        test.duration = time.time() - start_time

        if response.status_code == 201:
            data = response.json()
            if data.get('errorCode') == -1:
                test.pass_test(f"Service registered successfully in {test.duration:.2f}s")
            else:
                test.fail_test(f"Error: {data.get('errorMessage')}")
        else:
            test.fail_test(f"HTTP {response.status_code}: {response.text}")

    except Exception as e:
        test.duration = time.time() - start_time
        test.fail_test(f"Exception: {str(e)}")

    return test


def test_service_discovery(api_url: str, db_type: str) -> TestCase:
    """Test: Service discovery"""
    test = TestCase(
        f"service_discovery_{db_type}",
        f"Service discovery ({db_type})"
    )

    start_time = time.time()

    try:
        discovery_data = {
            "type": "authentication",
            "only_healthy": False
        }

        response = requests.post(
            f"{api_url}/api/services/discover",
            json=discovery_data,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )

        test.duration = time.time() - start_time

        if response.status_code == 200:
            data = response.json()
            if 'services' in data:
                test.pass_test(f"Discovered {data.get('total_count', 0)} services in {test.duration:.2f}s")
            else:
                test.fail_test(f"Invalid response format")
        else:
            test.fail_test(f"HTTP {response.status_code}")

    except Exception as e:
        test.duration = time.time() - start_time
        test.fail_test(f"Exception: {str(e)}")

    return test


def test_service_list(api_url: str, db_type: str) -> TestCase:
    """Test: List all services"""
    test = TestCase(
        f"service_list_{db_type}",
        f"List all services ({db_type})"
    )

    start_time = time.time()

    try:
        response = requests.get(f"{api_url}/api/services/list", timeout=10)
        test.duration = time.time() - start_time

        if response.status_code == 200:
            data = response.json()
            if 'services' in data:
                test.pass_test(f"Listed {len(data.get('services', []))} services in {test.duration:.2f}s")
            else:
                test.fail_test(f"Invalid response format")
        else:
            test.fail_test(f"HTTP {response.status_code}")

    except Exception as e:
        test.duration = time.time() - start_time
        test.fail_test(f"Exception: {str(e)}")

    return test


def test_invalid_registration(api_url: str, db_type: str) -> TestCase:
    """Test: Invalid service registration (security test)"""
    test = TestCase(
        f"invalid_registration_{db_type}",
        f"Invalid service registration rejection ({db_type})"
    )

    start_time = time.time()

    try:
        # Try to register with short admin token
        service_data = {
            "name": "Malicious Service",
            "type": "authentication",
            "version": "1.0.0",
            "url": "http://malicious:9999",
            "health_check_url": "http://malicious:9999/health",
            "role": "primary",
            "priority": 10,
            "metadata": "{}",
            "admin_token": "short"  # Too short!
        }

        response = requests.post(
            f"{api_url}/api/services/register",
            json=service_data,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )

        test.duration = time.time() - start_time

        # Should be rejected
        if response.status_code == 400:
            test.pass_test(f"Invalid registration properly rejected in {test.duration:.2f}s")
        else:
            test.fail_test(f"Expected 400, got HTTP {response.status_code}")

    except Exception as e:
        test.duration = time.time() - start_time
        test.fail_test(f"Exception: {str(e)}")

    return test


def test_concurrent_registrations(api_url: str, db_type: str) -> TestCase:
    """Test: Concurrent service registrations"""
    test = TestCase(
        f"concurrent_registrations_{db_type}",
        f"Concurrent service registrations ({db_type})"
    )

    start_time = time.time()

    try:
        import concurrent.futures
        admin_token = "test-admin-token-with-32-characters-minimum-length"

        def register_service(index):
            service_data = {
                "name": f"Concurrent Service {index} {db_type}",
                "type": "extension",
                "version": "1.0.0",
                "url": f"http://concurrent-{index}:8100",
                "health_check_url": f"http://concurrent-{index}:8100/health",
                "role": "primary",
                "priority": index,
                "metadata": "{}",
                "admin_token": admin_token
            }

            response = requests.post(
                f"{api_url}/api/services/register",
                json=service_data,
                headers={'Content-Type': 'application/json'},
                timeout=10
            )
            return response.status_code == 201

        # Register 5 services concurrently
        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            results = list(executor.map(register_service, range(1, 6)))

        test.duration = time.time() - start_time

        success_count = sum(1 for r in results if r)
        if success_count == 5:
            test.pass_test(f"All 5 concurrent registrations succeeded in {test.duration:.2f}s")
        else:
            test.fail_test(f"Only {success_count}/5 registrations succeeded")

    except Exception as e:
        test.duration = time.time() - start_time
        test.fail_test(f"Exception: {str(e)}")

    return test


def run_test_suite(api_url: str, db_type: str) -> List[TestCase]:
    """Run complete test suite for a database type"""
    log(f"\n{'='*80}", 'INFO')
    log(f"Running test suite for {db_type.upper()}", 'INFO')
    log(f"API URL: {api_url}", 'INFO')
    log(f"{'='*80}\n", 'INFO')

    tests = []

    # Wait for service to be ready
    if not wait_for_service(api_url, timeout=60):
        log(f"Service not ready, skipping tests for {db_type}", 'ERROR')
        return tests

    # Run all tests
    tests.append(test_health_check(api_url, db_type))
    tests.append(test_service_registration(api_url, db_type))
    tests.append(test_service_discovery(api_url, db_type))
    tests.append(test_service_list(api_url, db_type))
    tests.append(test_invalid_registration(api_url, db_type))
    tests.append(test_concurrent_registrations(api_url, db_type))

    # Log results
    for test in tests:
        test_results['total'] += 1
        test_results['tests'].append({
            'name': test.name,
            'description': test.description,
            'status': test.status,
            'message': test.message,
            'duration': test.duration
        })

        if test.status == 'passed':
            log(f"✓ {test.description}: {test.message}", 'SUCCESS')
        elif test.status == 'failed':
            log(f"✗ {test.description}: {test.message}", 'ERROR')
        elif test.status == 'skipped':
            log(f"⊘ {test.description}: {test.message}", 'WARNING')

    return tests


def generate_report():
    """Generate test report"""
    report_path = '/test-results/ai-qa-report.json'

    try:
        os.makedirs(os.path.dirname(report_path), exist_ok=True)

        with open(report_path, 'w') as f:
            json.dump(test_results, f, indent=2)

        log(f"\nTest report saved to: {report_path}", 'SUCCESS')
    except Exception as e:
        log(f"Failed to save report: {e}", 'ERROR')

    # Print summary
    print(f"\n{'='*80}")
    print(f"{Fore.CYAN}TEST SUMMARY{Style.RESET_ALL}")
    print(f"{'='*80}")
    print(f"Total:   {test_results['total']}")
    print(f"{Fore.GREEN}Passed:  {test_results['passed']}{Style.RESET_ALL}")
    print(f"{Fore.RED}Failed:  {test_results['failed']}{Style.RESET_ALL}")
    print(f"{Fore.YELLOW}Skipped: {test_results['skipped']}{Style.RESET_ALL}")
    print(f"{'='*80}\n")

    # Calculate success rate
    if test_results['total'] > 0:
        success_rate = (test_results['passed'] / test_results['total']) * 100
        print(f"Success Rate: {success_rate:.2f}%\n")

        if success_rate == 100:
            print(f"{Fore.GREEN}🎉 ALL TESTS PASSED! 🎉{Style.RESET_ALL}\n")
            return 0
        else:
            print(f"{Fore.RED}❌ SOME TESTS FAILED{Style.RESET_ALL}\n")
            return 1

    return 1


def main():
    """Main test runner"""
    log("HelixTrack Core - AI QA Test Suite", 'INFO')
    log("="*80, 'INFO')

    start_time = time.time()

    # Run tests for SQLite
    sqlite_tests = run_test_suite(SQLITE_API_URL, 'sqlite')

    # Run tests for PostgreSQL
    postgres_tests = run_test_suite(POSTGRES_API_URL, 'postgresql')

    total_duration = time.time() - start_time
    log(f"\nTotal test duration: {total_duration:.2f}s", 'INFO')

    # Generate report
    exit_code = generate_report()

    sys.exit(exit_code)


if __name__ == '__main__':
    main()
