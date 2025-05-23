import os
import json
from pathlib import Path
from typing import Optional, List, Dict
from dataclasses import dataclass
from openai import OpenAI
import sys

@dataclass
class BenchmarkConfig:
    prefixes: List[str]  # List of prefixes to analyze
    ignore: Optional[str] = None  # Optional comma-separated list of functions to ignore

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
    benchmark_configs: Dict[str, BenchmarkConfig]  # Map of benchmark name to its configuration

class ConfigManager:
    CONFIG_DIR = Path.home() / ".prof_benchmark"
    CONFIG_FILE = CONFIG_DIR / "config.json"
    DEFAULT_SYSTEM_PROMPT_TEMPLATE = "Optional: Your system prompt here"
    
    # Handle both development and PyInstaller environments
    if getattr(sys, 'frozen', False):
        # Running in a PyInstaller bundle
        TEMPLATES_DIR = Path(sys._MEIPASS) / "templates"
    else:
        # Running in normal Python environment
        TEMPLATES_DIR = Path(__file__).parent / "templates"
    
    SYSTEM_PROMPT_FILE = TEMPLATES_DIR / "system_prompt.txt"
    
    @classmethod
    def _load_default_system_prompt(cls) -> str:
        try:
            with open(cls.SYSTEM_PROMPT_FILE, 'r') as f:
                return f.read().strip()
        except Exception as e:
            raise ValueError(f"Error loading default system prompt template: {str(e)}")
    
    @classmethod
    def get_default_system_prompt(cls) -> str:
        if not hasattr(cls, '_default_system_prompt'):
            cls._default_system_prompt = cls._load_default_system_prompt()
        return cls._default_system_prompt
    
    @classmethod
    def create_template(cls, output_path: Optional[str] = None) -> None:
        template = {
            # API configuration
            "api_key": "your-api-key-here",
            "base_url": "https://api.openai.com/v1",
            
            # Model configuration for AI analysis
            "model_config": {
                "model": "gpt-4-turbo-preview",
                "max_tokens": 4096,
                "temperature": 0.7,
                "top_p": 1.0,
                "general_analyze_prompt_location": "path/to/your/system_prompt.txt",
                "deep_analyze_prompt_location": "path/to/your/deep_analyze_prompt.txt"
            },
            
            # Benchmark configurations
            # Each benchmark can have multiple prefixes to analyze and an optional list of functions to ignore
            # The prefixes are used to filter which functions to analyze in the profile output
            # The ignore list is a comma-separated string of function names to exclude from analysis
            "benchmark_configs": {
                # Example benchmark with multiple prefixes
                "BenchmarkGenPool": {
                    "prefixes": [
                        "github.com/example/GenPool",  # Main package prefix
                        "github.com/example/GenPool/internal",  # Internal package prefix
                        "github.com/example/GenPool/pkg"  # Public package prefix
                    ],
                    "ignore": "init,TestMain,BenchmarkMain"  # Functions to ignore in analysis
                },
                
                # Example benchmark with a single prefix
                "BenchmarkSyncPool": {
                    "prefixes": [
                        "github.com/example/SyncPool"
                    ],
                    "ignore": "setup,teardown"  # Functions to ignore in analysis
                },
                
                # Example benchmark with no ignore list
                "BenchmarkCustomPool": {
                    "prefixes": [
                        "github.com/example/CustomPool",
                        "github.com/example/CustomPool/optimizations"
                    ]
                }
            }
        }
        
        if output_path:
            template_path = Path(output_path)
        else:
            template_path = Path.cwd() / "config_template.json"
        
        with open(template_path, 'w') as f:
            json.dump(template, f, indent=4)
        
        print(f"\nTemplate configuration file created at: {template_path}")
        print("\nThe template includes example benchmark configurations with multiple prefixes.")
        print("For each benchmark, you can specify:")
        print("  - prefixes: A list of package prefixes to analyze (e.g., ['github.com/your/package'])")
        print("  - ignore: Optional comma-separated list of functions to exclude from analysis")
        print("\nPlease edit this file with your configuration and run:")
        print("  prof setup --config path/to/your/config.json")
    
    @classmethod
    def setup_from_file(cls, config_path: str) -> None:
        print("\nStarting configuration validation process...")
        print(f"Reading configuration from: {config_path}")
        try:
            with open(config_path, 'r') as f:
                config_data = json.load(f)
            print("✓ Successfully read configuration file")
        except Exception as e:
            raise ValueError(f"Error reading configuration file: {str(e)}")
        
        print("\nValidating configuration structure...")
        required_fields = ["api_key", "base_url", "model_config"]
        for field in required_fields:
            if field not in config_data:
                raise ValueError(f"Missing required field: {field}")
        print("✓ All required top-level fields are present")
        
        # Make benchmark_configs optional with a default empty dict
        if "benchmark_configs" not in config_data:
            config_data["benchmark_configs"] = {}
            print("✓ No benchmark configurations provided - will analyze all functions")
        
        print("\nValidating model configuration...")
        model_config_fields = ["model", "max_tokens", "temperature", "top_p"]
        for field in model_config_fields:
            if field not in config_data["model_config"]:
                raise ValueError(f"Missing required field in model_config: {field}")
        print("✓ All required model configuration fields are present")
        
        print("\nValidating benchmark configurations...")
        for benchmark, config in config_data["benchmark_configs"].items():
            if not isinstance(config, dict):
                raise ValueError(f"Invalid benchmark config format for {benchmark}")
            if "prefixes" not in config:
                raise ValueError(f"Missing 'prefixes' for benchmark {benchmark}")
            if not isinstance(config["prefixes"], list):
                raise ValueError(f"'prefixes' must be a list for benchmark {benchmark}")
            if "ignore" in config and not isinstance(config["ignore"], str):
                raise ValueError(f"'ignore' must be a string for benchmark {benchmark}")
        print("✓ All benchmark configurations are valid")
        
        # Store the config path as a class variable for future use
        cls._config_path = config_path
        print("\nConfiguration validation completed successfully! 🎉")
        print(f"Using configuration from: {config_path}")
    
    @classmethod
    def load(cls) -> Config:
        """Load the configuration from file."""
        if not hasattr(cls, '_config_path'):
            raise ValueError(
                "Configuration not found. Please run setup first using:\n"
                "prof setup --config path/to/your/config.json\n"
                "or create a template using:\n"
                "prof setup --create-template \n"
                "and then run the command: \n"
                "prof setup --config path/to/your/config.json"
            )
        
        try:
            with open(cls._config_path, 'r') as f:
                config_data = json.load(f)
            
            model_config = ModelConfig(**config_data["model_config"])
            
            # Convert benchmark configs to BenchmarkConfig objects, defaulting to empty dict if not present
            benchmark_configs = {}
            if "benchmark_configs" in config_data:
                for benchmark, config in config_data["benchmark_configs"].items():
                    benchmark_configs[benchmark] = BenchmarkConfig(
                        prefixes=config["prefixes"],
                        ignore=config.get("ignore")
                    )
            
            return Config(
                api_key=config_data["api_key"],
                base_url=config_data["base_url"],
                model_config=model_config,
                benchmark_configs=benchmark_configs
            )
        except Exception as e:
            raise ValueError(f"Error loading configuration: {str(e)}")
    
    @classmethod
    def get_client(cls) -> OpenAI:
        """Get an OpenAI client configured with the current settings."""
        config = cls.load()
        return OpenAI(
            api_key=config.api_key,
            base_url=config.base_url
        )
    
    @classmethod
    def is_configured(cls) -> bool:
        """Check if the configuration exists and is valid."""
        try:
            cls.load()
            return True
        except ValueError:
            return False
    
    @classmethod
    def clear_cache(cls) -> None:
        """This method is deprecated as caching has been removed."""
        pass 

    @classmethod
    def _load_system_prompt_from_location(cls, location: str) -> str:
        """Load system prompt from the specified file location."""
        try:
            with open(location, 'r') as f:
                return f.read().strip()
        except Exception as e:
            raise ValueError(f"Error loading system prompt from {location}: {str(e)}")
    
    @classmethod
    def get_deep_analyze_prompt(cls, config: Config) -> str:
        """Get the deep analyze prompt from the configured location or use default."""
        if config.model_config.deep_analyze_prompt_location:
            try:
                prompt = cls._load_system_prompt_from_location(config.model_config.deep_analyze_prompt_location)
                print(f"\nUsing custom deep analyze prompt from: {config.model_config.deep_analyze_prompt_location}")
                return prompt
            except Exception as e:
                print(f"\nWarning: Could not load custom deep analyze prompt: {str(e)}")
                print("Falling back to default system prompt.")
                return cls.get_default_system_prompt()
        return cls.get_default_system_prompt()

    @classmethod
    def get_system_prompt(cls, config: Config, prompt_type: str = "general") -> str:
        """Get the system prompt from the configured location.
        
        Args:
            config: The current configuration
            prompt_type: Either "general" or "deep" to specify which prompt to load
            
        Raises:
            ValueError: If prompt location is not provided or file cannot be read
        """
        if prompt_type == "deep":
            if not config.model_config.deep_analyze_prompt_location:
                raise ValueError("deep_analyze_prompt_location must be provided in config")
            try:
                prompt = cls._load_system_prompt_from_location(config.model_config.deep_analyze_prompt_location)
                print(f"\nUsing deep analyze prompt from: {config.model_config.deep_analyze_prompt_location}")
                return prompt
            except Exception as e:
                raise ValueError(f"Could not load deep analyze prompt: {str(e)}")
        elif prompt_type == "general":
            if not config.model_config.general_analyze_prompt_location:
                raise ValueError("general_analyze_prompt_location must be provided in config")
            try:
                prompt = cls._load_system_prompt_from_location(config.model_config.general_analyze_prompt_location)
                print(f"\nUsing general analyze prompt from: {config.model_config.general_analyze_prompt_location}")
                return prompt
            except Exception as e:
                raise ValueError(f"Could not load general analyze prompt: {str(e)}")
        else:
            raise ValueError(f"Invalid prompt type: {prompt_type}. Must be either 'general' or 'deep'") 