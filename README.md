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


## Resetting memory
- if for whatever reason you want to reset memory delete the .db file and it will

## Future enhancements
- change the Db to vector DB for faster query times by the model
- Add metadata filtering (timestamps, tags, importance)
- Implement hybrid search (semantic + keyword)