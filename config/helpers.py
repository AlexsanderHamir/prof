import json
from pathlib import Path
import sys
from typing import Dict, Any, List, Optional
from dataclasses import dataclass
from exit_codes import CONFIG_VALIDATION_ERROR, MISSING_CONFIG_FILE

REQUIRED_FIELDS = ["api_key", "base_url", "model_config"]
MODEL_CONFIG_FIELDS = ["model", "max_tokens", "temperature", "top_p", "prompt_location"]


@dataclass
class BenchmarkFilter:
    prefixes: List[str]
    ignore: Optional[str] = None


@dataclass
class ModelConfig:
    model: str
    max_tokens: int
    temperature: float
    top_p: float
    prompt_location: str


@dataclass
class Config:
    api_key: str
    base_url: str
    model_config: ModelConfig
    benchmark_filters: Dict[str, BenchmarkFilter]


def validate_config_structure(config_data: Dict[str, Any]) -> None:
    for field in REQUIRED_FIELDS:
        if field not in config_data:
            print(f"Missing required field: {field}", file=sys.stderr)
            sys.exit(CONFIG_VALIDATION_ERROR)


def validate_model_config(model_config: Dict[str, Any]) -> None:
    for field in MODEL_CONFIG_FIELDS:
        if field not in model_config:
            print(f"Missing required field in model_config: {field}", file=sys.stderr)
            sys.exit(CONFIG_VALIDATION_ERROR)


def validate_benchmark_configs(benchmark_filters: Dict[str, Any]) -> None:
    for benchmark, config in benchmark_filters.items():
        if not isinstance(config, dict):
            print(f"Invalid benchmark config format for {benchmark}", file=sys.stderr)
            sys.exit(CONFIG_VALIDATION_ERROR)
        if "prefixes" not in config:
            print(f"Missing 'prefixes' for benchmark {benchmark}", file=sys.stderr)
            sys.exit(CONFIG_VALIDATION_ERROR)
        if not isinstance(config["prefixes"], list):
            print(f"'prefixes' must be a list for benchmark {benchmark}", file=sys.stderr)
            sys.exit(CONFIG_VALIDATION_ERROR)
        if "ignore" in config and not isinstance(config["ignore"], str):
            print(f"'ignore' must be a string for benchmark {benchmark}", file=sys.stderr)
            sys.exit(CONFIG_VALIDATION_ERROR)


def fail(msg: str) -> None:
    print(f"[config error] {msg}", file=sys.stderr)
    sys.exit(CONFIG_VALIDATION_ERROR)


def validate_ai_config(ai_config: Dict[str, Any]) -> None:
    if not ai_config:
        fail("ai_config is required")

    all_benchmarks = ai_config.get("all_benchmarks", True)
    all_profiles = ai_config.get("all_profiles", True)
    specific_benchmarks = ai_config.get("specific_benchmarks", [])
    specific_profiles = ai_config.get("specific_profiles", [])
    per_benchmark_config = ai_config.get("per_benchmark_config", {})
    universal_profile_filter = ai_config.get("universal_profile_filter")

    if all_benchmarks and all_profiles:
        if specific_benchmarks:
            fail("When all_benchmarks and all_profiles are both True, specific_benchmarks must be empty")
        if specific_profiles:
            fail("When all_benchmarks and all_profiles are both True, specific_profiles must be empty")
        if per_benchmark_config:
            fail("When all_benchmarks and all_profiles are both True, per_benchmark_config must be empty")

    if not all_benchmarks and not (specific_benchmarks or per_benchmark_config):
        fail("When all_benchmarks is False, provide specific_benchmarks or per_benchmark_config")

    if not all_profiles and not (specific_profiles or per_benchmark_config):
        fail("When all_profiles is False, provide specific_profiles or per_benchmark_config")

    if bool(specific_benchmarks) != bool(specific_profiles):
        fail("specific_benchmarks and specific_profiles must both be set or both be empty")

    if per_benchmark_config and (specific_benchmarks or specific_profiles):
        fail("When per_benchmark_config exists, specific_benchmarks and specific_profiles must be empty")

    if universal_profile_filter:
        if not isinstance(universal_profile_filter, dict):
            fail("universal_profile_filter must be a dictionary")
        if "profile_values" not in universal_profile_filter:
            fail("universal_profile_filter must contain 'profile_values'")


def create_config_template() -> Dict[str, Any]:
    return {
        "api_key": "your-api-key-here",
        "base_url": "https://api.openai.com/v1",
        "model_config": {
            "model": "gpt-4-turbo-preview",
            "max_tokens": 4096,
            "temperature": 0.7,
            "top_p": 1.0,
            "prompt_location": "path/to/your/system_prompt.txt",
        },
        "benchmark_configs": {
            "BenchmarkGenPool": {
                "prefixes": [
                    "github.com/example/GenPool",
                    "github.com/example/GenPool/internal",
                    "github.com/example/GenPool/pkg",
                ],
                "ignore": "init,TestMain,BenchmarkMain",
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
        "ai_config": {
            "all_benchmarks": True,
            "all_profiles": True,
            "universal_profile_filter": {
                "profile_values": {
                    "flat": 0.1,
                    "flat%": 0.1,
                    "sum%": 0.1,
                    "cum": 0.1,
                    "cum%": 0.1,
                },
                "ignore_functions": ["init", "TestMain", "BenchmarkMain"],
                "ignore_prefixes": ["github.com/example/GenPool", "github.com/example/GenPool/internal", "github.com/example/GenPool/pkg"],
            },
            "specific_benchmarks": ["BenchmarkGenPool", "BenchmarkSyncPool"],
            "specific_profiles": ["cpu", "mem"],
            "per_benchmark_config": {
                "BenchmarkGenPool": {
                    "specific_profiles": ["cpu", "mem", "mutex"],
                },
                "BenchmarkSyncPool": {
                    "specific_profiles": ["cpu", "mem"],
                }
            }
        }
    }


def save_template_to_file(template: Dict[str, Any], output_path: Path) -> None:
    with open(output_path, "w") as f:
        json.dump(template, f, indent=4)


def load_config_from_file(config_path: str) -> Dict[str, Any]:
    try:
        with open(config_path, "r") as f:
            return json.load(f)
    except Exception as e:
        print(f"Error reading configuration file: {str(e)}", file=sys.stderr)
        sys.exit(MISSING_CONFIG_FILE)


def create_config_from_data(config_data: Dict[str, Any]) -> Config:
    model_config = ModelConfig(**config_data["model_config"])

    benchmark_filters = {}
    if "benchmark_configs" in config_data:
        for benchmark, config in config_data["benchmark_configs"].items():
            benchmark_filters[benchmark] = BenchmarkFilter(prefixes=config["prefixes"], ignore=config.get("ignore"))

    return Config(
        api_key=config_data["api_key"],
        base_url=config_data["base_url"],
        model_config=model_config,
        benchmark_filters=benchmark_filters,
    )


def print_template_creation_info(template_path: Path) -> None:

    print(f"\nTemplate configuration file created at: {template_path}")
    print("\nThe template includes example benchmark configurations with multiple prefixes.")
    print("For each benchmark, you can specify:")
    print("  - prefixes: A list of package prefixes to analyze (e.g., ['github.com/your/package'])")
    print("  - ignore: Optional comma-separated list of functions to exclude from analysis")
    print("\nPlease edit this file with your configuration and run:")
    print("  prof setup --config path/to/your/config.json")


def print_validation_progress(message: str, *args, **kwargs) -> None:
    print(f"\n{message.format(*args, **kwargs) if args or kwargs else message}")
