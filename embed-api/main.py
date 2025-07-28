from dotenv import load_dotenv
load_dotenv()
from fastapi import FastAPI
from pydantic import BaseModel
from typing import List
from sentence_transformers import SentenceTransformer
from crud import insert_embedding, create_tables  # ✅ Import your DB logic




app = FastAPI()

# Load the MiniLM model once at startup
model = SentenceTransformer("all-MiniLM-L6-v2")

# Ensure DB table exists (safe to call on every boot)
create_tables()  # ✅ Will not recreate if already there

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    embedding: List[float]

@app.post("/embed", response_model=EmbedResponse)
async def embed_text(payload: EmbedRequest):
    print(f"Embedding text: {payload.text}")

    # Generate real embedding (384-dimensional vector)
    vector = model.encode(payload.text).tolist()

    # ✅ Insert into Postgres
    try:
        insert_embedding(payload.text, vector)
    except Exception as e:
        print(f"DB insert error: {e}")

    return {"embedding": vector}
