#!/bin/bash
# Helper script for running tests in Docker containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Run tests in Docker containers for cross-platform testing.

OPTIONS:
    -h, --help          Show this help message
    -p, --platform      Platform to test (linux, macos) [default: linux]
    -t, --test-type     Test type (basic, race, coverage, full) [default: full]
    --no-build          Skip building Docker image
    --clean             Clean Docker resources after running

COMMANDS:
    build               Build Docker test image
    test                Run tests (default command)
    clean               Clean Docker test resources

EXAMPLES:
    $0                                    # Run all tests in Linux container
    $0 -p linux -t full                  # Run all tests in Linux container
    $0 -p linux -t coverage              # Generate coverage reports
    $0 -p macos                          # Run tests on macOS (requires macOS host)
    $0 build                             # Build Docker image only
    $0 clean                             # Clean Docker resources

EOF
}

# Default values
PLATFORM="linux"
TEST_TYPE="full"
NO_BUILD=false
CLEAN_AFTER=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -t|--test-type)
            TEST_TYPE="$2"
            shift 2
            ;;
        --no-build)
            NO_BUILD=true
            shift
            ;;
        --clean)
            CLEAN_AFTER=true
            shift
            ;;
        build)
            COMMAND="build"
            shift
            ;;
        test)
            COMMAND="test"
            shift
            ;;
        clean)
            COMMAND="clean"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Set default command if not specified
COMMAND=${COMMAND:-test}

# Build Docker image
build_image() {
    print_info "Building Docker test image for $PLATFORM..."
    docker build -f Dockerfile.test -t touchlog-test:linux .
    print_info "✓ Docker test image built successfully"
}

# Run tests in Linux container
run_linux_tests() {
    local test_cmd="make test"
    
    case $TEST_TYPE in
        basic)
            test_cmd="make test"
            ;;
        race)
            test_cmd="make test-race"
            ;;
        coverage)
            test_cmd="make test-coverage-xml"
            ;;
        full)
            test_cmd="make test-full"
            ;;
        *)
            print_error "Unknown test type: $TEST_TYPE"
            exit 1
            ;;
    esac
    
    print_info "Running $TEST_TYPE tests in Linux Docker container..."
    print_info "Command: $test_cmd"
    
    mkdir -p coverage
    
    docker run --rm \
        -v "$PROJECT_ROOT:/app" \
        -v "$PROJECT_ROOT/coverage:/app/coverage" \
        -v "$PROJECT_ROOT/coverage.out:/app/coverage.out" \
        -e CGO_ENABLED=1 \
        touchlog-test:linux \
        $test_cmd
    
    print_info "✓ Tests completed. Coverage reports saved to coverage/"
}

# Run tests on macOS (native)
run_macos_tests() {
    if [[ "$(uname)" != "Darwin" ]]; then
        print_error "macOS tests require a macOS host"
        exit 1
    fi
    
    local test_cmd="make test"
    
    case $TEST_TYPE in
        basic)
            test_cmd="make test"
            ;;
        race)
            test_cmd="make test-race"
            ;;
        coverage)
            test_cmd="make test-coverage-xml"
            ;;
        full)
            test_cmd="make test-full"
            ;;
        *)
            print_error "Unknown test type: $TEST_TYPE"
            exit 1
            ;;
    esac
    
    print_info "Running $TEST_TYPE tests natively on macOS..."
    print_info "Command: $test_cmd"
    
    mkdir -p coverage
    $test_cmd
    
    print_info "✓ Tests completed. Coverage reports saved to coverage/"
}

# Clean Docker resources
clean_resources() {
    print_info "Cleaning Docker test resources..."
    docker-compose -f docker-compose.test.yml down --rmi local 2>/dev/null || true
    docker rmi touchlog-test:linux 2>/dev/null || true
    print_info "✓ Docker test resources cleaned"
}

# Main execution
case $COMMAND in
    build)
        build_image
        ;;
    test)
        if [[ "$NO_BUILD" == false ]]; then
            build_image
        fi
        
        case $PLATFORM in
            linux)
                run_linux_tests
                ;;
            macos)
                run_macos_tests
                ;;
            *)
                print_error "Unknown platform: $PLATFORM"
                print_info "Supported platforms: linux, macos"
                exit 1
                ;;
        esac
        
        if [[ "$CLEAN_AFTER" == true ]]; then
            clean_resources
        fi
        ;;
    clean)
        clean_resources
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        usage
        exit 1
        ;;
esac
