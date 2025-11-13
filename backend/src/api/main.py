from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from typing import Optional, List
import logging

from ..core.config import settings
from ..core.database import get_db, engine, Base
from ..core.models import SearchHistory, ChatSession, ChatMessage
from ..tools.llm_factory import create_llm_client
from ..tools.search_client import SearchClient
from ..agents.simple_mode import process_simple_mode
from ..agents.pro_mode import process_pro_mode
from ..agents.router_agent import RouterAgent
from ..utils.context_manager import ContextManager

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
llm_client = create_llm_client(
    provider=settings.llm_provider,
    api_key=settings.openai_api_key if settings.llm_provider == "openai" else settings.llm_api_key,
    base_url=settings.llm_api_url if settings.llm_provider == "local" else None,
    model=settings.openai_model if settings.llm_provider == "openai" else None
)
search_client = SearchClient(settings.tavily_url)
context_manager = ContextManager(max_messages=10, max_tokens=4000)
router_agent = RouterAgent(llm_client)

# Pydantic Models
class SearchRequest(BaseModel):
    query: str
    mode: str = "auto"  # auto, simple, pro
    session_id: Optional[str] = None
    previous_messages: Optional[List[dict]] = None

class SearchResponse(BaseModel):
    query: str
    mode: str
    answer: str
    sources: List[dict]
    reasoning_steps: Optional[List[str]] = None
    search_queries: Optional[List[str]] = None
    response_time: float
    session_id: Optional[str] = None
    context_used: Optional[bool] = False
    router_decision: Optional[dict] = None  # Информация о выборе режима

class CreateSessionRequest(BaseModel):
    mode: str = "auto"

class SendMessageRequest(BaseModel):
    query: str
    mode: str = "auto"

# Events
@app.on_event("startup")
async def startup():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    logger.info("✓ Database initialized")
    logger.info(f"✓ LLM Provider: {settings.llm_provider}")
    logger.info("✓ Router Agent initialized")

@app.on_event("shutdown")
async def shutdown():
    await llm_client.close()
    await search_client.close()
    logger.info("✓ Connections closed")

# Basic Routes
@app.get("/")
async def root():
    return {
        "service": "Research Pro Mode API",
        "version": "1.0.0",
        "status": "running",
        "llm_provider": settings.llm_provider,
        "features": {
            "auto_routing": True,
            "context_management": True,
            "modes": ["auto", "simple", "pro"]
        }
    }

@app.get("/api/health")
async def health():
    return {
        "status": "ok",
        "llm_provider": settings.llm_provider
    }

# Search Route (без сессии)
@app.post("/api/search", response_model=SearchResponse)
async def search(request: SearchRequest, db: AsyncSession = Depends(get_db)):
    mode = request.mode
    router_decision = None
    
    # AUTO MODE: Используем Router Agent
    if mode == "auto":
        context_exists = bool(request.previous_messages and len(request.previous_messages) > 0)
        
        router_decision = await router_agent.route(
            query=request.query,
            use_llm=settings.use_llm_router,  # Добавим в конфиг
            context_exists=context_exists
        )
        
        mode = router_decision["mode"]
        logger.info(
            f"🤖 Router selected '{mode}' mode with {router_decision['confidence']:.0%} confidence: "
            f"{router_decision['reason']}"
        )
    
    # Simple mode ВСЕГДА без контекста
    if mode == "simple":
        result = await process_simple_mode(
            request.query,
            search_client,
            llm_client,
            settings.max_results_simple
        )
    else:
        # Pro mode может использовать контекст
        context = None
        if request.previous_messages:
            messages_dict = request.previous_messages
            if context_manager.should_use_context(request.query, messages_dict):
                context = context_manager.build_context(messages_dict)
        
        result = await process_pro_mode(
            request.query,
            search_client,
            llm_client,
            settings.max_results_pro,
            context=context,
            previous_messages=request.previous_messages
        )
    
    # Добавляем информацию о решении роутера
    result["router_decision"] = router_decision
    
    # Сохранение в БД
    try:
        history = SearchHistory(
            query=request.query,
            mode=result["mode"],
            answer=result["answer"],
            sources=result.get("sources", []),
            reasoning_steps=result.get("reasoning_steps"),
            response_time=result["response_time"],
            session_id=request.session_id
        )
        db.add(history)
        await db.commit()
    except Exception as e:
        logger.error(f"Failed to save history: {e}")
    
    return result

# Chat Session Routes
@app.post("/api/chat/session")
async def create_chat_session(
    request: CreateSessionRequest,
    db: AsyncSession = Depends(get_db)
):
    """Создание новой чат-сессии"""
    try:
        session = ChatSession(mode=request.mode)
        db.add(session)
        await db.commit()
        await db.refresh(session)
        
        logger.info(f"Created chat session: {session.id} with mode: {request.mode}")
        return session.to_dict()
    except Exception as e:
        logger.error(f"Failed to create session: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/chat/session/{session_id}")
