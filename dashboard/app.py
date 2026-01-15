"""Streamlit Dashboard for AI Agent Evaluation Pipeline"""
import streamlit as st
import requests
import pandas as pd
import plotly.express as px
import plotly.graph_objects as go
from datetime import datetime
import json
import os

# API Configuration
GO_API_URL = os.getenv("GO_API_URL", "http://localhost:8080")
PYTHON_API_URL = os.getenv("PYTHON_API_URL", "http://localhost:8081")

st.set_page_config(
    page_title="AI Agent Evaluation Pipeline",
    page_icon="🤖",
    layout="wide",
    initial_sidebar_state="expanded"
)

# Custom CSS with modern aesthetics
st.markdown("""
<style>
    @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Space+Grotesk:wght@400;500;600;700&display=swap');
    
    .main-header {
        font-family: 'Space Grotesk', sans-serif;
        font-size: 2.5rem;
        font-weight: 700;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        margin-bottom: 1.5rem;
    }
    
    .metric-card {
        background: linear-gradient(135deg, #1e1e2e 0%, #2d2d3f 100%);
        padding: 1.5rem;
        border-radius: 1rem;
        border-left: 4px solid #667eea;
        color: white;
    }
    
    .success-box {
        padding: 1rem;
        background: linear-gradient(135deg, #065f46 0%, #047857 100%);
        border-radius: 0.75rem;
        color: white;
    }
    
    .warning-box {
        padding: 1rem;
        background: linear-gradient(135deg, #92400e 0%, #b45309 100%);
        border-radius: 0.75rem;
        color: white;
    }
    
    .error-box {
        padding: 1rem;
        background: linear-gradient(135deg, #991b1b 0%, #dc2626 100%);
        border-radius: 0.75rem;
        color: white;
    }
    
    .stButton>button {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        color: white;
        border: none;
        border-radius: 0.5rem;
        padding: 0.5rem 1rem;
        font-weight: 600;
    }
</style>
""", unsafe_allow_html=True)


def make_request(base_url: str, endpoint: str, method: str = "GET", data: dict = None, params: dict = None):
    """Make API request with error handling"""
    try:
        url = f"{base_url}{endpoint}"
        if method == "GET":
            response = requests.get(url, params=params, timeout=10)
        elif method == "POST":
            response = requests.post(url, json=data, params=params, timeout=30)
        
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        return None


def render_sidebar():
    """Render sidebar navigation"""
    st.sidebar.markdown("# 🤖 AI Agent Eval")
    st.sidebar.markdown("### Go + Python Pipeline")
    st.sidebar.markdown("---")
    
    page = st.sidebar.radio(
        "Navigation",
        [
            "📊 Overview",
            "📥 Ingest Data",
            "📈 Evaluations",
            "✍️ Annotations",
            "💡 Improvements",
            "🎯 Meta-Evaluation",
            "🔍 Explorer"
        ]
    )
    
    st.sidebar.markdown("---")
    st.sidebar.markdown("### System Health")
    
    # Check Go API
    go_health = make_request(GO_API_URL, "/health")
    if go_health:
        st.sidebar.success("✓ Go API Connected")
    else:
        st.sidebar.error("✗ Go API Disconnected")
    
    # Check Python API
    python_health = make_request(PYTHON_API_URL, "/health")
    if python_health:
        st.sidebar.success("✓ Python Evaluator Connected")
    else:
        st.sidebar.warning("⚠ Python Evaluator Offline")
    
    return page


