.PHONY: test-build-flags
test-build-flags: test_build.sh
	@echo "==> Running build test..."	
	@chmod +x test_build.sh	
	@./test_build.sh	
	@echo "==> ✅ Build test completed successfully ✅"

# delete temporary binary file
	@rm -f test_app