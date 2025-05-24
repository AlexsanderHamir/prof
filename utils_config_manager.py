import json
from pathlib import Path
from typing import Dict, Any, List, Optional
from dataclasses import dataclass


@dataclass
class BenchmarkConfig:
    prefixes: List[str]
    ignore: Optional[str] = None


@dataclass
class ModelConfig:
    model: str
    max_tokens: int
    temperature: float
    top_p: float
    general_analyze_prompt_location: str
    deep_analyze_prompt_location: str


@dataclass
class Config:
    api_key: str
    base_url: str
    model_config: ModelConfig
    benchmark_configs: Dict[str, BenchmarkConfig]


def validate_config_structure(config_data: Dict[str, Any]) -> None:
    """Validate the basic structure of the configuration file."""
    required_fields = ["api_key", "base_url", "model_config"]
    for field in required_fields:
        if field not in config_data:
            raise ValueError(f"Missing required field: {field}")


def validate_model_config(model_config: Dict[str, Any]) -> None:
    """Validate the model configuration section."""
    model_config_fields = ["model", "max_tokens", "temperature", "top_p"]
    for field in model_config_fields:
        if field not in model_config:
            raise ValueError(
                f"Missing required field in model_config: {field}")


def validate_benchmark_configs(benchmark_configs: Dict[str, Any]) -> None:
    """Validate the benchmark configurations section."""
    for benchmark, config in benchmark_configs.items():
        if not isinstance(config, dict):
            raise ValueError(
                f"Invalid benchmark config format for {benchmark}")
        if "prefixes" not in config:
            raise ValueError(f"Missing 'prefixes' for benchmark {benchmark}")
        if not isinstance(config["prefixes"], list):
            raise ValueError(
                f"'prefixes' must be a list for benchmark {benchmark}")
        if "ignore" in config and not isinstance(config["ignore"], str):
            raise ValueError(
                f"'ignore' must be a string for benchmark {benchmark}")


def create_config_template() -> Dict[str, Any]:
    """Create a template configuration dictionary."""
    return {
        "api_key": "your-api-key-here",
        "base_url": "https://api.openai.com/v1",
        "model_config": {
            "model": "gpt-4-turbo-preview",
            "max_tokens": 4096,
            "temperature": 0.7,
            "top_p": 1.0,
            "general_analyze_prompt_location":
            "path/to/your/system_prompt.txt",
            "deep_analyze_prompt_location":
            "path/to/your/deep_analyze_prompt.txt",
        },
        "benchmark_configs": {
            "BenchmarkGenPool": {
                "prefixes": [
                    "github.com/example/GenPool",
                    "github.com/example/GenPool/internal",
                    "github.com/example/GenPool/pkg",
                ],
                "ignore":
                "init,TestMain,BenchmarkMain",
            },
            "BenchmarkSyncPool": {
                "prefixes": ["github.com/example/SyncPool"],
                "ignore": "setup,teardown",
            },
            "BenchmarkCustomPool": {
                "prefixes": [
                    "github.com/example/CustomPool",
                    "github.com/example/CustomPool/optimizations",
                ]
            },
        },
    }


def save_template_to_file(template: Dict[str, Any], output_path: Path) -> None:
    """Save the template configuration to a file."""
    with open(output_path, "w") as f:
        json.dump(template, f, indent=4)


def load_config_from_file(config_path: str) -> Dict[str, Any]:
    """Load and parse the configuration file."""
    try:
        with open(config_path, "r") as f:
            return json.load(f)
    except Exception as e:
        raise ValueError(f"Error reading configuration file: {str(e)}")


def create_config_from_data(config_data: Dict[str, Any]) -> Config:
    """Create a Config object from the configuration data."""
    model_config = ModelConfig(**config_data["model_config"])

    benchmark_configs = {}
    if "benchmark_configs" in config_data:
        for benchmark, config in config_data["benchmark_configs"].items():
            benchmark_configs[benchmark] = BenchmarkConfig(
                prefixes=config["prefixes"], ignore=config.get("ignore"))

    return Config(
        api_key=config_data["api_key"],
        base_url=config_data["base_url"],
        model_config=model_config,
        benchmark_configs=benchmark_configs,
    )


def print_template_creation_info(template_path: Path) -> None:
    """Print information about the created template file."""
    print(f"\nTemplate configuration file created at: {template_path}")
    print(
        "\nThe template includes example benchmark configurations with multiple prefixes."
    )
    print("For each benchmark, you can specify:")
    print(
        "  - prefixes: A list of package prefixes to analyze (e.g., ['github.com/your/package'])"
    )
    print(
        "  - ignore: Optional comma-separated list of functions to exclude from analysis"
    )
    print("\nPlease edit this file with your configuration and run:")
    print("  prof setup --config path/to/your/config.json")


def print_validation_progress(message: str, *args, **kwargs) -> None:
    print(
        f"\n{message.format(*args, **kwargs) if args or kwargs else message}")
