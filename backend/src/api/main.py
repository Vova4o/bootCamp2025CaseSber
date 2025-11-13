from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from sqlalchemy.sql import func
from typing import Optional, List
import logging

from ..core.config import settings
from ..core.database import get_db, engine, Base
from ..core.models import SearchHistory, ChatSession, ChatMessage
from ..tools.llm_factory import create_llm_client
from ..tools.duckduckgo_client import DuckDuckGoClient
from ..agents.langgraph_pipeline import ResearchGraph
from ..utils.context_manager import ContextManager

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastAPI application
app = FastAPI(
    title="Research Pro Mode API",
    version="2.0.0",
    description="Multi-agent research assistant with LangGraph pipeline and DuckDuckGo search"
)

# Configure CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.get_cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize singleton clients
llm_client = create_llm_client(
    provider=settings.llm_provider,
    api_key=settings.openai_api_key if settings.llm_provider == "openai" else settings.llm_api_key,
    base_url=settings.llm_api_url if settings.llm_provider == "local" else None,
    model=settings.openai_model if settings.llm_provider == "openai" else None
)
search_client = DuckDuckGoClient(max_workers=3)
context_manager = ContextManager(
    max_messages=settings.max_context_messages,
    max_tokens=settings.max_context_tokens
)

# Initialize LangGraph research pipeline
research_graph = ResearchGraph(
    llm_client=llm_client,
    search_client=search_client,
    context_manager=context_manager,
    settings=settings
)

# Pydantic request/response models
class SearchRequest(BaseModel):
    """Request model for search endpoint"""
    query: str
    mode: str = "auto"
    session_id: Optional[str] = None
    previous_messages: Optional[List[dict]] = None

class SearchResponse(BaseModel):
    """Response model for search results"""
    query: str
    mode: str
    answer: str
    sources: List[dict]
    reasoning_steps: Optional[List[str]] = None
    search_queries: Optional[List[str]] = None
    response_time: float
    session_id: Optional[str] = None
    context_used: Optional[bool] = False
    router_decision: Optional[dict] = None

class CreateSessionRequest(BaseModel):
    """Request model for creating chat session"""
    mode: str = "auto"

class SendMessageRequest(BaseModel):
    """Request model for sending message to session"""
    query: str
    mode: str = "auto"

# Application lifecycle events
@app.on_event("startup")
async def startup():
    """Initialize database and log configuration on startup"""
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    
    logger.info("Database initialized successfully")
    logger.info(f"LLM Provider: {settings.llm_provider}")
    logger.info(f"Search Engine: DuckDuckGo (region: {settings.search_region})")
    logger.info("LangGraph pipeline initialized")

@app.on_event("shutdown")
async def shutdown():
    """Clean up resources on shutdown"""
    await llm_client.close()
    await search_client.close()
    logger.info("All connections closed successfully")

# Basic endpoints
@app.get("/")
async def root():
    """Root endpoint with service information"""
    return {
        "service": "Research Pro Mode API",
        "version": "2.0.0",
        "status": "running",
        "llm_provider": settings.llm_provider,
        "search_engine": "DuckDuckGo",
        "search_region": settings.search_region,
        "pipeline": "LangGraph",
        "features": {
            "auto_routing": True,
            "context_management": True,
            "modes": ["auto", "simple", "pro"],
            "graph_visualization": True
        }
    }

@app.get("/api/health")
async def health():
    """Health check endpoint"""
    return {
        "status": "ok",
        "llm_provider": settings.llm_provider,
        "search_engine": "DuckDuckGo",
        "pipeline": "LangGraph"
    }

