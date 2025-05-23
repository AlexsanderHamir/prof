import os
import json
from pathlib import Path
from typing import Optional
from dataclasses import dataclass
from openai import OpenAI
import sys

@dataclass
class ModelConfig:
    model: str
    max_tokens: int
    temperature: float
    top_p: float
    general_analyze_prompt_location: Optional[str] = None
    deep_analyze_prompt_location: Optional[str] = None

@dataclass
class Config:
    api_key: str
    base_url: str
    model_config: ModelConfig

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
            "api_key": "your-api-key-here",
            "base_url": "https://api.openai.com/v1",
            "model_config": {
                "model": "gpt-4-turbo-preview",
                "max_tokens": 4096,
                "temperature": 0.7,
                "top_p": 1.0,
                "general_analyze_prompt_location": "path/to/your/system_prompt.txt",
                "deep_analyze_prompt_location": "path/to/your/deep_analyze_prompt.txt"
            }
        }
        
        if output_path:
            template_path = Path(output_path)
        else:
            template_path = Path.cwd() / "config_template.json"
        
        with open(template_path, 'w') as f:
            json.dump(template, f, indent=4)
        
        print(f"\nTemplate configuration file created at: {template_path}")
        print("Please edit this file with your configuration and run 'prof setup --config path/to/your/config.json'")
    
    @classmethod
    def setup_from_file(cls, config_path: str) -> None:
        print("\nStarting configuration setup process...")
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
        
        print("\nValidating model configuration...")
        model_config_fields = ["model", "max_tokens", "temperature", "top_p"]
        for field in model_config_fields:
            if field not in config_data["model_config"]:
                raise ValueError(f"Missing required field in model_config: {field}")
        print("✓ All required model configuration fields are present")
        
        print("\nCreating configuration directory...")
        cls.CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        print(f"✓ Configuration directory ready at: {cls.CONFIG_DIR}")
            
        try:
            print("\nSaving configuration with secure permissions...")
            # Save configuration making it restrictive
            with open(cls.CONFIG_FILE, 'w') as f:
                json.dump(config_data, f, indent=2)
            
            # Set restrictive permissions on the config file
            os.chmod(cls.CONFIG_FILE, 0o600)
            print(f"✓ Configuration successfully saved to: {cls.CONFIG_FILE}")
            print("✓ File permissions set to secure mode (600)")
            print("\nConfiguration setup completed successfully! 🎉")
        except Exception as e:
            raise ValueError(f"Error saving configuration: {str(e)}")
    
    @classmethod
    def load(cls) -> Config:
        """Load the configuration from file."""
        if not cls.CONFIG_FILE.exists():
            raise ValueError(
                "Configuration not found. Please run setup first using:\n"
                "prof setup --config path/to/your/config.json\n"
                "or create a template using:\n"
                "prof setup --create-template \n"
                "and then run the command: \n"
                "prof setup --config path/to/your/config.json"
            )
        
        try:
            with open(cls.CONFIG_FILE, 'r') as f:
                config_data = json.load(f)
            
            model_config = ModelConfig(**config_data["model_config"])
            return Config(
                api_key=config_data["api_key"],
                base_url=config_data["base_url"],
                model_config=model_config
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
    def clean_config(cls) -> None:
        """Remove the configuration file if it exists.
        
        This method will delete the configuration file from the system.
        Use with caution as this will require reconfiguration to use the application again.
        """
        if not cls.CONFIG_FILE.exists():
            print("No configuration file found to clean.")
            return
            
        try:
            cls.CONFIG_FILE.unlink()
            print(f"✓ Configuration file successfully removed from: {cls.CONFIG_FILE}")
            print("To reconfigure, run: prof setup --config path/to/your/config.json")
        except Exception as e:
            raise ValueError(f"Error removing configuration file: {str(e)}")
    
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
        """Get the system prompt from the configured location or use default.
        
        Args:
            config: The current configuration
            prompt_type: Either "general" or "deep" to specify which prompt to load
        """
        if prompt_type == "deep":
            return cls.get_deep_analyze_prompt(config)
        elif prompt_type == "general":
            if config.model_config.general_analyze_prompt_location:
                try:
                    prompt = cls._load_system_prompt_from_location(config.model_config.general_analyze_prompt_location)
                    print(f"\nUsing custom system prompt from: {config.model_config.general_analyze_prompt_location}")
                    return prompt
                except Exception as e:
                    print(f"\nWarning: Could not load custom system prompt: {str(e)}")
                    print("Falling back to default system prompt.")
                    return cls.get_default_system_prompt()
            return cls.get_default_system_prompt()
        else:
            raise ValueError(f"Invalid prompt type: {prompt_type}. Must be either 'general' or 'deep'") 