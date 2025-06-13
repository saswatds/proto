.PHONY: release-minor release-patch

# Get current version from version.go
CURRENT_VERSION := $(shell grep "const Version" pkg/version/version.go | cut -d'"' -f2)
VERSION_PARTS := $(subst ., ,$(patsubst v%,%,$(CURRENT_VERSION)))
MAJOR := $(word 1,$(VERSION_PARTS))
MINOR := $(word 2,$(VERSION_PARTS))
PATCH := $(word 3,$(VERSION_PARTS))

# Get conventional commit messages since last tag
get-commits:
	@echo "Fetching conventional commits since last tag..."
	@git log $(shell git describe --tags --abbrev=0 2>/dev/null || echo "HEAD")..HEAD --pretty=format:"%s" > commits.tmp

# Update version in all files
update-version:
	@echo "Updating version to $(NEW_VERSION)..."
	@echo 'package version\n\n// Version is the current version of the application\nconst Version = "$(NEW_VERSION)"' > pkg/version/version.go
	@echo "// $(NEW_VERSION)" > go.mod.tmp && cat go.mod | grep -v "^// v" >> go.mod.tmp && mv go.mod.tmp go.mod
	@sed -i.bak 's/go install github.com\/saswatds\/proto\/cmd\/proto@v.*/go install github.com\/saswatds\/proto\/cmd\/proto@$(NEW_VERSION)/' README.md && rm README.md.bak
	@sed -i.bak 's/The current version is v.*/The current version is $(NEW_VERSION)./' README.md && rm README.md.bak
	@echo "## [$(NEW_VERSION)] - $$(date +%Y-%m-%d)" > CHANGELOG.md.tmp
	@echo "" >> CHANGELOG.md.tmp
	@if [ -f commits.tmp ]; then \
		if grep -q -i "^feat" commits.tmp; then \
			echo "### Added" >> CHANGELOG.md.tmp; \
			grep -i "^feat" commits.tmp | sed 's/^feat:/-/' >> CHANGELOG.md.tmp; \
			echo "" >> CHANGELOG.md.tmp; \
		fi; \
		if grep -q -i "^refactor\|^perf" commits.tmp; then \
			echo "### Changed" >> CHANGELOG.md.tmp; \
			grep -i "^refactor\|^perf" commits.tmp | sed 's/^refactor:/-/' | sed 's/^perf:/-/' >> CHANGELOG.md.tmp; \
			echo "" >> CHANGELOG.md.tmp; \
		fi; \
		if grep -q -i "^fix" commits.tmp; then \
			echo "### Fixed" >> CHANGELOG.md.tmp; \
			grep -i "^fix" commits.tmp | sed 's/^fix:/-/' >> CHANGELOG.md.tmp; \
			echo "" >> CHANGELOG.md.tmp; \
		fi; \
		if grep -q -i "^security" commits.tmp; then \
			echo "### Security" >> CHANGELOG.md.tmp; \
			grep -i "^security" commits.tmp | sed 's/^security:/-/' >> CHANGELOG.md.tmp; \
			echo "" >> CHANGELOG.md.tmp; \
		fi; \
		if grep -q -i "^chore\|^docs\|^style\|^test\|^ci" commits.tmp; then \
			echo "### Chore" >> CHANGELOG.md.tmp; \
			grep -i "^chore\|^docs\|^style\|^test\|^ci" commits.tmp | sed 's/^chore:/-/' | sed 's/^docs:/-/' | sed 's/^style:/-/' | sed 's/^test:/-/' | sed 's/^ci:/-/' >> CHANGELOG.md.tmp; \
			echo "" >> CHANGELOG.md.tmp; \
		fi; \
	fi
	@cat CHANGELOG.md >> CHANGELOG.md.tmp
	@mv CHANGELOG.md.tmp CHANGELOG.md
	@rm -f commits.tmp
	@echo "Version updated to $(NEW_VERSION)"

# Push changes and tags
push-release:
	@echo "Pushing changes and tags..."
	@git push
	@git push --tags
	@git tag -f latest
	@git push -f origin latest
	@echo "Changes and tags pushed successfully"

# Create a minor release (e.g., v0.3.0 -> v0.4.0)
release-minor:
	@NEW_VERSION="v$(MAJOR).$$(( $(MINOR) + 1 )).0" $(MAKE) get-commits
	@NEW_VERSION="v$(MAJOR).$$(( $(MINOR) + 1 )).0" $(MAKE) update-version
	@git add pkg/version/version.go go.mod README.md CHANGELOG.md
	@git commit -m "chore: release $(NEW_VERSION)"
	@NEW_VERSION="v$(MAJOR).$$(( $(MINOR) + 1 )).0" && git tag -a "$$NEW_VERSION" -m "Release $$NEW_VERSION"
	@echo "Created minor release $(NEW_VERSION)"
	@$(MAKE) push-release

# Create a patch release (e.g., v0.4.0 -> v0.4.1)
release-patch:
	@NEW_VERSION="v$(MAJOR).$(MINOR).$$(( $(PATCH) + 1 ))" $(MAKE) get-commits
	@NEW_VERSION="v$(MAJOR).$(MINOR).$$(( $(PATCH) + 1 ))" $(MAKE) update-version
	@git add pkg/version/version.go go.mod README.md CHANGELOG.md
	@git commit -m "chore: release $(NEW_VERSION)"
	@NEW_VERSION="v$(MAJOR).$(MINOR).$$(( $(PATCH) + 1 ))" && git tag -a "$$NEW_VERSION" -m "Release $$NEW_VERSION"
	@echo "Created patch release $(NEW_VERSION)"
	@$(MAKE) push-release

# Show current version
version:
	@echo "Current version: $(CURRENT_VERSION)"
