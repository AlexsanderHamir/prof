import logging
from pathlib import Path
import sys
from typing import Optional
from openai import OpenAI

from exit_codes import EXIT_CODE_UNEXPECTED_ERROR, MISSING_CONFIG_FILE
from CONFIG.helpers import (
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

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class ConfigurationParsingError(Exception):

    def __init__(self, message: str = "Error with configuration"):
        super().__init__(message)


class ConfigurationNotFound(ConfigurationParsingError):

    def __init__(self):
        super().__init__("Configuration file not found. Please ensure the configuration file exists and the name is correct. (config_template.json)")


class ConfigurationSetupFailed(ConfigurationParsingError):

    def __init__(self, message: str = "Configuration setup failed, please check the configuration file and try again."):
        super().__init__(message)


class ConfigurationObjectNotSet(ConfigurationParsingError):

    def __init__(self):
        super().__init__("Configuration object not set. Please ensure the configuration file exists and the name is correct. (config_template.json)")


class ConfigManager:
    _config: Optional[Config] = None
    _config_path: Optional[Path] = None
    is_flagging: bool = False

    @staticmethod
    def create_template(output_path: Optional[str] = None) -> None:
        template = create_config_template()
        template_path = Path(output_path) if output_path else Path.cwd() / "config_template.json"
        template_path.parent.mkdir(parents=True, exist_ok=True)

        save_template_to_file(template, template_path)
        ConfigManager._config_path = template_path

        print_template_creation_info(template_path)

    @classmethod
    def setup_from_file(cls, config_path: Path) -> None:
        print_validation_progress("Starting configuration validation process...")
        print_validation_progress("Reading configuration from: {}", config_path)
        try:
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
            cls._config = create_config_from_data(config_data)
            print_validation_progress("Configuration validation completed successfully! 🎉")
        except Exception as e:
            print(f"Unexpected error during configuration setup: {e}", file=sys.stderr)
            sys.exit(EXIT_CODE_UNEXPECTED_ERROR)

    @classmethod
    def get_config(cls) -> Config:
        if not cls._config:
            print("Configuration file not found. Please ensure the configuration file exists and the name is correct. (config_template.json)", file=sys.stderr)
            sys.exit(MISSING_CONFIG_FILE)

        return cls._config

    @classmethod
    def get_client(cls) -> OpenAI:
        config = cls.get_config()
        return OpenAI(api_key=config.api_key, base_url=config.base_url)
