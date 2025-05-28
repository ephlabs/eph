#!/bin/bash
set -e

echo "🔧 Setting up CI tools..."

# Install golangci-lint
echo "📦 Installing golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install pre-commit
echo "📦 Installing pre-commit..."
pip install pre-commit || pip3 install pre-commit

# Install pre-commit hooks
echo "🔗 Setting up pre-commit hooks..."
pre-commit install

# Run initial checks
echo "🧪 Running initial checks..."
pre-commit run --all-files || true

echo "✅ CI tools setup complete!"
echo ""
echo "Pre-commit hooks are now active. They will run automatically before each commit."
echo "To run manually: pre-commit run --all-files"