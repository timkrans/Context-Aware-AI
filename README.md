# Context-Aware-AI
- This is a memory retrieval system that allows the user to have its memories encoded and saved for easy retrieval with an optional reasoning model. It also has rag so files can be uploaded and used as well.

## Prerequisites 
- install ollama
- run ```ollama serve```
- in another terminal run ```ollama pull nomic-embed-text```
- Ensure Ollama is running on localhost:11434

## Architecture
- Embedding model: nomic-embed-text (via Ollama)
- Current DB: SQLite (stores text + embeddings)
- Retrieval: 
  - Cosine similarity search
[read about cosine similarity](https://en.wikipedia.org/wiki/Cosine_similarity)
  - Recency weighting using timestamps
  - Final score = 0.8 * cosine_similarity + 0.2 * recency_score
- RAG support: Uploaded files are chunked, embedded, and stored as retrievable memory

## API Routes

- All endpoints are served at `http://localhost:3000`
- Switched to web based so it can be dockerized eventually


### 1. **Create User**
- **POST** `/create-user`
  - Request Body: `{ "username": "string", "password": "string" }`
  - Response: `201 Created` with user details

### 2. **Login**
- **POST** `/login`
  - Request Body: `{ "username": "string", "password": "string" }`
  - Response: `200 OK` with session token

### 3. **Get Tabs**
- **GET** `/tabs`
  - Request Header: `Authorization: Bearer <session_token>`
  - Response: `200 OK` with a list of tabs

### 4. **Create Tab**
- **POST** `/tabs`
  - Request Header: `Authorization: Bearer <session_token>`
  - Request Body: `{ "tab_name": "string" }`
  - Response: `201 Created` with new tab details

### 5. **Chat**
- **POST** `/chat`
  - Request Header: `Authorization: Bearer <session_token>`
  - Request Body: `{ "tab_id": <tab_id>, "message": "string" ,  "reasoning": <boolean>  // Optional}`
  - Response: `200 OK` with AI-generated response

### 6. Upload File
- POST /upload
  - Request Header: `Authorization: Bearer <session_token>`
  - Form Data:
      - `tab_id: <tab_index>`
      - `file: <uploaded_file>`
  - Response: 200 OK with { "status": "indexed" }

### 7. Delete Tab
- DELETE /tabs/:id
  - Request Header: `Authorization: Bearer <session_token>`
  - Path Param: `id = tab index (1 = first tab)`
  - Response: 200 OK with "Tab, memories, and documents deleted successfully"

## Resetting memory
- If for whatever reason you want to reset memory delete the .db file and it will

## Depency checks 
- Check vulnerabilities by running 
```govulncheck ./...```
- If govulncheck not instaled
  ```bash
    go install golang.org/x/vuln/cmd/govulncheck@latest

    echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.zshrc
    source ~/.zshrc

    govulncheck -h
    ```


## Future enhancements
- Change the Db to vector DB for faster query times by the model
- Add metadata filtering (timestamps, tags, importance)
- Implement hybrid search (semantic + keyword)
- Build a docker-compose
- Add agents to this with ability to web browse 