def render_overview():
    """Render overview dashboard"""
    st.markdown('<div class="main-header">📊 System Overview</div>', unsafe_allow_html=True)
    
    stats = make_request(GO_API_URL, "/api/v1/stats")
    
    if stats:
        col1, col2, col3, col4 = st.columns(4)
        
        with col1:
            st.metric("Total Conversations", f"{stats.get('total_conversations', 0):,}")
        with col2:
            st.metric("Total Evaluations", f"{stats.get('total_evaluations', 0):,}")
        with col3:
            st.metric("Total Annotations", f"{stats.get('total_annotations', 0):,}")
        with col4:
            avg_score = stats.get('average_quality_score')
            st.metric("Avg Quality Score", f"{avg_score:.2f}" if avg_score else "N/A")
        
        st.markdown("---")
        
        col1, col2, col3 = st.columns(3)
        
        with col1:
            st.info(f"**Open Issues**: {stats.get('open_issues_count', 0)}")
        with col2:
            st.info(f"**Pending Suggestions**: {stats.get('pending_suggestions_count', 0)}")
        with col3:
            st.info(f"**Evaluations (24h)**: {stats.get('evaluations_last_24h', 0)}")
        
        st.markdown("---")
        
        # Recent evaluations chart
        st.subheader("Recent Evaluation Scores")
        evaluations = make_request(GO_API_URL, "/api/v1/evaluations", params={"limit": 50})
        
        if evaluations and evaluations.get("evaluations"):
            df = pd.DataFrame(evaluations["evaluations"])
            if 'created_at' in df.columns:
                df['created_at'] = pd.to_datetime(df['created_at'])
                df = df.sort_values('created_at')
                
                fig = px.line(
                    df,
                    x='created_at',
                    y='overall_score',
                    title="Quality Score Trend",
                    labels={'overall_score': 'Overall Score', 'created_at': 'Time'}
                )
                fig.add_hline(y=0.7, line_dash="dash", line_color="red", annotation_text="Threshold")
                fig.update_layout(
                    plot_bgcolor='rgba(0,0,0,0)',
                    paper_bgcolor='rgba(0,0,0,0)'
                )
                st.plotly_chart(fig, use_container_width=True)
    else:
        st.warning("Unable to fetch system stats. Is the Go API running?")


def render_ingest_data():
    """Render data ingestion page"""
    st.markdown('<div class="main-header">📥 Ingest Data</div>', unsafe_allow_html=True)
    
    tab1, tab2 = st.tabs(["Single Conversation", "Batch Upload"])
    
    with tab1:
        st.subheader("Ingest Single Conversation")
        
        with st.form("ingest_form"):
            conv_id = st.text_input("Conversation ID", value=f"conv_{datetime.now().strftime('%Y%m%d_%H%M%S')}")
            agent_version = st.text_input("Agent Version", value="v1.0.0")
            
            st.markdown("**Conversation JSON**")
            default_json = '''{
    "turns": [
        {
            "turn_id": 1,
            "role": "user",
            "content": "I need to book a flight to NYC next week",
            "timestamp": "2024-01-15T10:30:00Z"
        },
        {
            "turn_id": 2,
            "role": "assistant",
            "content": "I'd be happy to help you book a flight to NYC!",
            "tool_calls": [
                {
                    "tool_name": "flight_search",
                    "parameters": {"destination": "NYC"},
                    "result": {"status": "success"},
                    "latency_ms": 450
                }
            ],
            "timestamp": "2024-01-15T10:30:02Z"
        }
    ],
    "metadata": {
        "total_latency_ms": 1200,
        "mission_completed": true
    }
}'''
            conversation_json = st.text_area("Paste conversation JSON here", height=300, value=default_json)
            
            auto_evaluate = st.checkbox("Auto-evaluate", value=True)
            submitted = st.form_submit_button("Ingest Conversation")
            
            if submitted:
                try:
                    conv_data = json.loads(conversation_json)
                    conv_data["conversation_id"] = conv_id
                    conv_data["agent_version"] = agent_version
                    
                    result = make_request(
                        GO_API_URL,
                        "/api/v1/conversations",
                        method="POST",
                        data=conv_data,
                        params={"auto_evaluate": str(auto_evaluate).lower()}
                    )
                    
                    if result:
                        st.success(f"✓ Conversation ingested: {result.get('conversation_id', conv_id)}")
                        if auto_evaluate:
                            st.info("Evaluation triggered automatically")
                    else:
                        st.error("Failed to ingest conversation")
                except json.JSONDecodeError:
                    st.error("Invalid JSON format")
    
    with tab2:
        st.subheader("Batch Upload")
        st.info("Upload a JSON file containing an array of conversations")
        
        uploaded_file = st.file_uploader("Choose JSON file", type=['json'])
        
        if uploaded_file:
            try:
                conversations = json.load(uploaded_file)
                st.success(f"Loaded {len(conversations)} conversations")
                
                if st.button("Ingest Batch"):
                    result = make_request(
                        GO_API_URL,
                        "/api/v1/conversations/batch",
                        method="POST",
                        data=conversations,
                        params={"auto_evaluate": "true"}
                    )
                    
                    if result:
                        st.success(f"✓ Ingested {result.get('ingested', 0)} conversations")
            except json.JSONDecodeError:
                st.error("Invalid JSON file")


