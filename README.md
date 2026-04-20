# EcoScanAI 🌍♻️  
**AI-powered waste scanning to improve disposal decisions and reduce carbon impact.**

Live Demo (Render): **https://ecoscan-ai-sha-c2ba78b.onrender.com/**

---
<img width="1151" height="776" alt="Screenshot 2026-04-20 at 12 29 16 PM" src="https://github.com/user-attachments/assets/efdfcbbe-7581-403c-b4f0-34125b72f576" />


## Why EcoScanAI?
EcoScanAI helps people quickly understand **what an item is**, **what it’s made of**, and **how to dispose of it responsibly** (Recycle / Compost / Landfill / Hazardous / Reuse).  
It also provides a simple **carbon-saved estimate** and a practical tip to encourage better habits.

Built for **DEV Weekend Challenge: Earth Day Edition (April 2026)**.

---

## Key Features
- Image-based waste/item scanning (upload a photo)
- AI returns **structured JSON**:
  - item name
  - material
  - disposal category
  - estimated carbon saved
  - eco tip
- Caching (by image hash) to reduce repeated AI calls
- Rate limiting to protect the API
- Dockerized + CI pipeline

---

## App Links
- **Web App**: https://ecoscan-ai-sha-c2ba78b.onrender.com/  
- **Health Check**: https://ecoscan-ai-sha-c2ba78b.onrender.com/health  

---

## Architecture (Diagram)
```mermaid
flowchart LR
  U[User] -->|Upload photo| UI[Web UI]
  UI -->|POST /api/scan (multipart image)| API[Go (Gin) API]

  API --> RL[Rate Limiter]
  API --> SEC[Security Headers]
  API --> C{Cache hit?}

  C -->|Yes| R1[Return cached result]
  C -->|No| AI[AI Vision Provider\n(Gemini / Azure OpenAI)]
  AI --> J[Validate JSON response]
  J --> STORE[Store in cache (TTL)]
  STORE --> R2[Return result]
```

---

## API
### `POST /api/scan`
- Upload field: `image`
- Type: `multipart/form-data`

**Example response**
```json
{
  "result": {
    "item": "Plastic bottle",
    "material": "PET plastic",
    "disposal": "Recycle",
    "carbon_save": "50g CO2",
    "tips": "Rinse the bottle and remove the cap before recycling."
  }
}
```

---

## Run Locally (Go)
```bash
git clone https://github.com/praveenarjun/EcoScanAI.git
cd EcoScanAI
go mod download
go run .
```

Open: `http://localhost:8080`

---

## Run with Docker
```bash
docker build -t ecoscanai .
docker run --rm -p 8080:8080 \
  -e AI_PROVIDER=gemini \
  -e GEMINI_API_KEY=YOUR_KEY_HERE \
  ecoscanai
```

---

## Environment Variables (Summary)
### Gemini
- `AI_PROVIDER=gemini`
- `GEMINI_API_KEY=...`
- `GEMINI_MODEL=...` (optional)

### Azure OpenAI (optional)
- `AI_PROVIDER=azure`
- `AZURE_OPENAI_ENDPOINT=...`
- `AZURE_OPENAI_API_KEY=...`
- `AZURE_OPENAI_DEPLOYMENT=...`

---

## CI/CD
GitHub Actions runs:
- `gofmt` formatting check
- `go test ./...`
- `go build`
- Docker image build/push to GHCR on push to the default branch

---

## Screenshots
> Add screenshots here for maximum judging impact:
- `assets/home.png`
- `assets/result.png`
- `assets/mobile.png`

---

## Roadmap
- Add region-based disposal rules (country/city)
- Add scan history + streaks
- Add confidence score + follow-up questions
- Add barcode/label detection

---

## Author
Built by **@praveenarjun**