@app.get("/api/graph/visualize")
async def visualize_graph():
    """
    Visualize the LangGraph pipeline structure.
    Returns Mermaid diagram markup.
    """
    try:
        mermaid = research_graph.visualize()
        return {
            "format": "mermaid",
            "diagram": mermaid
        }
    except Exception as e:
        logger.error(f"Graph visualization failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))

# Search endpoint
@app.post("/api/search", response_model=SearchResponse)
async def search(request: SearchRequest, db: AsyncSession = Depends(get_db)):
    """
    Execute search query through LangGraph pipeline.
    
    Pipeline stages:
        1. Router determines query complexity
        2. Simple or Pro search based on routing
        3. Analysis and answer synthesis
    
    Returns enriched search results with reasoning steps.
    """
    logger.info(f"Search request: '{request.query}' (mode: {request.mode})")
    
    try:
        # Execute LangGraph pipeline
        result = await research_graph.run(
            query=request.query,
            previous_messages=request.previous_messages,
            mode=request.mode
        )
        
        # Save to database
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
            logger.error(f"Failed to save search history: {e}")
        
        return result
        
    except Exception as e:
        logger.error(f"Search failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))

# Chat session endpoints
@app.post("/api/chat/session")
async def create_chat_session(
    request: CreateSessionRequest,
    db: AsyncSession = Depends(get_db)
):
    """Create new chat session"""
    try:
        session = ChatSession(mode=request.mode)
        db.add(session)
        await db.commit()
        await db.refresh(session)
        
        logger.info(f"Created chat session: {session.id} (mode: {request.mode})")
        
        return {
            "id": session.id,
            "mode": session.mode,
            "created_at": session.created_at.isoformat() if session.created_at else None,
            "updated_at": session.updated_at.isoformat() if session.updated_at else None,
            "messages": []
        }
    except Exception as e:
        logger.error(f"Failed to create session: {e}")
        await db.rollback()
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/chat/session/{session_id}")
async def get_chat_session(
    session_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Retrieve chat session with message history"""
    try:
        result = await db.execute(
            select(ChatSession)
            .options(selectinload(ChatSession.messages))
            .where(ChatSession.id == session_id)
        )
        session = result.scalar_one_or_none()
        
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        
        return {
            "id": session.id,
            "mode": session.mode,
            "created_at": session.created_at.isoformat() if session.created_at else None,
            "updated_at": session.updated_at.isoformat() if session.updated_at else None,
            "messages": [msg.to_dict() for msg in session.messages]
        }
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to retrieve session: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/chat/session/{session_id}/message")
async def send_message(
    session_id: str,
    request: SendMessageRequest,
    db: AsyncSession = Depends(get_db)
):
    """
    Send message to chat session and get response.
    Uses LangGraph pipeline with conversation context.
    """
    try:
        # Verify session exists
        result = await db.execute(
            select(ChatSession).where(ChatSession.id == session_id)
        )
        session = result.scalar_one_or_none()
        
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        
        # Retrieve message history
        messages_result = await db.execute(
            select(ChatMessage)
            .where(ChatMessage.session_id == session_id)
            .order_by(ChatMessage.timestamp)
        )
        previous_messages = messages_result.scalars().all()
        messages_dict = [msg.to_dict() for msg in previous_messages]
        
        # Save user message
        user_message = ChatMessage(
            session_id=session_id,
            role="user",
            content=request.query
        )
        db.add(user_message)
        await db.flush()
        
        logger.info(f"Processing message in session {session_id}")
        
        # Execute pipeline with context
        result = await research_graph.run(
            query=request.query,
            previous_messages=messages_dict,
            mode=request.mode
        )
        
        # Save assistant message
        assistant_message = ChatMessage(
            session_id=session_id,
            role="assistant",
            content=result["answer"],
            sources=result.get("sources", []),
            reasoning="\n".join(result.get("reasoning_steps", [])) if result.get("reasoning_steps") else None
        )
        db.add(assistant_message)
        
        # Update session timestamp
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
            "context_used": result.get("context_used", False),
            "router_decision": result.get("router_decision")
        }
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to process message: {e}")
        await db.rollback()
        raise HTTPException(status_code=500, detail=str(e))

@app.delete("/api/chat/session/{session_id}")
async def delete_chat_session(
    session_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Delete chat session and all associated messages"""
    try:
        result = await db.execute(
            select(ChatSession).where(ChatSession.id == session_id)
        )
        session = result.scalar_one_or_none()
        
        if not session:
            raise HTTPException(status_code=404, detail="Session not found")
        
        await db.delete(session)
        await db.commit()
        
        logger.info(f"Deleted session: {session_id}")
        
        return {"status": "deleted", "session_id": session_id}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to delete session: {e}")
        raise HTTPException(status_code=500, detail=str(e))

# History endpoints
@app.get("/api/history")
async def get_history(limit: int = 20, db: AsyncSession = Depends(get_db)):
    """Retrieve recent search history"""
    try:
        result = await db.execute(
            select(SearchHistory)
            .order_by(SearchHistory.created_at.desc())
            .limit(limit)
        )
        history = result.scalars().all()
        
        return [item.to_dict() for item in history]
    except Exception as e:
        logger.error(f"Failed to retrieve history: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/sessions")
async def get_sessions(limit: int = 20, db: AsyncSession = Depends(get_db)):
    """Retrieve list of all chat sessions"""
    try:
        result = await db.execute(
            select(ChatSession)
            .order_by(ChatSession.updated_at.desc())
            .limit(limit)
        )
        sessions = result.scalars().all()
        
        return [session.to_dict() for session in sessions]
    except Exception as e:
        logger.error(f"Failed to retrieve sessions: {e}")
        raise HTTPException(status_code=500, detail=str(e))