def render_evaluations():
    """Render evaluations page"""
    st.markdown('<div class="main-header">📈 Evaluations</div>', unsafe_allow_html=True)
    
    col1, col2 = st.columns([2, 1])
    
    with col1:
        conv_id_filter = st.text_input("Filter by Conversation ID", "")
    with col2:
        limit = st.number_input("Limit", min_value=10, max_value=500, value=50)
    
    params = {"limit": limit}
    if conv_id_filter:
        params["conversation_id"] = conv_id_filter
    
    evaluations = make_request(GO_API_URL, "/api/v1/evaluations", params=params)
    
    if evaluations and evaluations.get("evaluations"):
        st.success(f"Found {evaluations.get('count', 0)} evaluations")
        
        df = pd.DataFrame(evaluations["evaluations"])
        st.dataframe(df, use_container_width=True, hide_index=True)
        
        st.markdown("---")
        st.subheader("Evaluation Details")
        
        if len(df) > 0:
            eval_ids = df['evaluation_id'].tolist()
            selected_eval_id = st.selectbox("Select Evaluation", eval_ids)
            
            if selected_eval_id:
                eval_detail = make_request(GO_API_URL, f"/api/v1/evaluations/{selected_eval_id}")
                
                if eval_detail:
                    col1, col2 = st.columns(2)
                    
                    with col1:
                        st.markdown("### Scores")
                        scores = eval_detail.get('scores', {})
                        
                        fig = go.Figure(data=[
                            go.Bar(
                                x=list(scores.keys()),
                                y=list(scores.values()),
                                marker_color=['#667eea', '#764ba2', '#f093fb', '#f5576c']
                            )
                        ])
                        fig.update_layout(
                            title="Score Breakdown",
                            yaxis_range=[0, 1],
                            yaxis_title="Score",
                            plot_bgcolor='rgba(0,0,0,0)'
                        )
                        st.plotly_chart(fig, use_container_width=True)
                    
                    with col2:
                        st.markdown("### Issues Detected")
                        issues = eval_detail.get('issues_detected', [])
                        
                        if issues:
                            for issue in issues:
                                severity = issue.get('severity', 'info')
                                with st.container():
                                    if severity == 'critical':
                                        st.error(f"**{issue.get('type')}**: {issue.get('description', '')}")
                                    elif severity == 'warning':
                                        st.warning(f"**{issue.get('type')}**: {issue.get('description', '')}")
                                    else:
                                        st.info(f"**{issue.get('type')}**: {issue.get('description', '')}")
                        else:
                            st.success("No issues detected")
    else:
        st.warning("No evaluations found")


