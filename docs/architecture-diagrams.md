# LogGuardian Architecture Diagrams

## How LogGuardian Works - Sequence Diagram

```mermaid
sequenceDiagram
    participant CW as CloudWatch Logs
    participant Config as AWS Config
    participant EB as EventBridge
    participant LG as LogGuardian Lambda
    participant KMS as AWS KMS

    Note over CW,KMS: Scenario 1: New Log Group Created
    CW->>Config: Log group created event
    Config->>EB: Config change notification
    EB->>LG: Scheduled compliance check
    LG->>CW: Check encryption & retention
    alt Log group non-compliant
        LG->>KMS: Apply encryption
        LG->>CW: Set retention policy
    else Log group compliant
        Note over LG: Log compliance status
    end

    Note over Config,LG: Scenario 2: Config Rule Evaluation
    Config->>LG: Batch evaluation request
    loop For each log group
        LG->>CW: Evaluate compliance
        LG->>Config: Return compliance status
    end
```

## Architecture Overview

```mermaid
graph TB
    subgraph "AWS Account"
        CWL[CloudWatch Logs]
        CONFIG[AWS Config Service]
        EB[EventBridge Rules]
        
        subgraph "LogGuardian Stack"
            LAMBDA[LogGuardian Lambda]
            KMS[KMS Key]
            ROLE[IAM Role]
        end
        
        subgraph "Monitoring"
            CW[CloudWatch Dashboard]
            METRICS[CloudWatch Metrics]
        end
    end
    
    CWL --> CONFIG
    CONFIG --> EB
    EB --> LAMBDA
    LAMBDA --> CWL
    LAMBDA --> KMS
    LAMBDA --> CW
    LAMBDA --> METRICS
```

## Deployment Architecture Options

### Option 1: New Infrastructure (Default)
```mermaid
graph LR
    A[LogGuardian Stack] --> B[New KMS Key]
    A --> C[New Config Service]
    A --> D[New Config Rules]
    A --> E[New EventBridge Rules]
    A --> F[Lambda Function]
```

### Option 2: Existing Infrastructure (Enterprise)
```mermaid
graph LR
    A[LogGuardian Stack] --> B[Existing KMS Key]
    A --> C[Existing Config Service]
    A --> D[Existing Config Rules]
    A --> E[Optional EventBridge]
    A --> F[Lambda Function]
```

### Option 3: Manual Invocation Only
```mermaid
graph LR
    A[LogGuardian Stack] --> B[KMS Key]
    A --> C[Config Service]
    A --> D[Config Rules]
    A --> E[Lambda Function Only]
    F[Manual Trigger] --> E
```
