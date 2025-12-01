.PHONY: lint release-patch release-minor release-major

lint:
	uv run ruff format .
	uv run ruff check --fix .
	uv run mypy --config-file pyproject.toml .

release-patch:
	./release.sh patch

release-minor:
	./release.sh minor

release-major:
	./release.sh major