def render_annotations():
    """Render annotations page"""
    st.markdown('<div class="main-header">✍️ Annotations</div>', unsafe_allow_html=True)
    
    st.subheader("Add Annotation")
    
    with st.form("annotation_form"):
        col1, col2 = st.columns(2)
        
        with col1:
            conv_id = st.text_input("Conversation ID")
            annotator_id = st.text_input("Annotator ID", value="annotator_001")
            annotation_type = st.selectbox(
                "Annotation Type",
                ["tool_accuracy", "response_quality", "coherence", "general_quality"]
            )
        
        with col2:
            label = st.text_input("Label", value="correct")
            score = st.slider("Score", 0.0, 1.0, 0.8, 0.05)
            confidence = st.slider("Confidence", 0.0, 1.0, 0.9, 0.05)
        
        notes = st.text_area("Notes")
        
        submitted = st.form_submit_button("Submit Annotation")
        
        if submitted and conv_id and annotator_id:
            annotation_data = {
                "conversation_id": conv_id,
                "annotator_id": annotator_id,
                "annotation_type": annotation_type,
                "label": label,
                "score": score,
                "confidence": confidence,
                "notes": notes
            }
            
            result = make_request(GO_API_URL, "/api/v1/annotations", method="POST", data=annotation_data)
            
            if result:
                st.success("✓ Annotation added successfully")
            else:
                st.error("Failed to add annotation")


def render_improvements():
    """Render improvements page"""
    st.markdown('<div class="main-header">💡 Improvements</div>', unsafe_allow_html=True)
    
    tab1, tab2, tab3 = st.tabs(["Suggestions", "Generate New", "Failure Patterns"])
    
    with tab1:
        st.subheader("Pending Suggestions")
        
        col1, col2 = st.columns(2)
        with col1:
            min_confidence = st.slider("Min Confidence", 0.0, 1.0, 0.7, 0.05)
        with col2:
            sugg_type_filter = st.selectbox("Type", ["All", "prompt", "tool_schema", "validation"])
        
        params = {"min_confidence": min_confidence}
        if sugg_type_filter != "All":
            params["suggestion_type"] = sugg_type_filter
        
        suggestions = make_request(GO_API_URL, "/api/v1/improvements/suggestions", params=params)
        
        if suggestions and suggestions.get("suggestions"):
            st.success(f"Found {suggestions.get('count', 0)} suggestions")
            
            for sugg in suggestions["suggestions"]:
                with st.expander(f"🎯 {sugg.get('suggestion_type', 'N/A')} - Confidence: {sugg.get('confidence', 0):.2f}"):
                    st.markdown(f"**Suggestion ID:** `{sugg.get('suggestion_id', 'N/A')}`")
                    st.markdown(f"**Suggestion:** {sugg.get('suggestion', 'N/A')}")
                    st.markdown(f"**Rationale:** {sugg.get('rationale', 'N/A')}")
                    st.markdown(f"**Status:** {sugg.get('status', 'N/A')}")
                    
                    if st.button(f"Mark Implemented", key=sugg.get('suggestion_id')):
                        result = make_request(
                            GO_API_URL,
                            f"/api/v1/improvements/suggestions/{sugg.get('suggestion_id')}/implement",
                            method="POST"
                        )
                        if result:
                            st.success("✓ Marked as implemented")
                            st.rerun()
        else:
            st.info("No pending suggestions")
    
    with tab2:
        st.subheader("Generate New Suggestions")
        
        lookback_days = st.slider("Analysis Period (days)", 1, 90, 7)
        
        if st.button("Analyze & Generate"):
            with st.spinner("Analyzing patterns..."):
                result = make_request(
                    GO_API_URL,
                    "/api/v1/improvements/analyze",
                    method="POST",
                    params={"lookback_days": lookback_days}
                )
                
                if result:
                    st.success("✓ Analysis complete")
                    
                    col1, col2, col3 = st.columns(3)
                    with col1:
                        st.metric("Patterns Detected", result.get('patterns_detected', 0))
                    with col2:
                        st.metric("Suggestions Generated", result.get('suggestions_generated', 0))
                    with col3:
                        st.metric("Prompt Suggestions", result.get('prompt_suggestions', 0))
                    
                    st.json(result)
    
    with tab3:
        st.subheader("Failure Patterns")
        
        patterns = make_request(GO_API_URL, "/api/v1/improvements/patterns", params={"limit": 50})
        
        if patterns and patterns.get("patterns"):
            for p in patterns["patterns"]:
                with st.expander(f"🔴 {p.get('pattern_type', 'N/A')} ({p.get('severity', 'N/A')})"):
                    st.markdown(f"**Occurrences:** {p.get('occurrence_count', 0)}")
                    st.markdown(f"**Description:** {p.get('description', 'N/A')}")
                    st.markdown(f"**Resolved:** {p.get('resolved', False)}")
        else:
            st.info("No failure patterns detected")


