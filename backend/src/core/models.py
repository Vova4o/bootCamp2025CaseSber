from sqlalchemy import Column, String, Text, DateTime, Float, JSON, ForeignKey
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from .database import Base
import uuid

class ChatSession(Base):
    __tablename__ = "chat_sessions"
    
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    mode = Column(String(20), nullable=False, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), index=True)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())
    
    # Убираем lazy="selectinload" - это неправильный синтаксис
    messages = relationship(
        "ChatMessage", 
        back_populates="session", 
        cascade="all, delete-orphan",
        lazy="select"  # Изменили на select
    )
    
    def to_dict(self):
        """Базовый словарь без сообщений"""
        return {
            "id": self.id,
            "mode": self.mode,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "messages": []
        }

class ChatMessage(Base):
    __tablename__ = "chat_messages"
    
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    session_id = Column(String, ForeignKey("chat_sessions.id", ondelete="CASCADE"), nullable=False, index=True)
    role = Column(String(20), nullable=False)
    content = Column(Text, nullable=False)
    sources = Column(JSON, nullable=True)
    reasoning = Column(Text, nullable=True)
    timestamp = Column(DateTime(timezone=True), server_default=func.now(), index=True)
    
    session = relationship("ChatSession", back_populates="messages")
    
    def to_dict(self):
        return {
            "id": self.id,
            "role": self.role,
            "content": self.content,
            "sources": self.sources or [],
            "reasoning": self.reasoning,
            "timestamp": self.timestamp.isoformat() if self.timestamp else None
        }

class SearchHistory(Base):
    __tablename__ = "search_history"
    
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    query = Column(Text, nullable=False, index=True)
    mode = Column(String(20), nullable=False, index=True)
    answer = Column(Text, nullable=False)
    sources = Column(JSON, nullable=True)
    reasoning_steps = Column(JSON, nullable=True)
    response_time = Column(Float, nullable=False)
    session_id = Column(String, nullable=True, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), index=True)
    
    def to_dict(self):
        return {
            "id": self.id,
            "query": self.query,
            "mode": self.mode,
            "answer": self.answer,
            "sources": self.sources or [],
            "reasoning_steps": self.reasoning_steps or [],
            "response_time": self.response_time,
            "session_id": self.session_id,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }