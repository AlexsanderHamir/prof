from pathlib import Path
from typing import Optional
from openai import OpenAI

from utils_config_manager import (
    Config,
    validate_config_structure,
    validate_model_config,
    validate_benchmark_configs,
    create_config_template,
    save_template_to_file,
    load_config_from_file,
    create_config_from_data,
    print_template_creation_info,
    print_validation_progress,
)


class ConfigurationError(Exception):
    pass


class ConfigurationNotFound(ConfigurationError):

    def __init__(self):
        super().__init__("Configuration file not found. Please ensure the configuration file exists and the name is correct. (config_template.json)")


class ConfigurationSetupFailed(ConfigurationError):

    def __init__(self, message: str = "Configuration setup failed, please check the configuration file and try again."):
        super().__init__(message)


class ConfigManager:
    _config_path: Optional[str] = None
    is_flagging: bool = False

    @staticmethod
    def create_template(output_path: Optional[str] = None) -> None:
        template = create_config_template()
        template_path = Path(output_path) if output_path else Path.cwd() / "config_template.json"

        save_template_to_file(template, template_path)
        print_template_creation_info(template_path)

    @classmethod
    def setup_from_file(cls, config_path: str) -> None:
        print_validation_progress("Starting configuration validation process...")
        print_validation_progress("Reading configuration from: {}", config_path)

        config_data = load_config_from_file(config_path)
        print("✓ Successfully read configuration file")

        print_validation_progress("Validating configuration structure...")

        validate_config_structure(config_data)
        print("✓ All required top-level fields are present")

        if "benchmark_configs" not in config_data:
            config_data["benchmark_configs"] = {}
            print_validation_progress("No benchmark configurations provided - will analyze all functions")

        print_validation_progress("Validating model configuration...")

        validate_model_config(config_data["model_config"])
        print("✓ All required model configuration fields are present")

        print_validation_progress("Validating benchmark configurations...")

        validate_benchmark_configs(config_data["benchmark_configs"])
        print("✓ All benchmark configurations are valid")

        cls._config_path = config_path
        print_validation_progress("Configuration validation completed successfully! 🎉")

    @classmethod
    def load(cls) -> Config:
        if not cls._config_path:
            raise ConfigurationNotFound()
        try:
            config_data = load_config_from_file(cls._config_path)
            return create_config_from_data(config_data)
        except Exception as e:
            raise ConfigurationSetupFailed(f"Error loading configuration: {str(e)}")

    @classmethod
    def get_client(cls) -> OpenAI:
        config = cls.load()
        return OpenAI(api_key=config.api_key, base_url=config.base_url)

    @classmethod
    def get_api_key(cls) -> str:
        config = cls.load()
        return config.api_key
