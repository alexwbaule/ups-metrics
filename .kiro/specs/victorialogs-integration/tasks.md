# Implementation Plan

- [x] 1. Create domain ports and interfaces
  - Define LogWriter interface with WriteLog and Close methods
  - Create LogEntry struct with timestamp, level, message, source, and metadata fields
  - Define LogWriterFactory interface for creating writers based on configuration
  - Create git commit: "feat: add domain ports for log writers and factory interface"
  - _Requirements: 2.1, 4.1_

- [x] 1.1 Write property test for LogWriter interface compliance
  - **Property 2: Writer factory correctness**
  - **Validates: Requirements 1.2**

- [x] 2. Extend configuration model for log type switch
  - Add Type field to Logs struct with "gelf" or "victorialogs" values
  - Add VictoriaLogs struct to device.Config with address, port, username, password, timeout fields
  - Update Logs struct to include VictoriaLogs configuration
  - Add getter methods in config package for log type and VictoriaLogs settings
  - Create git commit: "feat: add log type switch and VictoriaLogs configuration model"
  - _Requirements: 1.1, 5.1_

- [x] 2.1 Write property test for configuration parsing
  - **Property 1: Configuration parsing completeness**
  - **Validates: Requirements 1.1**

- [x] 2.2 Write property test for explicit type requirement
  - **Property 15: Explicit type requirement**
  - **Validates: Requirements 5.1**

- [x] 3. Implement VictoriaLogs writer adapter
  - Create VictoriaLogs writer struct implementing LogWriter interface
  - Implement WriteLog method with HTTP API calls to VictoriaLogs
  - Add JSON formatting for structured logs with required fields
  - Implement authentication and error handling
  - Create git commit: "feat: implement VictoriaLogs writer adapter with HTTP API integration"
  - _Requirements: 3.1, 3.2, 3.4_

- [x] 3.1 Write property test for JSON structure
  - **Property 6: JSON structure completeness**
  - **Validates: Requirements 3.1**

- [x] 3.2 Write property test for notification field preservation
  - **Property 7: Notification field preservation**
  - **Validates: Requirements 3.2**

- [x] 3.3 Write property test for API error resilience
  - **Property 9: API error resilience**
  - **Validates: Requirements 3.4**

- [x] 4. Refactor existing Graylog writer to implement LogWriter interface
  - Modify Gelf struct to implement LogWriter interface
  - Update LogNotifications method to use new WriteLog signature
  - Ensure backwards compatibility with existing functionality
  - Create git commit: "refactor: update Graylog writer to implement LogWriter interface"
  - _Requirements: 2.4, 5.1_

- [x] 4.1 Write property test for field naming consistency
  - **Property 8: Field naming consistency**
  - **Validates: Requirements 3.3**

- [-] 5. Create log writer switch logic
  - Implement switch logic based on log_type configuration field
  - "gelf" → create GraylogWriter, "victorialogs" → create VictoriaLogsWriter
  - Invalid or missing log_type → return explicit error
  - Create git commit: "feat: implement log writer switch based on configuration type"
  - _Requirements: 1.2, 1.4_

- [x] 5.1 Write property test for switch-based destination selection
  - **Property 3: Switch-based destination selection**
  - **Validates: Requirements 1.2**

- [ ] 6. Implement log writer factory with mandatory switch
  - Create factory function that reads log_type and creates the appropriate writer
  - Fail with clear error when log_type is not specified (no defaults)
  - Fail with clear error when log_type is invalid or destination config is missing
  - Add configuration validation and connection testing
  - Create git commit: "feat: implement log writer factory with mandatory type switch"
  - _Requirements: 1.4, 4.5, 5.1_

- [ ] 6.1 Write property test for configuration validation
  - **Property 4: Configuration validation**
  - **Validates: Requirements 1.4**

- [ ] 6.2 Write property test for initialization validation
  - **Property 14: Initialization validation**
  - **Validates: Requirements 4.5**

- [ ] 7. Update notification service to use new log writer architecture
  - Modify GetNotification struct to accept LogWriter interface via dependency injection
  - Update notification processing to use WriteLog method instead of direct Graylog calls
  - Ensure proper error handling and context cancellation support
  - Create git commit: "refactor: update notification service to use dependency injection for log writers"
  - _Requirements: 2.3, 4.2, 4.3_

- [ ] 7.1 Write property test for context cancellation
  - **Property 11: Context cancellation responsiveness**
  - **Validates: Requirements 4.2**

- [ ] 7.2 Write property test for error propagation
  - **Property 12: Error propagation correctness**
  - **Validates: Requirements 4.3**

- [ ] 8. Update main application to wire new log writer dependencies
  - Modify main.go to use log writer factory
  - Update notification service initialization with proper dependency injection
  - Ensure graceful shutdown of log writers
  - Create git commit: "feat: integrate new log writer architecture in main application"
  - _Requirements: 1.2, 2.3_

- [ ] 8.1 Write property test for service continuity
  - **Property 10: Service continuity**
  - **Validates: Requirements 3.5**

- [ ] 9. Add comprehensive error handling and resilience
  - Implement retry logic with exponential backoff for VictoriaLogs
  - Add circuit breaker pattern for failing destinations
  - Enhance error logging and monitoring capabilities
  - Create git commit: "feat: add comprehensive error handling and resilience patterns"
  - _Requirements: 1.5, 3.5_

- [ ] 9.1 Write property test for graceful error handling
  - **Property 5: Graceful error handling**
  - **Validates: Requirements 1.5**

- [ ] 9.2 Write property test for validation error clarity
  - **Property 19: Validation error clarity**
  - **Validates: Requirements 5.5**

- [ ] 9.3 Write property test for explicit error on missing configuration
  - **Property 20: Explicit error for missing configuration**
  - **Validates: Requirements 1.4, 1.5**

- [ ] 10. Update configuration sample and documentation
  - Add log type switch and VictoriaLogs configuration example to config.sample.yaml
  - Document the explicit switch mechanism: logs.type = "gelf" | "victorialogs"
  - Update README with VictoriaLogs integration instructions and migration guide
  - Create git commit: "docs: add log type switch and VictoriaLogs configuration examples"
  - _Requirements: 1.1, 5.4_

- [ ] 10.1 Write property test for type-based configuration validation
  - **Property 16: Type-based configuration validation**
  - **Validates: Requirements 5.4**

- [ ] 10.2 Write property test for switch-based configuration handling
  - **Property 17: Switch-based configuration handling**
  - **Validates: Requirements 5.2**

- [ ] 10.3 Write property test for deployment compatibility
  - **Property 18: Deployment compatibility**
  - **Validates: Requirements 5.4**

- [ ] 11. Final integration testing and validation
  - Test complete log flow from notification to VictoriaLogs
  - Validate writer selection logic with different configuration scenarios
  - Verify explicit error handling when log_type is missing or invalid
  - Ensure all tests pass, ask the user if questions arise
  - Create git commit: "test: add integration tests and validate complete log flow"
  - _Requirements: 5.1, 5.2, 5.4_