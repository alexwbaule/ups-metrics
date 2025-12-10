# State Management Solutions

Este documento descreve as diferentes soluções implementadas para gerenciar o estado persistente da aplicação, especificamente o último ID de notificação processado.

## Problema Original

O problema era que salvar o último ID no mesmo arquivo de configuração principal (`config.yaml`) sobrescrevia todas as configurações, perdendo as configurações originais.

## Soluções Implementadas

### 1. Arquivo Separado (Atual - Melhorada)

**Arquivo**: `conf/count.yaml`

A implementação atual já usa um arquivo separado, mas foi melhorada para ser mais robusta:

```go
// Uso simples (compatibilidade com código existente)
err := config.SaveLastIdConfig(notif.LastId())
```

**Vantagens**:
- Não interfere com o arquivo de configuração principal
- Simples de usar
- Compatível com código existente

**Desvantagens**:
- Salva apenas no shutdown (pode perder dados se a aplicação crashar)
- Usa YAML que é mais pesado para dados simples

### 2. State Manager (Nova Implementação)

**Arquivo**: `conf/state.json`

Sistema mais robusto usando JSON com operações atômicas:

```go
// Uso através do Config
err := app.Config.UpdateLastNotificationId(notif.LastId())

// Ou acesso direto ao state manager
lastId := app.Config.GetLastKnowId()
```

**Vantagens**:
- Operações atômicas (escreve em arquivo temporário e renomeia)
- JSON mais leve que YAML
- Thread-safe com mutexes
- Inclui timestamp de atualização
- Melhor tratamento de erros

**Desvantagens**:
- Ainda salva apenas quando chamado explicitamente

### 3. Periodic Saver (Recomendada)

Sistema que salva periodicamente em background:

```go
// Criar o periodic saver
periodicSaver := config.NewPeriodicSaver(app.Config, 30*time.Second)

// Iniciar em goroutine
go periodicSaver.Start(ctx)

// Atualizar ID em memória (muito rápido)
periodicSaver.UpdateLastId(notif.LastId())

// Salva automaticamente a cada 30 segundos se houver mudanças
```

**Vantagens**:
- Salva automaticamente em intervalos regulares
- Não bloqueia o processamento principal
- Salva apenas quando há mudanças (eficiente)
- Salva no shutdown para garantir consistência
- Thread-safe

**Desvantagens**:
- Ligeiramente mais complexo de configurar

## Recomendação de Uso

### Para Máxima Simplicidade
Use a implementação atual melhorada:

```go
g.Go(func() error {
    <-ctx.Done()
    return config.SaveLastIdConfig(notif.LastId())
})
```

### Para Máxima Robustez
Use o Periodic Saver (veja exemplo em `cmd/ups-metrics/main_with_periodic_saver.go.example`):

```go
periodicSaver := config.NewPeriodicSaver(app.Config, 30*time.Second)

g.Go(func() error {
    return periodicSaver.Start(ctx)
})

g.Go(func() error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    lastKnownId := notif.LastId()
    
    for {
        select {
        case <-ctx.Done():
            periodicSaver.UpdateLastId(notif.LastId())
            return periodicSaver.Stop()
        case <-ticker.C:
            currentId := notif.LastId()
            if currentId != lastKnownId {
                periodicSaver.UpdateLastId(currentId)
                lastKnownId = currentId
            }
        }
    }
})
```

## Estrutura dos Arquivos de Estado

### count.yaml (Atual)
```yaml
last: 12345
updated_at: "2024-12-10T20:30:00Z"
```

### state.json (Novo)
```json
{
  "last_notification_id": 12345,
  "updated_at": "2024-12-10T20:30:00Z",
  "version": "1.0"
}
```

## Migração

O código atual continuará funcionando sem modificações. Para migrar para o novo sistema:

1. **Gradual**: Use `app.Config.UpdateLastNotificationId()` em vez de `config.SaveLastIdConfig()`
2. **Completa**: Implemente o Periodic Saver para máxima robustez

Ambos os sistemas podem coexistir durante a migração.