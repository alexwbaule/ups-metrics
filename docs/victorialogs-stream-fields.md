# VictoriaLogs Stream Fields - Detecção Automática

Este documento explica como os campos de stream do VictoriaLogs são detectados automaticamente pelo ups-metrics.

## Campos de Stream

O VictoriaLogs usa campos de stream para organizar e filtrar logs eficientemente. O ups-metrics detecta automaticamente os seguintes campos:

### app_name
- **Fonte**: Nome do executável
- **Detecção**: `os.Executable()` ou `os.Args[0]`
- **Exemplo**: `ups-metrics`
- **Propósito**: Identifica a aplicação que gerou o log

### hostname
- **Fonte**: Hostname do sistema
- **Detecção**: `os.Hostname()`
- **Exemplo**: `ups-server.example.com`
- **Propósito**: Identifica o servidor que gerou o log

### remote_ip
- **Fonte**: Interface de rede primária
- **Detecção**: Conexão UDP simulada para 8.8.8.8:80
- **Fallback**: Primeira interface não-loopback
- **Exemplo**: `192.168.1.100`
- **Propósito**: Identifica o IP do servidor que gerou o log

## Exemplo de Log Gerado

```json
{
  "_time": "2024-12-10T20:30:15.123456789Z",
  "level": "info",
  "message": "Notification 1001 on 15/12/2024 14:30:00 with Power failure detected",
  "source": "ups-metrics",
  "_msg": "Notification 1001 on 15/12/2024 14:30:00 with Power failure detected",
  "app_name": "ups-metrics",
  "hostname": "ups-server.example.com",
  "remote_ip": "192.168.1.100",
  "application_name": "ups-metrics",
  "id": 1001,
  "date": "15/12/2024 14:30:00",
  "device_address": "ups-storage"
}
```

## Vantagens da Detecção Automática

1. **Zero Configuração**: Não é necessário configurar manualmente os campos
2. **Consistência**: Os campos são sempre preenchidos corretamente
3. **Manutenibilidade**: Não há risco de configuração incorreta
4. **Portabilidade**: Funciona automaticamente em qualquer ambiente

## Queries no VictoriaLogs

Com os campos de stream detectados automaticamente, você pode fazer queries eficientes:

```bash
# Logs de uma aplicação específica
{app_name="ups-metrics"}

# Logs de um servidor específico
{hostname="ups-server.example.com"}

# Logs de uma rede específica
{remote_ip=~"192.168.1.*"}

# Combinação de filtros
{app_name="ups-metrics",hostname="ups-server.example.com"}
```

## Implementação Técnica

### Detecção do App Name
```go
func detectAppName() string {
    execPath, err := os.Executable()
    if err != nil {
        if len(os.Args) > 0 {
            return filepath.Base(os.Args[0])
        }
        return "ups-metrics" // Fallback
    }
    
    appName := filepath.Base(execPath)
    if strings.HasSuffix(appName, ".exe") {
        appName = strings.TrimSuffix(appName, ".exe")
    }
    
    return appName
}
```

### Detecção do Hostname
```go
func detectHostname() string {
    hostname, err := os.Hostname()
    if err != nil {
        return "unknown-host"
    }
    return hostname
}
```

### Detecção do Remote IP
```go
func detectRemoteIP() string {
    // Tenta conectar para descobrir a interface primária
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return getFirstNonLoopbackIP() // Fallback
    }
    defer conn.Close()
    
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return localAddr.IP.String()
}
```

## Configuração

Não é necessária nenhuma configuração adicional. Os campos são detectados automaticamente quando o tipo de log é definido como `victorialogs`:

```yaml
logs:
    type: "victorialogs"
    victorialogs:
        address: victoria-logs.example.com
        port: 9428
        # Stream fields são detectados automaticamente
```

## Troubleshooting

### App Name Incorreto
- Verifique se o executável tem o nome correto
- Em desenvolvimento, pode aparecer como `main` ou `go-build`

### Hostname Não Resolvido
- Verifique a configuração de DNS do sistema
- O hostname pode ser apenas o nome da máquina sem domínio

### Remote IP Incorreto
- Pode mostrar IP interno em ambientes com NAT
- Em containers, pode mostrar IP da rede interna do container
- Para forçar um IP específico, considere usar variável de ambiente (futura implementação)

## Logs de Exemplo por Ambiente

### Desenvolvimento Local
```json
{
  "app_name": "main",
  "hostname": "developer-laptop",
  "remote_ip": "192.168.1.50"
}
```

### Produção
```json
{
  "app_name": "ups-metrics",
  "hostname": "ups-prod-01.company.com",
  "remote_ip": "10.0.1.100"
}
```

### Container Docker
```json
{
  "app_name": "ups-metrics",
  "hostname": "ups-metrics-pod-abc123",
  "remote_ip": "172.17.0.2"
}
```