async def get_chat_session(
    session_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Получение истории чат-сессии"""
    try:
        result = await db.execute(
            select(ChatSession).where(ChatSession.id == session_id)
        )
        session = result.scalar_one_or_none()
        
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        
        return session.to_dict()
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get session: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/chat/session/{session_id}/message")
async def send_message(
    session_id: str,
    request: SendMessageRequest,
    db: AsyncSession = Depends(get_db)
):
    """Отправка сообщения в чат-сессию"""
    try:
        # Проверяем существование сессии
        result = await db.execute(
            select(ChatSession).where(ChatSession.id == session_id)
        )
        session = result.scalar_one_or_none()
        
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        
        # Получаем историю сообщений
        messages_result = await db.execute(
            select(ChatMessage)
            .where(ChatMessage.session_id == session_id)
            .order_by(ChatMessage.timestamp)
        )
        previous_messages = messages_result.scalars().all()
        messages_dict = [msg.to_dict() for msg in previous_messages]
        
        # Сохраняем сообщение пользователя
        user_message = ChatMessage(
            session_id=session_id,
            role="user",
            content=request.query
        )
        db.add(user_message)
        await db.flush()
        
        # Определяем режим
        mode = request.mode
        router_decision = None
        
        # AUTO MODE: Используем Router Agent
        if mode == "auto":
            context_exists = len(messages_dict) > 0
            
            router_decision = await router_agent.route(
                query=request.query,
                use_llm=settings.use_llm_router,
                context_exists=context_exists
            )
            
            mode = router_decision["mode"]
            logger.info(
                f"🤖 Session {session_id}: Router selected '{mode}' mode "
                f"({router_decision['confidence']:.0%}) - {router_decision['reason']}"
            )
        
        # Simple Mode - ВСЕГДА без контекста
        if mode == "simple":
            logger.info(f"Processing in Simple Mode (no context) for session {session_id}")
            result = await process_simple_mode(
                request.query,
                search_client,
                llm_client,
                settings.max_results_simple
            )
        else:
            # Pro Mode - с контекстом если нужен
            context = None
            
            if context_manager.should_use_context(request.query, messages_dict):
                context = context_manager.build_context(messages_dict)
                logger.info(f"Using context for Pro Mode in session {session_id}")
            else:
                logger.info(f"No context needed for Pro Mode in session {session_id}")
            
            result = await process_pro_mode(
                request.query,
                search_client,
                llm_client,
                settings.max_results_pro,
                context=context,
                previous_messages=messages_dict
            )
        
        # Сохраняем ответ ассистента
        assistant_message = ChatMessage(
            session_id=session_id,
            role="assistant",
            content=result["answer"],
            sources=result.get("sources", []),
            reasoning="\n".join(result.get("reasoning_steps", [])) if result.get("reasoning_steps") else None
        )
        db.add(assistant_message)
        
        # Обновляем время сессии
        from sqlalchemy.sql import func
        session.updated_at = func.now()
        
        await db.commit()
        
        response = {
            "query": request.query,
            "mode": result["mode"],
            "answer": result["answer"],
            "sources": result.get("sources", []),
            "reasoning": "\n".join(result.get("reasoning_steps", [])) if result.get("reasoning_steps") else None,
            "processing_time": result["response_time"],
            "timestamp": assistant_message.timestamp.isoformat(),
            "session_id": session_id,
            "context_used": result.get("context_used", False)
        }
        
        # Добавляем информацию о решении роутера
        if router_decision:
            response["router_decision"] = router_decision
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to send message: {e}")
        await db.rollback()
        raise HTTPException(status_code=500, detail=str(e))

@app.delete("/api/chat/session/{session_id}")
async def delete_chat_session(
    session_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Удаление чат-сессии"""
    try:
        result = await db.execute(
            select(ChatSession).where(ChatSession.id == session_id)
        )
        session = result.scalar_one_or_none()
        
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        
        await db.delete(session)
        await db.commit()
        
        return {"status": "deleted", "session_id": session_id}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to delete session: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/history")
async def get_history(limit: int = 20, db: AsyncSession = Depends(get_db)):
    """Получение истории поисков"""
    try:
        result = await db.execute(
            select(SearchHistory)
            .order_by(SearchHistory.created_at.desc())
            .limit(limit)
        )
        history = result.scalars().all()
        
        return [item.to_dict() for item in history]
    except Exception as e:
        logger.error(f"Failed to get history: {e}")
        raise HTTPException(status_code=500, detail=str(e))