def render_meta_evaluation():
    """Render meta-evaluation page"""
    st.markdown('<div class="main-header">🎯 Meta-Evaluation</div>', unsafe_allow_html=True)
    
    tab1, tab2 = st.tabs(["Calibrate", "Performance"])
    
    with tab1:
        st.subheader("Calibrate Evaluators")
        st.info("Compare evaluator predictions with human annotations to improve accuracy")
        
        lookback_days = st.slider("Calibration Period (days)", 7, 180, 30)
        
        if st.button("Run Calibration"):
            with st.spinner("Calibrating evaluators..."):
                result = make_request(
                    GO_API_URL,
                    "/api/v1/meta-evaluation/calibrate",
                    method="POST",
                    params={"lookback_days": lookback_days}
                )
                
                if result:
                    if result.get('status') == 'success' or result.get('calibrations'):
                        st.success(f"✓ Calibration complete")
                        
                        for cal in result.get('calibrations', []):
                            with st.expander(f"📊 {cal.get('evaluator_type', 'N/A')}"):
                                if cal.get('metrics'):
                                    metrics = cal['metrics']
                                    col1, col2 = st.columns(2)
                                    with col1:
                                        st.metric("Correlation", f"{metrics.get('correlation', 0):.3f}")
                                        st.metric("Precision", f"{metrics.get('precision', 0):.3f}")
                                    with col2:
                                        st.metric("Recall", f"{metrics.get('recall', 0):.3f}")
                                        st.metric("F1 Score", f"{metrics.get('f1_score', 0):.3f}")
                                    
                                    if cal.get('blind_spots'):
                                        st.markdown("**Blind Spots:**")
                                        for bs in cal['blind_spots']:
                                            st.warning(f"- {bs.get('description', 'N/A')}")
                    else:
                        st.warning(result.get('message', 'Calibration returned unexpected result'))
                        st.json(result)
    
    with tab2:
        st.subheader("Evaluator Performance")
        
        performance = make_request(GO_API_URL, "/api/v1/meta-evaluation/performance")
        
        if performance and performance.get("evaluators"):
            for perf in performance["evaluators"]:
                with st.expander(f"📈 {perf.get('evaluator_type', 'N/A')}"):
                    col1, col2 = st.columns(2)
                    with col1:
                        st.markdown(f"**Version:** {perf.get('evaluator_version', 'N/A')}")
                        st.markdown(f"**Samples:** {perf.get('calibration_samples', 0)}")
                    with col2:
                        if perf.get('correlation_with_human'):
                            st.metric("Human Correlation", f"{perf['correlation_with_human']:.3f}")
                        if perf.get('f1_score'):
                            st.metric("F1 Score", f"{perf['f1_score']:.3f}")
        else:
            st.info("No evaluator performance data available")


def render_explorer():
    """Render conversation explorer"""
    st.markdown('<div class="main-header">🔍 Explorer</div>', unsafe_allow_html=True)
    
    conversations = make_request(GO_API_URL, "/api/v1/conversations", params={"limit": 100})
    
    if conversations and conversations.get("conversations"):
        df = pd.DataFrame(conversations["conversations"])
        st.dataframe(df, use_container_width=True, hide_index=True)
    else:
        st.info("No conversations found")


def main():
    """Main dashboard application"""
    page = render_sidebar()
    
    if "📊 Overview" in page:
        render_overview()
    elif "📥 Ingest Data" in page:
        render_ingest_data()
    elif "📈 Evaluations" in page:
        render_evaluations()
    elif "✍️ Annotations" in page:
        render_annotations()
    elif "💡 Improvements" in page:
        render_improvements()
    elif "🎯 Meta-Evaluation" in page:
        render_meta_evaluation()
    elif "🔍 Explorer" in page:
        render_explorer()


if __name__ == "__main__":
    main()
