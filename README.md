# EcoScanAI (Earth Day Edition)

## Demo URLs
- [Demo 1](http://demo-url-1.com)
- [Demo 2](http://demo-url-2.com)

## High-Level Architecture
![](architecture-diagram.png)

## Tech Stack
- Frontend: React
- Backend: Node.js
- Database: MongoDB
- Cloud Services: Gemini, Azure OpenAI

## API Documentation
### POST /api/scan
- **Request Body:**
  ```json
  {
    "image": "base64-image-string",
    "metadata": { "key": "value" }
  }
  ```
- **Response:**
  ```json
  {
    "result": "scan-result"
  }
  ```

## Environment Variables
- **Gemini:**
  - `GEMINI_API_KEY`: Your Gemini API key.
- **Azure OpenAI:**
  - `AZURE_OPENAI_API_KEY`: Your Azure OpenAI API key.

## Local Run Steps
1. Clone the repository.
2. Install dependencies: `npm install`.
3. Start the server: `npm start`.

## Docker Instructions
1. Build the Docker image: `docker build -t ecoscanai .`
2. Run the Docker container: `docker run -p 3000:3000 ecoscanai`

## CI/CD Notes
- Integrate with GitHub Actions for automated testing and deployment.

## Sustainability Angle
EcoScanAI aims to leverage AI to promote sustainability through efficient resource management.

## Roadmap
- **Q2 2026:** Launch MVP.
- **Q3 2026:** Develop additional features based on user feedback.

## License
This project is licensed under the MIT License.

## Author
Credit to [@praveenarjun](https://github.com/praveenarjun) for the development of EcoScanAI.