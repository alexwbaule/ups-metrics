# Requirements Document

## Introduction

Este documento especifica os requisitos para integrar o VictoriaLogs como um novo destino de logs no serviço ups-metrics, além de melhorar a arquitetura hexagonal existente. O sistema atualmente envia logs para Graylog via GELF e métricas para InfluxDB/Prometheus. A integração com VictoriaLogs permitirá uma alternativa moderna e eficiente para armazenamento de logs.

## Glossary

- **VictoriaLogs**: Sistema de armazenamento de logs de alta performance da VictoriaMetrics
- **Log Writer**: Interface que define contratos para escrita de logs em diferentes destinos
- **Hexagonal Architecture**: Padrão arquitetural que separa a lógica de negócio dos adapters externos
- **Adapter**: Componente que implementa interfaces de portas para comunicação com recursos externos
- **Port**: Interface que define contratos entre o domínio e os adapters
- **Domain Service**: Serviço que contém lógica de negócio do domínio
- **Resource**: Adapter que implementa comunicação com recursos externos (databases, APIs, etc.)

## Requirements

### Requirement 1

**User Story:** Como um administrador de sistema, eu quero configurar um switch explícito para escolher entre Graylog e VictoriaLogs, para que eu tenha controle total sobre qual destino de logs usar.

#### Acceptance Criteria

1. WHEN the system reads configuration THEN the system SHALL support a log_type switch with values "gelf" or "victorialogs"
2. WHEN log_type is "victorialogs" THEN the system SHALL initialize only the VictoriaLogs writer adapter
3. WHEN log_type is "gelf" THEN the system SHALL initialize only the Graylog writer adapter
4. WHEN log_type is not specified or invalid THEN the system SHALL fail with a clear error message
5. WHEN no log destination is properly configured THEN the system SHALL fail to start with descriptive error messages

### Requirement 2

**User Story:** Como um desenvolvedor, eu quero que a arquitetura siga princípios hexagonais claros, para que o sistema seja mais testável e maintível.

#### Acceptance Criteria

1. WHEN implementing log writers THEN the system SHALL define clear port interfaces for log writing operations
2. WHEN adding new log destinations THEN the system SHALL implement adapters that conform to port interfaces
3. WHEN the domain service processes notifications THEN the system SHALL use dependency injection for log writers
4. WHILE maintaining backwards compatibility, THE system SHALL refactor existing Graylog integration to follow hexagonal patterns
5. WHEN testing log functionality THEN the system SHALL allow easy mocking of log writer dependencies

### Requirement 3

**User Story:** Como um operador do sistema, eu quero que logs sejam enviados para VictoriaLogs no formato JSON estruturado, para que eu possa fazer queries eficientes nos logs.

#### Acceptance Criteria

1. WHEN sending notifications to VictoriaLogs THEN the system SHALL format logs as structured JSON with timestamp, level, message, and metadata
2. WHEN processing device notifications THEN the system SHALL include device information, notification ID, and original timestamp in log entries
3. WHEN log formatting occurs THEN the system SHALL ensure consistent field naming across all log destinations
4. WHILE sending logs, THE system SHALL handle VictoriaLogs API authentication and error responses appropriately
5. WHEN VictoriaLogs is unavailable THEN the system SHALL log errors and continue operation without failing

### Requirement 4

**User Story:** Como um desenvolvedor, eu quero interfaces bem definidas para writers de logs, para que seja fácil adicionar novos destinos de logs no futuro.

#### Acceptance Criteria

1. WHEN defining log writer interfaces THEN the system SHALL create a common contract for all log destinations
2. WHEN implementing log writers THEN the system SHALL support context cancellation for graceful shutdowns
3. WHEN writing logs THEN the system SHALL return appropriate errors for handling by the domain service
4. WHILE processing notifications, THE system SHALL allow configuration of multiple concurrent log writers
5. WHEN initializing log writers THEN the system SHALL validate configuration and establish connections

### Requirement 5

**User Story:** Como um administrador, eu quero que o sistema exija configuração explícita do tipo de log, para que não haja ambiguidade sobre qual destino está sendo usado.

#### Acceptance Criteria

1. WHEN reading configuration without log_type THEN the system SHALL fail with a clear error message requiring explicit type specification
2. WHEN log_type is explicitly set to "gelf" THEN the system SHALL use only Graylog configuration and ignore VictoriaLogs settings
3. WHEN log_type is explicitly set to "victorialogs" THEN the system SHALL use only VictoriaLogs configuration and ignore Graylog settings
4. WHEN log_type is set but corresponding destination configuration is missing THEN the system SHALL fail with descriptive error messages
5. WHEN configuration validation occurs THEN the system SHALL provide clear error messages indicating exactly what configuration is missing or invalid