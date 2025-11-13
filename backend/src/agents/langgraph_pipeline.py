from typing import TypedDict, Annotated, Sequence, Literal
from langgraph.graph import StateGraph, END
import logging
import time

logger = logging.getLogger(__name__)


class AgentState(TypedDict):
    """
    State schema for the research agent graph.
    Contains all data that flows through the pipeline.
    """
    query: str
    mode: Literal["simple", "pro"]
    search_results: list
    reasoning_steps: list
    answer: str
    sources: list
    context: str
    confidence: float
    router_decision: dict
    response_time: float
    context_used: bool
    subqueries: list


class ResearchGraph:
    """
    LangGraph-based pipeline for multi-agent research assistant.
    
    Graph structure:
        START -> router -> [simple_search | pro_search] -> analyze -> END
    
    Nodes:
        - router: Determines query complexity and selects mode
        - simple_search: Fast single-query search for simple questions
        - pro_search: Multi-query deep research with reasoning
        - analyze: Synthesizes answer from search results
    """
    
    def __init__(self, llm_client, search_client, context_manager, settings):
        """
        Initialize the research graph with required dependencies.
        
        Args:
            llm_client: LLM client for text generation
            search_client: Search client (DuckDuckGo)
            context_manager: Manages conversation context
            settings: Application settings
        """
        self.llm_client = llm_client
        self.search_client = search_client
        self.context_manager = context_manager
        self.settings = settings
        
        self.graph = self._build_graph()
        self.app = self.graph.compile()
    
    def _build_graph(self) -> StateGraph:
        """
        Build the state graph with nodes and edges.
        
        Returns:
            Compiled StateGraph ready for execution
        """
        workflow = StateGraph(AgentState)
        
        # Add nodes
        workflow.add_node("router", self.router_node)
        workflow.add_node("simple_search", self.simple_search_node)
        workflow.add_node("pro_search", self.pro_search_node)
        workflow.add_node("analyze", self.analyze_node)
        
        # Set entry point
        workflow.set_entry_point("router")
        
        # Conditional routing based on query complexity
        workflow.add_conditional_edges(
            "router",
            self.route_decision,
            {
                "simple": "simple_search",
                "pro": "pro_search"
            }
        )
        
        # Both search paths lead to analysis
        workflow.add_edge("simple_search", "analyze")
        workflow.add_edge("pro_search", "analyze")
        
        # Analysis is terminal node
        workflow.add_edge("analyze", END)
        
        return workflow
    
    async def router_node(self, state: AgentState) -> AgentState:
        """
        Router node: Determines query complexity using heuristics.
        
        Decision criteria:
            - Query length (word count)
            - Presence of conversation context
            - Keyword analysis for complexity markers
        
        Returns:
            Updated state with mode and router decision
        """
        logger.info("Router node: analyzing query complexity")
        
        query = state["query"]
        context_exists = bool(state.get("context"))
        words = query.split()
        query_lower = query.lower()
        
        # Heuristic 1: Query length
        if len(words) <= 4:
            mode = "simple"
            confidence = 0.9
            reason = "Short query (<=4 words)"
        elif len(words) >= 15:
            mode = "pro"
            confidence = 0.85
            reason = "Long complex query (>=15 words)"
        
        # Heuristic 2: Context presence
        elif context_exists:
            mode = "pro"
            confidence = 0.8
            reason = "Conversation context exists"
        
        # Heuristic 3: Complexity keywords
        elif any(kw in query_lower for kw in ["сравни", "compare", "почему", "why", "как работает", "how works"]):
            mode = "pro"
            confidence = 0.75
            reason = "Complexity keywords detected"
        
        # Default: Simple mode
        else:
            mode = "simple"
            confidence = 0.6
            reason = "Default: simple mode for speed"
        
        router_decision = {
            "mode": mode,
            "confidence": confidence,
            "reason": reason
        }
        
        logger.info(f"Router decision: {mode} (confidence: {confidence:.2f})")
        
        return {
            **state,
            "mode": mode,
            "confidence": confidence,
            "router_decision": router_decision,
            "reasoning_steps": [f"Router: {reason}"]
        }
    
    def route_decision(self, state: AgentState) -> str:
        """
        Decision function for conditional edge routing.
        
        Returns:
            Next node name based on selected mode
        """
        return state["mode"]
    
    async def simple_search_node(self, state: AgentState) -> AgentState:
        """
        Simple search node: Single query, fast results.
        
        Process:
            1. Execute single search query
            2. Return top N results
            3. No subquery generation
        
        Returns:
            Updated state with search results
        """
        logger.info("Simple search node: executing single query")
        
        state["reasoning_steps"].append("Simple Mode: fast single query search")
        
        # Execute search
        search_results = await self.search_client.search(
            query=state["query"],
            max_results=self.settings.max_results_simple,
            region=self.settings.search_region
        )
        
        results = search_results.get("results", [])
        state["reasoning_steps"].append(f"Found {len(results)} results")
        
        return {
            **state,
            "search_results": results,
            "sources": results,
            "subqueries": [state["query"]]
        }
    
    async def pro_search_node(self, state: AgentState) -> AgentState:
        """
        Pro search node: Multi-query deep research.
        
        Process:
            1. Generate 2-3 subqueries using LLM
            2. Execute parallel searches for each subquery
            3. Aggregate and deduplicate results
            4. Return top results
        
        Returns:
            Updated state with aggregated search results
        """
        logger.info("Pro search node: multi-query deep research")
        
        state["reasoning_steps"].append("Pro Mode: generating subqueries")
        
        # Generate subqueries using LLM
        messages = [
            {
                "role": "system",
                "content": "Break down the query into 2-3 specific search subqueries. Format: 1. query\\n2. query\\n3. query"
            },
            {
                "role": "user",
                "content": f"Break down this query: {state['query']}"
            }
        ]
        
        try:
            subqueries_text = await self.llm_client.chat_completion(
                messages, temperature=0.5, max_tokens=200
            )
            subqueries = [
                q.strip() 
                for q in subqueries_text.split('\n') 
                if q.strip() and not q.strip().startswith('#')
            ][:3]
        except Exception as e:
            logger.error(f"Subquery generation failed: {e}")
            subqueries = [state["query"]]
        
        if not subqueries:
            subqueries = [state["query"]]
        
        state["reasoning_steps"].append(f"Generated {len(subqueries)} subqueries")
        
        # Execute multiple searches
        all_results = []
        for subquery in subqueries:
            try:
                search_results = await self.search_client.search(
                    query=subquery,
                    max_results=self.settings.max_results_pro,
                    region=self.settings.search_region
                )
                all_results.extend(search_results.get("results", []))
            except Exception as e:
                logger.error(f"Search failed for subquery '{subquery}': {e}")
        
        # Deduplicate by URL
        seen_urls = set()
        unique_results = []
        for result in all_results:
            url = result.get("url")
            if url and url not in seen_urls:
                seen_urls.add(url)
                unique_results.append(result)
        
        state["reasoning_steps"].append(f"Aggregated {len(unique_results)} unique results")
        
        return {
            **state,
            "search_results": unique_results[:10],
            "sources": unique_results[:5],
            "subqueries": subqueries
        }
    
    async def analyze_node(self, state: AgentState) -> AgentState:
        """
        Analysis node: Synthesize final answer from search results.
        
        Process:
            1. Format search results as context
            2. Include conversation context if available (Pro mode)
            3. Generate answer using LLM
            4. Extract source citations
        
        Returns:
            Updated state with final answer
        """
        logger.info("Analyze node: synthesizing answer")
        
        state["reasoning_steps"].append("Analyzing sources and generating answer")
        
        # Handle empty results
        if not state["search_results"]:
            return {
                **state,
                "answer": "Could not find sufficient information to answer your query.",
                "response_time": 0.0
            }
        
        # Build search context from results
        search_context = "\n\n".join([
            f"[{i+1}] {r.get('title', 'No title')}\n{r.get('content', '')[:500]}\nURL: {r.get('url', 'No URL')}"
            for i, r in enumerate(state["search_results"][:5])
        ])
        
        # System prompt based on mode
        if state["mode"] == "simple":
            system_prompt = (
                "You are a helpful assistant. Provide a concise and accurate answer "
                "based on the provided sources. Cite sources using [1], [2] notation."
            )
        else:
            system_prompt = (
                "You are a research assistant. Provide a comprehensive answer with analysis, "
                "comparisons, and reasoning. Cite sources using [1], [2] notation. "
                "Structure your answer with clear sections if needed."
            )
        
        # User prompt with optional context
        if state["mode"] == "pro" and state.get("context"):
            user_prompt = f"""Conversation context:
{state['context']}

Current question: {state['query']}

Information from sources:
{search_context}

Provide a detailed answer with analysis and citations."""
        else:
            user_prompt = f"""Question: {state['query']}

Information from sources:
{search_context}

Provide a clear and accurate answer with citations."""
        
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt}
        ]
        
        # Generate answer
        try:
            answer = await self.llm_client.chat_completion(
                messages, 
                temperature=0.3 if state["mode"] == "simple" else 0.5,
                max_tokens=500 if state["mode"] == "simple" else 1500
            )
        except Exception as e:
            logger.error(f"Answer generation failed: {e}")
            answer = f"Error generating answer: {str(e)}"
        
        state["reasoning_steps"].append("Answer generated successfully")
        
        return {
            **state,
            "answer": answer,
            "context_used": bool(state.get("context"))
        }
    
    async def run(
        self, 
        query: str, 
        previous_messages: list = None,
        mode: str = "auto"
    ) -> dict:
        """
        Execute the research pipeline.
        
        Args:
            query: User query string
            previous_messages: Optional conversation history
            mode: Ignored - router decides automatically
        
        Returns:
            Dictionary containing:
                - mode: Selected mode (simple/pro)
                - query: Original query
                - answer: Generated answer
                - sources: List of source documents
                - reasoning_steps: Pipeline execution steps
                - response_time: Total execution time
                - context_used: Whether context was utilized
                - router_decision: Router's decision details
        """
        start_time = time.time()
        
        # Build conversation context if available
        context = None
        if previous_messages and self.context_manager.should_use_context(query, previous_messages):
            context = self.context_manager.build_context(previous_messages)
        
        # Initialize state
        initial_state = {
            "query": query,
            "mode": "simple",  # Will be overridden by router
            "search_results": [],
            "reasoning_steps": [],
            "answer": "",
            "sources": [],
            "context": context or "",
            "confidence": 0.0,
            "router_decision": {},
            "response_time": 0.0,
            "context_used": False,
            "subqueries": []
        }
        
        # Execute graph
        try:
            final_state = await self.app.ainvoke(initial_state)
        except Exception as e:
            logger.error(f"Graph execution failed: {e}")
            return {
                "mode": "error",
                "query": query,
                "answer": f"Pipeline execution failed: {str(e)}",
                "sources": [],
                "reasoning_steps": ["Error in pipeline execution"],
                "response_time": time.time() - start_time,
                "context_used": False,
                "router_decision": {"mode": "error", "confidence": 0.0, "reason": str(e)}
            }
        
        # Calculate total execution time
        final_state["response_time"] = time.time() - start_time
        
        return {
            "mode": final_state["mode"],
            "query": query,
            "answer": final_state["answer"],
            "sources": final_state["sources"],
            "reasoning_steps": final_state["reasoning_steps"],
            "search_queries": final_state.get("subqueries", [query]),
            "response_time": final_state["response_time"],
            "context_used": final_state["context_used"],
            "router_decision": final_state["router_decision"]
        }
    
    def visualize(self) -> str:
        """
        Generate Mermaid diagram of the graph structure.
        
        Returns:
            Mermaid markup string for visualization
        """
        try:
            return self.app.get_graph().draw_mermaid()
        except Exception as e:
            logger.error(f"Graph visualization failed: {e}")
            return "Graph visualization not available"