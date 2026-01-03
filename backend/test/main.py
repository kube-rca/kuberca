from google import genai
import os

os.environ["GEMINI_API_KEY"] = "AIzaSyDB-DrB1u8GTFUBNRm6YVbAqaMivSkwG04"

client = genai.Client()

result = client.models.embed_content(
        model="gemini-embedding-001",
        contents="What is the meaning of life?",)

print(result.embeddings)