# Design Document - VictoriaLogs Integration

## Overview

Este design implementa a integração do VictoriaLogs como um novo destino de logs no serviço ups-metrics, seguindo princípios de arquitetura hexagonal. A solução refatora a estrutura atual para usar interfaces bem definidas (ports) e implementações específicas (adapters), permitindo múltiplos destinos de logs simultâneos e facilitando futuras extensões.

## Architecture

### Current Architecture Issues
- Acoplamento direto entre domain service e Graylog
- Falta de interfaces claras para log writers
- Configuração rígida que não suporta múltiplos destinos
- Violação dos princípios de arquitetura hexagonal

### Proposed Hexagonal Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │   Config        │  │    Logger       │                 │
│  └─────────────────┘  └─────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
│  ┌─────────────────────────────────────────────────────────┐│
│  │           Notification Service                          ││
│  │  ┌─────────────────┐  ┌─────────────────┐             ││
│  │  │   LogWriter     │  │  NotificationRepo│             ││
│  │  │   (Port)        │  │    (Port)       │             ││
│  │  └─────────────────┘  └─────────────────┘             ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                       │
│  ┌─────────────────┐     OR     ┌─────────────────┐        │
│  │  GraylogWriter  │ ◄────────► │VictoriaLogsWriter│       │
│  │   (Adapter)     │            │   (Adapter)     │       │
│  └─────────────────┘            └─────────────────┘       │
│                    ┌─────────────┐                        │
│                    │  SMSUpsRepo │                        │
│                    │  (Adapter)  │                        │
│                    └─────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Domain Ports

#### LogWriter Interface
```go
type LogWriter interface {
    WriteLog(ctx context.Context, entry LogEntry) error
    Close() error
}

type LogEntry struct {
    Timestamp    time.Time
    Level        string
    Message      string
    Source       string
    Metadata     map[string]interface{}
}
```

#### LogWriterFactory Interface
```go
type LogWriterFactory interface {
    CreateLogWriters(config Config) ([]LogWriter, error)
}
```

### Infrastructure Adapters

#### VictoriaLogs Writer
- Implementa LogWriter interface
- Envia logs via HTTP API do VictoriaLogs
- Suporta autenticação e retry logic
- Formata logs como JSON estruturado

#### Graylog Writer (Refatorado)
- Refatora implementação atual para seguir LogWriter interface
- Mantém compatibilidade com GELF
- Preserva funcionalidade existente

#### Log Writer Factory
- Cria o writer baseado no switch log_type da configuração
- "gelf" → cria GraylogWriter
- "victorialogs" → cria VictoriaLogsWriter
- Valor inválido ou ausente → retorna erro explícito

## Data Models

### Configuration Extensions
```go
type Logs struct {
    Type         string       `mapstructure:"type"`         // "gelf" or "victorialogs"
    Gelf         Gelf         `mapstructure:"gelf"`
    VictoriaLogs VictoriaLogs `mapstructure:"victorialogs"`
}

type VictoriaLogs struct {
    Address  string        `mapstructure:"address"`
    Port     string        `mapstructure:"port"`
    Username string        `mapstructure:"username"`
    Password string        `mapstructure:"password"`
    Timeout  time.Duration `mapstructure:"timeout"`
}
```

### Log Entry Model
```go
type LogEntry struct {
    Timestamp    time.Time              `json:"timestamp"`
    Level        string                 `json:"level"`
    Message      string                 `json:"message"`
    Source       string                 `json:"source"`
    Metadata     map[string]interface{} `json:"metadata"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Property 1: Configuration parsing completeness
*For any* valid VictoriaLogs configuration containing address, port, and authentication parameters, the configuration parser should successfully extract all parameters without loss
**Validates: Requirements 1.1**

Property 2: Writer factory correctness
*For any* configuration with VictoriaLogs enabled, the factory should create a VictoriaLogs writer instance that implements the LogWriter interface
**Validates: Requirements 1.2**

Property 3: Switch-based destination selection
*For any* configuration with log_type set to "victorialogs", only the VictoriaLogs writer should be created and used for logging
**Validates: Requirements 1.2**

Property 4: Configuration validation
*For any* VictoriaLogs configuration, the validation process should verify connection parameters and return appropriate success or failure results
**Validates: Requirements 1.4**

Property 5: Graceful error handling
*For any* invalid configuration, the system should return descriptive error messages and not crash or enter an undefined state
**Validates: Requirements 1.5**

Property 6: JSON structure completeness
*For any* notification sent to VictoriaLogs, the resulting JSON should contain timestamp, level, message, and metadata fields
**Validates: Requirements 3.1**

Property 7: Notification field preservation
*For any* device notification, the log entry should preserve device information, notification ID, and original timestamp
**Validates: Requirements 3.2**

Property 8: Field naming consistency
*For any* log entry formatted by different writers, the field names should be consistent across all destinations
**Validates: Requirements 3.3**

Property 9: API error resilience
*For any* VictoriaLogs API error response, the writer should handle the error appropriately without crashing the application
**Validates: Requirements 3.4**

