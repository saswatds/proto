.PHONY: release-minor release-patch

# Get current version from version.go
CURRENT_VERSION := $(shell grep "const Version" pkg/version/version.go | cut -d'"' -f2)
VERSION_PARTS := $(subst ., ,$(patsubst v%,%,$(CURRENT_VERSION)))
MAJOR := $(word 1,$(VERSION_PARTS))
MINOR := $(word 2,$(VERSION_PARTS))
PATCH := $(word 3,$(VERSION_PARTS))

# Update version in all files
update-version:
	@echo "Updating version to $(NEW_VERSION)..."
	@echo 'package version\n\n// Version is the current version of the application\nconst Version = "$(NEW_VERSION)"' > pkg/version/version.go
	@echo "// $(NEW_VERSION)" > go.mod.tmp && cat go.mod | grep -v "^// v" >> go.mod.tmp && mv go.mod.tmp go.mod
	@sed -i.bak 's/go install github.com\/saswatds\/proto\/cmd\/proto@v.*/go install github.com\/saswatds\/proto\/cmd\/proto@$(NEW_VERSION)/' README.md && rm README.md.bak
	@sed -i.bak 's/The current version is v.*/The current version is $(NEW_VERSION)./' README.md && rm README.md.bak
	@echo "Version updated to $(NEW_VERSION)"

# Push changes and tags
push-release:
	@echo "Pushing changes and tags..."
	@git push
	@git push --tags
	@echo "Changes and tags pushed successfully"

# Create a minor release (e.g., v0.3.0 -> v0.4.0)
release-minor:
	@NEW_VERSION="v$(MAJOR).$$(( $(MINOR) + 1 )).0" $(MAKE) update-version
	@git add pkg/version/version.go go.mod README.md
	@git commit -m "Release $(NEW_VERSION)"
	@NEW_VERSION="v$(MAJOR).$$(( $(MINOR) + 1 )).0" && git tag -a "$$NEW_VERSION" -m "Release $$NEW_VERSION"
	@echo "Created minor release $(NEW_VERSION)"
	@$(MAKE) push-release

# Create a patch release (e.g., v0.4.0 -> v0.4.1)
release-patch:
	@NEW_VERSION="v$(MAJOR).$(MINOR).$$(( $(PATCH) + 1 ))" $(MAKE) update-version
	@git add pkg/version/version.go go.mod README.md
	@git commit -m "Release $(NEW_VERSION)"
	@NEW_VERSION="v$(MAJOR).$(MINOR).$$(( $(PATCH) + 1 ))" && git tag -a "$$NEW_VERSION" -m "Release $$NEW_VERSION"
	@echo "Created patch release $(NEW_VERSION)"
	@$(MAKE) push-release

# Show current version
version:
	@echo "Current version: $(CURRENT_VERSION)"
