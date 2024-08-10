# AutoGPT API

## Overview

AutoGPT API is a powerful and scalable API built with Go and Fiber. It allows seamless interaction with various large language models (LLMs) from top providers like OpenAI, Google, and Anthropic. The API tracks usage and manages costs associated with these services, storing data in MongoDB for detailed analysis and auditing.

## Features

- **Unified API**: Interact with multiple LLM providers (OpenAI, Google, Anthropic) through a single, easy-to-use API.
- **Usage Tracking**: Automatically tracks token usage and associated costs for each user.
- **Cost Management**: Helps in managing and understanding the costs incurred from using different models.
- **Flexible and Extensible**: Easily add new models or providers by updating the `models.json` configuration.
- **Built with Fiber**: High-performance web framework for Go, with built-in middleware and utilities.
- **MongoDB Integration**: Stores all user data, usage history, and cost information in MongoDB for persistence and querying.

## Getting Started

### Prerequisites

- Go 1.19 or later
- MongoDB instance, accessible via a connection string set in the `MONGODB_URI` environment variable
- API keys for OpenAI, Google Gemini, and Anthropic Claude

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/pedcapa/autogpt-api.git
   cd autogpt-api
   ```

2. **Install dependencies:**

   ```bash
   go mod tidy
   ```

3. **Set environment variables:**

   Ensure you have the following environment variables set:
   - `MONGODB_URI`: The connection string for your MongoDB instance.
   - `OPENAI_API_KEY`
   - `GEMINI_API_KEY`
   - `CLAUDE_API_KEY`

   You can set these in your terminal session:

   ```bash
   export MONGODB_URI=your_mongodb_uri
   export OPENAI_API_KEY=your_openai_key
   export GEMINI_API_KEY=your_google_key
   export CLAUDE_API_KEY=your_anthropic_key
   ```

4. **Run the API:**

   ```bash
   go run main.go
   ```

   The API will start on port `8080` by default.

### Usage

The API provides endpoints to interact with language models from different providers:

- **OpenAI**: `/openai`
- **Google**: `/google`
- **Anthropic**: `/anthropic`
- **Whisper**: `/whisper` (under development)
- **Brain**: `/brain` (under development)

Example of a POST request to OpenAI:

```bash
curl -X POST http://localhost:8080/openai \
     -H "Content-Type: application/json" \
     -d '{
           "id_user": "pedro",
           "model": "gpt-4o-mini",
           "messages": [
             {
               "role": "system",
               "content": "You are a helpful assistant."
             },
             {
               "role": "user",
               "content": "Who won the world series in 2020?"
             },
             {
               "role": "assistant",
               "content": "The Los Angeles Dodgers won the World Series in 2020."
             },
             {
               "role": "user",
               "content": "Where was it played?"
             }
           ]
         }'
```

### Contributing

Contributions are currently not available, but we are working on setting up the contribution guidelines and infrastructure. Stay tuned for updates!

### License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

### Contact

For questions or support, please contact [pedro@galliard.mx].
