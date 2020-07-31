baseURL = https://iter8.tools/docs/
archiveVersionNumber = v1.0.0

run-docs: ## Run in development mode
	hugo serve -D

docs: ## Build the site
	hugo -d public --gc --minify --cleanDestinationDir && make copy-archive

docs-baseURL: ## Build the site with the base URL
	hugo -b $(baseURL) -d public --gc --minify --cleanDestinationDir && make copy-archive

docs-archive: ## Build the site with the archival base URL
	hugo -b $(baseURL)archive/$(archiveVersionNumber) -d public --gc --minify --cleanDestinationDir && make copy-archive

copy-archive: ## Copies archive into public
	mkdir public/archive/ && rsync -r archive/ public/archive/