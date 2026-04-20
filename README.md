# EcoScanAI

EcoScanAI is an AI-powered image scanning app that analyzes an uploaded image and returns a **structured JSON response** (e.g., classification/insights) using a Vision-capable AI provider.  
It includes **rate limiting**, **security headers**, **hash-based caching**, and **strict JSON validation** to keep results consistent and production-friendly.

> Built for a competition / hackathon project — add your competition name and details in the section below.

---

## Highlights
- Upload an image and get back **normalized JSON results**
- **Go (Gin) API** with a clean request flow
- **SHA-256 image hashing** for cache keys
- **In-memory TTL cache** to reduce cost and latency
- **Provider selector** (Gemini Vision + optional Azure OpenAI Vision)
- **Strict schema validation** for reliable output
- Basic production middleware: **rate limiter** + **security headers**

---

## Tech Stack
- **Backend:** Go + Gin
- **AI Providers:** Gemini Vision (primary), Azure OpenAI Vision (optional)
- **Caching:** In-memory TTL cache (keyed by image hash)
- **Security:** Security headers middleware
- **Reliability:** Strict JSON schema validation/normalization

---

## API
### `POST /api/scan`
Uploads an image and returns a structured JSON analysis.

**Request**
- `Content-Type: multipart/form-data`
- Form field: `image`

**Response**
- `200 OK` with JSON results (normalized to a strict schema)
- Cached results may be returned when the same image is uploaded again (hash match)

---

## Setup (Local Development)
> Update this section based on your actual project structure (I can tailor it perfectly if you share your folder layout and run commands).

Typical steps:
1. Install Go (latest stable recommended)
2. Set required environment variables for your AI provider(s)
3. Run the server

Example:
```bash
go run .
```

---

## Environment Variables
Add these in your shell or a `.env` file (depending on how your project loads env vars):

### Gemini (example)
- `GEMINI_API_KEY=...`

### Azure OpenAI (optional, example)
- `AZURE_OPENAI_ENDPOINT=...`
- `AZURE_OPENAI_API_KEY=...`
- `AZURE_OPENAI_DEPLOYMENT=...`

> Replace the names above with your repo’s actual variable names if they differ.

---

## Competition / Hackathon
**Competition:** _<Add competition name here>_  
**Year/Date:** _<Add date here>_  
**Role:** _<Solo / Team>_  
**What I built:** _<1–2 strong lines explaining your solution>_

**Problem Statement**
- _<What problem does EcoScanAI solve?>_

**Solution**
- _<How EcoScanAI solves it + why AI vision helps>_

---

## Roadmap
- [ ] Add persistent caching (Redis) for multi-instance deployments
- [ ] Add auth / API keys for public deployments
- [ ] Add CI + tests for schema validation
- [ ] Improve UI/UX for scan results and history

---

## Contributing
Contributions are welcome:
1. Fork the repo
2. Create a feature branch
3. Open a Pull Request

---

## License
License by MIT 
