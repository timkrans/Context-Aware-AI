# Context-Aware-AI
- This is a memory retrieval system that allows the user to have its memories encoded and saved for easy retrieval.

## Prerequisites 
- install ollama
- run ```ollama serve```
- in another terminal run ```ollama pull nomic-embed-text```
- Ensure Ollama is running on localhost:11434

## Architecture
- Embedding model: nomic-embed-text (via Ollama)
- Current DB: SQLite (stores text + embeddings)
- Retrieval: cosine similarity search
[read about cosine similarity](https://en.wikipedia.org/wiki/Cosine_similarity)

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
  - Request Body: `{ "tab_id": <tab_id>, "message": "string" }`
  - Response: `200 OK` with AI-generated response


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