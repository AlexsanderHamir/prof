from typing import Dict, List
from dataclasses import dataclass
from langchain_deepseek import ChatDeepSeek
from config_manager import ConfigManager, ConfigurationError
import json


@dataclass
class ProfileMetrics:
    """Structured representation of profile metrics."""
    cpu_time: float
    memory_usage: float
    goroutines: int
    raw_metrics: Dict[str, float]
    function_calls: Dict[str, int]


@dataclass
class ParsedBenchmarkData:
    """Structured representation of parsed benchmark data."""
    metrics: List[Dict[str, float]]
    key_findings: List[str]
    performance_insights: List[str]
    raw_parsed_data: Dict


class DataPreprocessingAgent:

    def __init__(self, config_path: str = "config_template.json"):
        try:
            ConfigManager.setup_from_file(config_path)
            self.config = ConfigManager.load()
            self.llm = ChatDeepSeek(
                api_key=self.config.api_key,
                model=self.config.model_config.model,
                temperature=self.config.model_config.temperature,
                max_tokens=None)
        except ConfigurationError as e:
            raise RuntimeError(f"Failed to initialize agent: {e}")

    def _parse_text_content(self, text_content: str) -> ParsedBenchmarkData:
        """Parse benchmark text content using AI to extract structured data."""
        prompt = """Analyze the following benchmark output and extract key information in a structured format.
        Focus on:
        1. Performance metrics (CPU time, memory usage, etc.)
        2. Key findings and observations
        3. Performance insights and potential bottlenecks
        
        Format the response as a JSON object with the following structure:
        {
            "metrics": [{"metric_name": value, ...}],
            "key_findings": ["finding1", "finding2", ...],
            "performance_insights": ["insight1", "insight2", ...]
        }
        
        Benchmark output:
        {text_content}
        """

        try:
            response = self.llm.predict(
                prompt.format(text_content=text_content))
            parsed_data = json.loads(response)
            return ParsedBenchmarkData(
                metrics=parsed_data.get("metrics", []),
                key_findings=parsed_data.get("key_findings", []),
                performance_insights=parsed_data.get("performance_insights",
                                                     []),
                raw_parsed_data=parsed_data)
        except Exception as e:
            raise RuntimeError(f"Failed to parse benchmark text: {e}")

    def process_benchmark(self, files: Dict[str,
                                            str]) -> Dict[str, ProfileMetrics]:
        """Process benchmark files and return structured metrics."""
        print(f"Functions content: {files['functions_content']}")
        print(f"Text content: {files['text_content']}")

        # Parse the text content using AI
        parsed_data = self._parse_text_content(files['text_content'])

        # Convert parsed metrics into ProfileMetrics objects
        results = {}
        for metric_set in parsed_data.metrics:
            # Create a ProfileMetrics object for each set of metrics
            profile_metrics = ProfileMetrics(
                cpu_time=metric_set.get('cpu_time', 0.0),
                memory_usage=metric_set.get('memory_usage', 0.0),
                goroutines=metric_set.get('goroutines', 0),
                raw_metrics=metric_set,
                function_calls=metric_set.get('function_calls', {}))
            # Use a meaningful key for the results dictionary
            key = f"benchmark_{len(results)}"
            results[key] = profile_metrics

        print(f"Results: {results}")
        return results
