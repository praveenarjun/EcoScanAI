# EcoScanAI

## Architecture (Diagram)

```mermaid
flowchart LR
  %% EcoScanAI — Request flow

  U[User] --> UI[Web UI]
  UI -->|POST /api/scan\nmultipart/form-data: image| API[Go Server (Gin)]

  API --> RL[Rate limiter]
  RL --> SEC[Security headers]
  SEC --> SCAN[Scan handler]

  SCAN --> HASH[Compute image hash\nSHA-256]
  HASH --> CHK{Cache hit?}

  CHK -->|Yes| CACHE[(In-memory cache\nTTL)] -->|Return cached JSON| API

  CHK -->|No| PROVIDER[AI provider selector]
  PROVIDER --> GEM[Gemini Vision]
  PROVIDER --> AZ[Azure OpenAI Vision\n(optional)]

  GEM --> VAL[Validate + normalize JSON\n(strict schema)]
  AZ --> VAL

  VAL -->|Store by hash| CACHE
  VAL -->|Return JSON| API
  API -->|200 OK: JSON| UI

  %% Layout styling (supported by Mermaid on GitHub)
  classDef store fill:#f6f8fa,stroke:#8c959f,stroke-width:1px;
  classDef svc fill:#e7f5ff,stroke:#1f6feb,stroke-width:1px;
  classDef ai fill:#fff5e6,stroke:#bf8700,stroke-width:1px;
  classDef edge fill:#ffffff,stroke:#8c959f,stroke-dasharray: 3 3;

  class CACHE store;
  class API,RL,SEC,SCAN,HASH,VAL,PROVIDER svc;
  class GEM,AZ ai;
```