Property 10: Service continuity
*For any* VictoriaLogs unavailability scenario, the system should continue processing notifications and log the error condition
**Validates: Requirements 3.5**

Property 11: Context cancellation responsiveness
*For any* context cancellation signal, log writers should terminate gracefully within a reasonable timeout
**Validates: Requirements 4.2**

Property 12: Error propagation correctness
*For any* log writing error condition, the writer should return an appropriate error type that can be handled by the domain service
**Validates: Requirements 4.3**



Property 14: Initialization validation
*For any* log writer initialization, the process should validate configuration and establish necessary connections before returning success
**Validates: Requirements 4.5**

Property 15: Explicit type requirement
*For any* configuration without log_type specified, the system should fail with a clear error message requiring explicit type specification
**Validates: Requirements 5.1**

Property 16: Type-based configuration validation
*For any* configuration with log_type specified, the system should validate that the corresponding destination configuration is present and valid
**Validates: Requirements 5.4**

Property 17: Switch-based configuration handling
*For any* configuration with log_type set to "gelf", only Graylog should be used and VictoriaLogs settings should be ignored
**Validates: Requirements 5.2**

Property 18: Deployment compatibility
*For any* existing deployment configuration file, the updated system should parse and execute without breaking changes
**Validates: Requirements 5.4**

Property 19: Validation error clarity
*For any* invalid VictoriaLogs configuration, the validation error message should clearly indicate the specific configuration problem
**Validates: Requirements 5.5**

Property 20: Explicit error for missing configuration
*For any* configuration with invalid log_type or missing destination configuration, the system should fail with a clear error message indicating the problem
**Validates: Requirements 1.4, 1.5**

## Error Handling

### VictoriaLogs API Errors
- HTTP connection failures: Retry with exponential backoff
- Authentication errors: Log error and disable writer temporarily
- Rate limiting: Implement backoff and queue management
- Invalid JSON responses: Log error and continue operation

### Configuration Errors
- Missing required fields: Provide specific error messages
- Invalid URLs/ports: Validate format and connectivity
- Authentication failures: Test credentials during initialization
- Network connectivity: Validate endpoints are reachable

### Single Writer Error Handling
- Writer failures should be logged and reported appropriately
- Graceful degradation when the configured writer becomes unavailable
- Circuit breaker pattern for the active destination
- Fallback behavior when the primary writer fails

## Testing Strategy

### Unit Testing Approach
- Test individual components in isolation
- Mock external dependencies (HTTP clients, configuration)
- Verify error handling paths
- Test configuration parsing and validation
- Verify log formatting and field mapping

### Property-Based Testing Approach
- Use Go's testing/quick package or a library like gopter
- Generate random configurations to test parsing robustness
- Generate random log entries to verify formatting consistency
- Test concurrent access patterns with multiple goroutines
- Verify error handling with simulated failure conditions
- Test backwards compatibility with existing configuration formats

**Property-based testing requirements:**
- Each property-based test should run a minimum of 100 iterations
- Tests should be tagged with comments referencing design document properties
- Use format: `// Feature: victorialogs-integration, Property X: [property description]`
- Each correctness property must be implemented by a single property-based test
- Configure generators to create realistic test data within valid input domains

**Testing library selection:**
- Primary: `testing/quick` (built-in Go library)
- Alternative: `github.com/leanovate/gopter` for more advanced property testing
- HTTP testing: `net/http/httptest` for mocking VictoriaLogs API
- Configuration testing: In-memory configuration objects

### Integration Testing
- Test complete log flow from notification to destination
- Verify multi-writer scenarios with real backends
- Test configuration loading from actual YAML files
- Validate network connectivity and authentication flows

## Implementation Notes

### Dependency Injection
- Use constructor injection for log writers in domain services
- Factory pattern for creating writers based on configuration
- Interface segregation for testability

### Configuration Management
- Extend existing device.Config struct
- Maintain backwards compatibility with current YAML structure
- Add validation methods for new VictoriaLogs fields

### HTTP Client Management
- Reuse existing HTTP client infrastructure
- Add VictoriaLogs-specific authentication headers
- Implement proper timeout and retry mechanisms

### Logging and Monitoring
- Add metrics for log writer performance
- Monitor success/failure rates per destination
- Include health checks for log writer connectivity

### Configuration Example

```yaml
logs:
  type: "victorialogs"  # OBRIGATÓRIO: "gelf" ou "victorialogs"
  gelf:
    address: graylog
    port: 12201
  victorialogs:
    address: victoria-logs
    port: 9428
    username: admin
    password: secret
    timeout: 30s
```

**Comportamento do Switch:**
- `type: "gelf"` → usa apenas configuração gelf, ignora victorialogs
- `type: "victorialogs"` → usa apenas configuração victorialogs, ignora gelf  
- `type` ausente → **ERRO OBRIGATÓRIO** - deve ser especificado explicitamente
- `type` inválido → erro explícito na inicialização
- Configuração do destino escolhido ausente/inválida → erro explícito