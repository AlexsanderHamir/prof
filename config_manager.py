import logging
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
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


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
        logger.info("Starting configuration validation process...")
        logger.info(f"Reading configuration from: {config_path}")

        config_data = load_config_from_file(config_path)
        logger.info("✓ Successfully read configuration file")

        logger.info("Validating configuration structure...")
        validate_config_structure(config_data)
        logger.info("✓ All required top-level fields are present")

        if "benchmark_configs" not in config_data:
            config_data["benchmark_configs"] = {}
            logger.info("No benchmark configurations provided - will analyze all functions")

        logger.info("Validating model configuration...")
        validate_model_config(config_data["model_config"])
        logger.info("✓ All required model configuration fields are present")

        logger.info("Validating benchmark configurations...")
        validate_benchmark_configs(config_data["benchmark_configs"])
        logger.info("✓ All benchmark configurations are valid")

        cls._config_path = config_path
        logger.info("Configuration validation completed successfully! 🎉")

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
