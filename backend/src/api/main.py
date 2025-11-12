from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Optional, List
import logging

from ..core.config import settings
from ..core.database import get_db, engine, Base
from ..core.models import SearchHistory
from ..tools.llm_client import LLMClient
from ..tools.search_client import SearchClient
from ..agents.simple_mode import process_simple_mode
from ..agents.pro_mode import process_pro_mode

# Logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# App
app = FastAPI(
    title="Research Pro Mode API",
    version="1.0.0",
    description="Умный поисковый ассистент с мультиагентной системой"
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Клиенты (синглтоны)
llm_client = LLMClient(settings.llm_api_url, settings.llm_api_key)
search_client = SearchClient(settings.tavily_url)

# Models
class SearchRequest(BaseModel):
    query: str
    mode: str = "auto"  # auto, simple, pro

class SearchResponse(BaseModel):
    mode: str
    query: str
    answer: str
    sources: List[dict]
    reasoning_steps: Optional[List[str]] = None
    search_queries: Optional[List[str]] = None
    response_time: float

# Events
@app.on_event("startup")
async def startup():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    logger.info("✓ Database initialized")

@app.on_event("shutdown")
async def shutdown():
    await llm_client.close()
    await search_client.close()
    logger.info("✓ Connections closed")

# Routes
@app.get("/")
async def root():
    return {
        "service": "Research Pro Mode API",
        "version": "1.0.0",
        "status": "running"
    }

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/api/search", response_model=SearchResponse)
async def search(request: SearchRequest, db: AsyncSession = Depends(get_db)):
    # Выбор режима
    mode = request.mode
    if mode == "auto":
        # Простая эвристика
        mode = "simple" if len(request.query.split()) <= 6 else "pro"
    
    # Выполнение поиска
    if mode == "simple":
        result = await process_simple_mode(
            request.query,
            search_client,
            llm_client,
            settings.max_results_simple
        )
    else:
        result = await process_pro_mode(
            request.query,
            search_client,
            llm_client,
            settings.max_results_pro
        )
    
    # Сохранение в БД
    try:
        history = SearchHistory(
            query=request.query,
            mode=result["mode"],
            answer=result["answer"],
            sources=result.get("sources", []),
            reasoning_steps=result.get("reasoning_steps"),
            response_time=result["response_time"]
        )
        db.add(history)
        await db.commit()
    except Exception as e:
        logger.error(f"Failed to save history: {e}")
    
    return result

@app.get("/api/history")
async def get_history(limit: int = 20, db: AsyncSession = Depends(get_db)):
    from sqlalchemy import select
    
    result = await db.execute(
        select(SearchHistory)
        .order_by(SearchHistory.created_at.desc())
        .limit(limit)
    )
    history = result.scalars().all()
    
    return [item.to_dict() for item in history]