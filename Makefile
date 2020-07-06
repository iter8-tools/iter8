.PHONY: changelog
changelog:
	@sed -n '/$(ver)/,/=====/p' CHANGELOG | grep -v $(ver) | grep -v "====="
