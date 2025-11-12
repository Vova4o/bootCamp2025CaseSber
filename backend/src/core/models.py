from sqlalchemy import Column, String, Text, DateTime, Float, JSON
from sqlalchemy.sql import func
from .database import Base
import uuid

class SearchHistory(Base):
    __tablename__ = "search_history"
    
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    query = Column(Text, nullable=False, index=True)
    mode = Column(String(20), nullable=False, index=True)
    answer = Column(Text, nullable=False)
    sources = Column(JSON, nullable=True)
    reasoning_steps = Column(JSON, nullable=True)
    response_time = Column(Float, nullable=False)
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
            "created_at": self.created_at.isoformat() if self.created_at else None
        }