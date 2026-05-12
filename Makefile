test:
	@mkdir -p test_logs && rm -rf ./test_logs/*.log && \
	LOG_FILE="test_logs/test_output_$$(date +%Y-%m-%d_%H-%M-%S).log" && \
	capivara ./... -v --race -cover > "$$LOG_FILE" 2>&1 ; \
	TEST_EXIT_CODE=$$? ; \
	echo "✨ Test execution finished! 🦫   | Check the logs in $$LOG_FILE" ; \