test:
	for test in $(shell find . -name e2etest.sh -print); do \
		echo "\033[0;32mRUNNING:\033[0m $$test"; \
		$$test || exit 1; \
	done
